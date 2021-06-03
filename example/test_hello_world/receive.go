package test_hello_world

import (
	"github.com/streadway/amqp"
	"log"
)

func Receive(ch *amqp.Channel) {
	que, err := ch.QueueDeclare(
		"durable_queue", //name
		true,  //durable
		false,  //delete when unused
		false,  //exclusive
		false,  //no wait
		nil, //arguments)
	)
	if err != nil{
		panic(err)
	}
	// 指定监听的队列
	msgs, err := ch.Consume(
		que.Name,     // queue
		"",         // consumer
		false,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // arguments
		)
	if err != nil{
		panic(err)
	}
	go func() {
		for msg := range msgs{
			log.Printf("Received a message : %s", msg.Body)
			// 确认收到消息，参照TCP
			err = msg.Ack(false)
		}
	}()
	select {}
}
