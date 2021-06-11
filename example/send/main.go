package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"order_center/example/test_hello_world"
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
	// 公平分发
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		panic(err)
	}
	test_hello_world.Send(ch)
}
