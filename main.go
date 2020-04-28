package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mariiatuzovska/rabbit-mq/api"
	"github.com/mariiatuzovska/rabbit-mq/api/forms"
	pc "github.com/mariiatuzovska/rabbit-mq/producer-consumer"
	ps "github.com/mariiatuzovska/rabbit-mq/publish-subscribe"
	"github.com/urfave/cli"
)

var (
	ServiceName = "rabbit-mq"
	Version     = "0.0.1"
)

func main() {
	app := cli.NewApp()
	app.Name = ServiceName
	app.Usage = "Example of message publisher & message consumer for RabbitMQ"
	app.Description = ""
	app.Version = Version
	app.Copyright = "2020, mariiatuzovska"
	app.Authors = []cli.Author{cli.Author{Name: "Tuzovska Mariia"}}
	app.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: "starting as api service",
			Action: func(c *cli.Context) error {
				srv := api.New()
				go ps.Daemon("PUBLISHER", "SUBSCRIBER")
				go pc.Daemon("PRODUCER", "COMSUMER")
				return srv.Start(fmt.Sprintf("127.0.0.1:%s", c.String("p")))
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "p",
					Usage: "port",
					Value: "8080",
				},
			},
		},
		{
			Name:  "bad-ack-example",
			Usage: "PRODUCER-CONSUMER example with bad ack",
			Action: func(c *cli.Context) error {

				producer := pc.New("P")
				badConsumer := pc.New("FIRST")
				goodConsumer := pc.New("SECOND")

				// background process
				badConsumerMSGS, err := badConsumer.Receive()
				if err != nil {
					log.Fatal(err)
				}
				goodConsumerMSGS, err := goodConsumer.Receive()
				if err != nil {
					log.Fatal(err)
				}
				forever := make(chan bool)
				go func() {
					for d := range badConsumerMSGS {
						req := new(forms.Message)
						json.Unmarshal(d.Body, req)
						log.Printf("Received a message from FIRST cunsumer: %s", req.Text)
					}
				}()
				go func() {
					for d := range goodConsumerMSGS {
						req := new(forms.Message)
						json.Unmarshal(d.Body, req)
						log.Printf("Received a message from SECOND cunsumer: %s", req.Text)
					}
				}()

				log.Printf(" [*] Waiting for messages. To exit press OPTION+C")

				// try to make bad ack
				go pc.DaemonWithException("P", "FIRST")
				go pc.Daemon("P", "SECOND")

				time.Sleep(time.Duration(5) * time.Second)

				err = producer.Send(forms.NewMessage("1-th message"))
				log.Println(1, err)
				err = producer.Send(forms.NewMessage("2-th message"))
				log.Println(2, err)
				err = producer.Send(forms.NewMessage("3-th message"))
				log.Println(3, err)
				err = producer.Send(forms.NewMessage("4-th message"))
				log.Println(4, err)
				err = producer.Send(forms.NewMessage("5-th message"))
				log.Println(5, err)

				time.Sleep(time.Duration(10) * time.Second)

				err = producer.Send(forms.NewMessage("6-th message"))
				log.Println(6, err)
				err = producer.Send(forms.NewMessage("7-th message"))
				log.Println(7, err)
				err = producer.Send(forms.NewMessage("8-th message"))
				log.Println(8, err)

				<-forever

				return nil
			},
		},
		{
			Name:  "pc-send",
			Usage: "PRODUCER-CONSUMER send (point-to-point)", // user application that sends messages
			Action: func(c *cli.Context) error { // producer-consumer.go
				rabbit := pc.New("myqueue")
				rabbit.Send(forms.NewMessage(c.String("m")))
				return nil
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "m",
					Usage: "message",
					Value: "Hello World!",
				},
			},
		},
		{
			Name:  "pc-receive",
			Usage: "PRODUCER-CONSUMER receive (point-to-point)", // user application that receives messages
			Action: func(c *cli.Context) error { // producer-consumer.go
				rabbit := pc.New("myqueue")
				msgs, err := rabbit.Receive()
				if err != nil {
					log.Fatal(err)
				}
				forever := make(chan bool)
				go func() {
					for d := range msgs {
						req := new(forms.Message)
						json.Unmarshal(d.Body, req)
						log.Printf("Received a message: %s", req.Text)
					}
				}()
				log.Printf(" [*] Waiting for messages. To exit press OPTION+C")
				<-forever
				return nil
			},
		},
		{
			Name:  "ps-send",
			Usage: "PUBLISH-SUBSCRIBE publish (topic)", // user application that sends messages
			Action: func(c *cli.Context) error { // publish-subscribe.go
				rabbit := ps.New("mytopic")
				rabbit.Send(forms.NewTopic(c.String("m")))
				return nil
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "m",
					Usage: "message",
					Value: "Hello World!",
				},
			},
		},
		{
			Name:  "ps-receive",
			Usage: "PUBLISH-SUBSCRIBE receive (topic)", // user application that receives messages
			Action: func(c *cli.Context) error { // publish-subscribe.go
				rabbit := ps.New("mytopic")
				msgs, err := rabbit.Receive()
				if err != nil {
					log.Fatal(err)
				}
				forever := make(chan bool)
				go func() {
					for d := range msgs {
						req := new(forms.Topic)
						json.Unmarshal(d.Body, req)
						log.Printf(" [x] %s", req.Text)
					}
				}()
				log.Printf(" [*] Waiting for logs. To exit press OPTION+C")
				<-forever
				return nil
			},
		},
	}

	app.Run(os.Args)
}
