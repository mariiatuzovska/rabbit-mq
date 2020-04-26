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
	Consumer      string
}

func New(name string) *Rabbit {
	return &Rabbit{
		ExchangeName:  name,
		ServerAddress: "amqp://guest:guest@localhost:5672/",
		Consumer:      "",
	}
}

func (rabbit *Rabbit) Send(request *forms.Topic) error {
	conn, err := amqp.Dial(rabbit.ServerAddress)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	err = ch.ExchangeDeclare(
		rabbit.ExchangeName, // name
		"topic",             // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return err
	}
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	err = ch.Publish(
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
	log.Printf(" [x] Sent %s", body)
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
	err = ch.ExchangeDeclare(
		rabbit.ExchangeName, // name
		"topic",             // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return nil, err
	}
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(
		q.Name,              // queue name
		"",                  // routing key
		rabbit.ExchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return ch.Consume(
		q.Name,          // queue
		rabbit.Consumer, // consumer
		true,            // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
}

func Daemon() {

	forever := make(chan bool)

	requests, responses := New("PUBLISHER"), New("SUBSCRIBER")
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
			req.Nonce = fmt.Sprintf("PUBLISH-SUBSCRIBE Daemon like it #%d", counter)
			responses.Send(req)
		}
	}()
	log.Printf(" [*] PUBLISH-SUBSCRIBE Daemon waiting for messages...")

	for {
		<-forever
	}

}
