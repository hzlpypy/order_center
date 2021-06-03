package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"os"
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
	if len(os.Args) < 2 {
		panic("os.Args len < 2")
	}
	for _, routingKey := range os.Args[1:] {
		log.Printf("Binding queue %s to exchange %s with routing key %s",
			q.Name, "logs_direct", routingKey)
		err = ch.QueueBind(
			q.Name,        // queue name
			routingKey,    // routing key
			"test_direct", // exchange
			false,
			nil)
		failOnError(err, "Failed to bind a queue")
	}
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
