package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/mariiatuzovska/rabbit-mq/api/forms"
	pc "github.com/mariiatuzovska/rabbit-mq/producer-consumer"
	ps "github.com/mariiatuzovska/rabbit-mq/publish-subscribe"
	"github.com/streadway/amqp"
)

type (
	Service struct {
		*echo.Echo
		producer   *pc.Rabbit
		consumer   <-chan amqp.Delivery
		publisher  *ps.Rabbit
		subscriber <-chan amqp.Delivery
	}
)

func New() *Service {

	consumer := pc.New("COMSUMER")
	msgs, err := consumer.Receive()
	if err != nil {
		log.Fatal(err)
	}
	subscriber := ps.New("SUBSCRIBER")
	topics, err := subscriber.Receive()
	if err != nil {
		log.Fatal(err)
	}
	srv := &Service{echo.New(), pc.New("PRODUCER"), msgs, ps.New("PUBLISHER"), topics}
	srv.HideBanner = true
	srv.HidePort = true

	srv.POST("/message", srv.produce)
	srv.GET("/message", srv.consume)

	srv.POST("/topic", srv.publish)
	srv.GET("/topic", srv.subscribe)

	return srv
}

func (srv *Service) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	log.Println("SHUT DOWN")
	return srv.Shutdown(ctx)
}

func (srv *Service) produce(c echo.Context) error {
	var req forms.Message
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err := srv.producer.Send(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, req)
}

func (srv *Service) consume(c echo.Context) error {
	body := <-srv.consumer
	resp := new(forms.Message)
	err := json.Unmarshal(body.Body, resp)
	if err != nil {
		return c.JSON(http.StatusForbidden, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (srv *Service) publish(c echo.Context) error {
	var req forms.Topic
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err := srv.publisher.Send(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, req)
}

func (srv *Service) subscribe(c echo.Context) error {
	body := <-srv.subscriber
	resp := new(forms.Topic)
	err := json.Unmarshal(body.Body, resp)
	if err != nil {
		return c.JSON(http.StatusForbidden, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}
