package main

import (
	"fmt"
	"github.com/joejoe-am/namego/configs"
	"github.com/joejoe-am/namego/examples/example-service/gateway"
	"github.com/joejoe-am/namego/examples/example-service/service"
	"github.com/joejoe-am/namego/pkg/rpc"
	"github.com/joejoe-am/namego/pkg/web"
	"log"
)

// TODO: change package name

func main() {
	amqpConnection := InitRabbitMQ(configs.RabbitMQURL)
	defer amqpConnection.Close()

	authRpc, quotaRpc, err := SetupRPCClients(amqpConnection)

	if err != nil {
		log.Fatalf("failed to initialize RPC clients: %v", err)
	}

	response, err := authRpc.CallRpc("health_check", map[string]string{})
	fmt.Println(response, err)

	response, err = quotaRpc.CallRpc("health_check", map[string]string{})
	fmt.Println(response, err)

	server := web.New()
	server.Get("/health", gateway.HealthHandler)
	server.Get("/auth-health", gateway.AuthHealthHandler(authRpc), LoggingMiddleware)

	rpcServer := rpc.NewServer("nameko", amqpConnection)
	rpcServer.RegisterMethod("multiply", service.Multiply)

	handlerConfig := rpc.EventConfig{
		SourceService:    "authnzng",
		EventType:        "EVENT_EXAMPLE",
		HandlerType:      rpc.ServicePool,
		ReliableDelivery: true,
		HandlerFunction:  service.EventHandlerFunction,
	}

	eventHandler, err := rpc.NewEventHandler(handlerConfig, amqpConnection)
	if err != nil {
		log.Fatalf("failed to create event handler: %v", err)
	}

	err = eventHandler.SetupQueue()
	if err != nil {
		log.Fatalf("failed to setup queue: %v", err)
	}

	err = eventHandler.Start()
	if err != nil {
		log.Fatalf("failed to start event handler: %v", err)
	}

	app := NewApp(rpcServer, server)
	app.Run()
}
