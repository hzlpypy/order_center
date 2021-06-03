package main

import (
	"fmt"
	"github.com/streadway/amqp"
)

func main() {
	conn, err := amqp.Dial("amqp://admin:admin@localhost:5672")
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to RabbitMQ, err=%v", err))
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		panic(fmt.Sprintf("make conn chan error, err=%v", err))
	}
	defer ch.Close()
	_ = ch.ExchangeDeclare(
		"callback",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
}
