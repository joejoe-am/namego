package main

import (
	"fmt"
	"github.com/joejoe-am/namego/pkg/rpc"
	"github.com/joejoe-am/namego/pkg/web"
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

type App struct {
	RPCServer *rpc.Server
	WebServer *web.Http
}

func NewApp(rpcServer *rpc.Server, webServer *web.Http) *App {
	return &App{RPCServer: rpcServer, WebServer: webServer}
}

func (a *App) Run() {
	fmt.Println("Web server running on :8080")
	go log.Fatal(a.WebServer.Listen(":8080"))

	log.Println("Starting RPC server")
	if err := a.RPCServer.Start(); err != nil {
		log.Fatalf("RPC server error: %v", err)
	}
}
