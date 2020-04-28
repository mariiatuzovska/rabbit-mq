package pc

// POINT-TO-POINT

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/mariiatuzovska/rabbit-mq/api/forms"
	"github.com/streadway/amqp"
)

type Rabbit struct {
	QueueName     string
	ServerAddress string

	channel *amqp.Channel
	queue   amqp.Queue
}

func New(name string) *Rabbit {
	rabbit := &Rabbit{
		QueueName: name,
		// ServerAddress: "amqp://guest:guest@localhost:5672/",
		ServerAddress: "amqp://vwaqpldk:9Z97UAS6P7dGFY84oHxTDryrijrMHl2S@roedeer.rmq.cloudamqp.com/vwaqpldk",
	}
	conn, err := amqp.Dial(rabbit.ServerAddress)
	if err != nil {
		log.Fatal(err)
	}
	rabbit.channel, err = conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	rabbit.queue, err = rabbit.channel.QueueDeclare(
		rabbit.QueueName, // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		true,             // no-wait
		nil,
	// amqp.Table{ // arguments
	// 	"x-message-ttl": int32(60000), // Declares a queue with the x-message-ttl extension
	// 	// to exercise integer serialization. 60 sec.)
	// 	"x-max-length": 10, // Maximum number of messages can be set
	// 	// by supplying the `x-max-length` queue declaration argument with a non-negative integer value.
	// },
	)

	rabbit.queue.Messages = 2 // count of messages not awaiting acknowledgment (#1)
	if err != nil {
		log.Fatal(err)
	}
	return rabbit
}

func (rabbit *Rabbit) Send(request *forms.Message) error {

	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	err = rabbit.channel.Publish(
		"",                // exchange
		rabbit.queue.Name, // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
		})
	if err != nil {
		return err
	}
	return nil
}

func (rabbit *Rabbit) Receive() (<-chan amqp.Delivery, error) {
	return rabbit.channel.Consume(
		rabbit.queue.Name, // queue
		"",                // consumer
		true,              // auto-ack
		false,             // exclusive
		false,             // no-local
		true,              // no-wait
		nil,               // args
	)
}

func Daemon(reqestor, responser string) {

	forever := make(chan bool)

	requests, responses := New(reqestor), New(responser)
	msgs, err := requests.Receive()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		var counter int = 0
		// time.Sleep(time.Duration(15) * time.Second)
		for d := range msgs {
			counter++
			req := new(forms.Message)
			json.Unmarshal(d.Body, req)
			req.Nonce = fmt.Sprintf("%s-%s Daemon #%d", reqestor, responser, counter)
			responses.Send(req)
		}
	}()
	log.Printf(fmt.Sprintf(" [*] %s-%s Daemon waiting for messages...", reqestor, responser))

	for {
		<-forever
	}

}

func DaemonWithException(reqestor, responser string) {

	exit := make(chan bool)

	requests, responses := New(reqestor), New(responser)
	msgs, err := requests.Receive()
	if err != nil {
		log.Fatal(err)
	}

	go func(a, b string, exit chan bool) {
		var counter int = 0
		for d := range msgs {
			counter++
			exit <- false
			req := new(forms.Message)
			json.Unmarshal(d.Body, req)
			if counter == 3 {
				exit <- true
			}
			req.Nonce = fmt.Sprintf("%s-%s Daemon #%d", a, b, counter)
			responses.Send(req)
		}
	}(reqestor, responser, exit)

	log.Printf(fmt.Sprintf(" [*] %s-%s Daemon waiting for messages...", reqestor, responser))

	for {
		end := <-exit
		if end {
			break
		}
	}

}
