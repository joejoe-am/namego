package events

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Dispatch sends an event with the given type and payload.
func Dispatch(conn *amqp.Connection, sourceService string, eventType string, payload []byte) error {
	exchangeName := fmt.Sprintf("%s.events", sourceService)

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open RabbitMQ channel: %v", err)
	}

	err = ch.ExchangeDeclare(
		exchangeName,
		"topic",
		true,  // durable
		true,  // autoDelete
		false, // internal
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %v", err)
	}

	err = ch.Publish(
		exchangeName, // exchange
		eventType,    // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %v", err)
	}

	fmt.Printf("Event dispatched: %s to %s.exchange\n", eventType, sourceService)
	return nil
}
