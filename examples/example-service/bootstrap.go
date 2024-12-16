package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func InitRabbitMQ(RabbitMQURL string) *amqp.Connection {
	conn, err := amqp.Dial(RabbitMQURL)
	if err != nil {
		log.Fatalf("failed to establish RabbitMQ connection: %v", err)
	}
	return conn
}
