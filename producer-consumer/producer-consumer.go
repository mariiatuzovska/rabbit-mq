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
	Consumer      string
}

func New(name string) *Rabbit {
	return &Rabbit{
		QueueName:     name,
		ServerAddress: "amqp://guest:guest@localhost:5672/",
		Consumer:      "",
	}
}

func (rabbit *Rabbit) Send(request *forms.Message) error {
	conn, err := amqp.Dial(rabbit.ServerAddress)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	q, err := ch.QueueDeclare(
		rabbit.QueueName, // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return err
	}
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
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
	conn, err := amqp.Dial(rabbit.ServerAddress)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	q, err := ch.QueueDeclare(
		rabbit.QueueName, // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return nil, err
	}
	return ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
}

func Daemon() {

	forever := make(chan bool)

	requests, responses := New("PRODUCER"), New("COMSUMER")
	msgs, err := requests.Receive()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		var counter int = 0
		for d := range msgs {
			counter++
			req := new(forms.Message)
			json.Unmarshal(d.Body, req)
			req.Nonce = fmt.Sprintf("PRODUCER-COMSUMER Daemon like it #%d", counter)
			responses.Send(req)
		}
	}()
	log.Printf(" [*] PRODUCER-COMSUMER Daemon waiting for messages...")

	for {
		<-forever
	}

}
