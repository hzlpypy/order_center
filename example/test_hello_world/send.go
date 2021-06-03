package test_hello_world

import (
	"github.com/streadway/amqp"
	"log"
)

func Send(ch *amqp.Channel) {
	que, err := ch.QueueDeclare("durable_queue", //name
		true,  //durable 持久化
		false, //delete when unused
		false, //exclusive
		false, //no wait
		nil,   //arguments)
	)
	if err != nil {
		panic(err)
	}
	body := []byte("hello world")
	err = ch.Publish("", que.Name, false, false, amqp.Publishing{
		ContentType:  "text/plain",
		Body:         body,
		DeliveryMode: amqp.Persistent, // 发送模式 消息持久化
	})
	if err != nil {
		panic(err)
	}
	log.Println("send a message: hello world")

}
