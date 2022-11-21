package msg_broker

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"golek_posts_service/pkg/contracts"
	"log"
	"time"
)

type MQPublisher struct {
	amqpConnection *amqp.Connection
	amqpChannel    *amqp.Channel
}

func (m *MQPublisher) Publish(title string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.amqpChannel.PublishWithContext(ctx,
		"posts_exchange", // exchange
		"new_post_route", // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(title),
		})
	if err != nil {
		return err
	}

	log.Printf("[x] Sent Notification %s", title)

	return nil

}

func NewMQPublisher(amqpConn *amqp.Connection) contracts.MessageQueue {
	return &MQPublisher{amqpConnection: amqpConn}
}

func (m *MQPublisher) Setup() {

	//open channel
	channel, err := m.amqpConnection.Channel()
	failOnError(err, "Failed to open a channel")
	//defer channel.Close()

	//declare an exchange
	err = channel.ExchangeDeclare(
		"posts_exchange", // name
		"topic",          // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	m.amqpChannel = channel
}
