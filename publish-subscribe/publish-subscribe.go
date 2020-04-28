package ps

// TOPIC

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/mariiatuzovska/rabbit-mq/api/forms"
	"github.com/streadway/amqp"
)

type Rabbit struct {
	ExchangeName  string
	ServerAddress string

	channel *amqp.Channel
	queue   amqp.Queue
}

func New(name string) *Rabbit {
	rabbit := &Rabbit{
		ExchangeName: name,
		// ServerAddress: "amqp://guest:guest@localhost:5672/",
		ServerAddress: "amqp://vwaqpldk:9Z97UAS6P7dGFY84oHxTDryrijrMHl2S@roedeer.rmq.cloudamqp.com/vwaqpldk",
	}
	conn, err := amqp.Dial(rabbit.ServerAddress)
	if err != nil {
		log.Fatal(err)
	}
	err = conn.Config.Properties.Validate()
	if err != nil {
		log.Fatal(err)
	}
	rabbit.channel, err = conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	err = rabbit.channel.ExchangeDeclare(
		rabbit.ExchangeName, // name
		"topic",             // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		log.Fatal(err)
	}
	rabbit.queue, err = rabbit.channel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatal(err)
	}
	err = rabbit.channel.QueueBind(
		rabbit.queue.Name,   // queue name
		"",                  // routing key
		rabbit.ExchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	return rabbit
}

func (rabbit *Rabbit) Send(request *forms.Topic) error {
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	err = rabbit.channel.Publish(
		rabbit.ExchangeName, // exchange
		"",                  // routing key
		false,               // mandatory
		false,               // immediate
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
		false,             // no-wait
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
		for d := range msgs {
			counter++
			req := new(forms.Topic)
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
