package msg_broker

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func New() *amqp.Connection {
	conn, err := amqp.Dial("amqp://guest:guest@172.17.0.2:5672/")
	if err != nil {
		log.Fatalf("Failed to establish RabbitMQ connection: %v", err.Error())
	}
	return conn
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
