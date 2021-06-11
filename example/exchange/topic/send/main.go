package main

import (
	"fmt"
	"github.com/hzlpypy/common/utils"
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
	// 公平分发
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "")
	exchangeName := "test_topic"
	exchangeType := "topic" //
	// 创建交换机
	err = ch.ExchangeDeclare(
		exchangeName, // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "create exchange error")
	if err != nil {
		panic(err)
	}
	// 将消息发送到交换机
	msg := "fa fa fa topic"
	err = ch.Publish(
		exchangeName,
		severityFrom(os.Args),
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plan",
			Body:          []byte(msg),
			CorrelationId: utils.NewUUID(),
			//ReplyTo:
		},
	)
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s", msg)
}

func severityFrom(args []string) string {
	var s string
	if len(args) < 2 || args[1] == "" {
		s = "info"
	} else {
		s = os.Args[1]
	}

	return s
}
