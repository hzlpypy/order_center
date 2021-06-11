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
	// 创建队列
	q, err := ch.QueueDeclare(
		"",
		false,
		false,
		false,
		false,
		map[string]interface{}{"x-max-length": 10}, // 设置队列最大长度（参数限制了一个队列的消息总数，当消息总数达到限定值时，队列头的消息会被抛弃）
	)
	failOnError(err, "")
	if len(os.Args) < 2 {
		panic("os.Args len < 2")
	}
	for _, routingKey := range os.Args[1:] {
		log.Printf("Binding queue %s to exchange %s with routing key %s",
			q.Name, "logs_topic", routingKey)
		err = ch.QueueBind(
			q.Name,       // queue name
			routingKey,   // routing key
			"test_topic", // exchange
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
