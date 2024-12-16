package events

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

// Dispatch sends an event with the given type and payload.
func Dispatch(conn *amqp.Connection, sourceService string, eventType string, payload []byte) error {
	exchangeName := sourceService + ".events"

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("failed to open RabbitMQ channel: %v", err)
		return err
	}

	err = ch.ExchangeDeclare(
		exchangeName,
		"topic",
		true,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("failed to declare RabbitMQ exchange: %v", err)
		return err
	}

	err = ch.Publish(
		exchangeName,
		eventType, // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		},
	)
	if err != nil {
		log.Printf("failed to publish event: %v", err)
		return err
	}

	log.Printf("Event dispatched: %s to %s.exchange\n", eventType, sourceService)
	return nil
}
