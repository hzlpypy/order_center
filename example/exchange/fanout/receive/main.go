package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

func failOnError(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %v", msg, err))
	}
}

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
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	q, err := ch.QueueDeclare(
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "")

	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		"logs", // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")
	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")
	for msg := range msgs {
		log.Printf(" [x] %s", msg.Body)
	}
	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	select {}
}
