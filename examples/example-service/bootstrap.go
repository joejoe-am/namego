package main

import (
	"github.com/joejoe-am/namego/pkg/rpc"
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

func SetupRPCClients(conn *amqp.Connection) (authClient, quotaClient *rpc.Client, err error) {
	err = rpc.InitClient(conn)
	if err != nil {
		return nil, nil, err
	}

	authClient = rpc.NewClient("authnzng")
	quotaClient = rpc.NewClient("quota")
	return authClient, quotaClient, nil
}
