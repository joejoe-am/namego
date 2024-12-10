package main

import (
	"fmt"
	"github.com/joejoe-am/namego/configs"
	"github.com/joejoe-am/namego/pkg/rpc"
	"github.com/joejoe-am/namego/pkg/web"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/valyala/fasthttp"
	"log"
)

// TODO: change package name

func main() {
	amqpConnection, err := amqp.Dial(configs.RabbitMQURL)

	if err != nil {
		log.Fatalf("failed to establish RabbitMQ connection: %v", err)
	}

	err = rpc.InitClient(amqpConnection)

	if err != nil {
		log.Fatal(err)
	}

	authRpc := rpc.NewClient("authnzng")
	quotaRpc := rpc.NewClient("quota")

	response, err := authRpc.CallRpc("health_check", map[string]string{})
	fmt.Println(response, err)

	response, err = quotaRpc.CallRpc("health_check", map[string]string{})
	fmt.Println(response, err)

	server := web.New()

	server.Get("/health", func(ctx *fasthttp.RequestCtx) { ctx.WriteString("OK") })
	server.Get("/auth-health", AuthHealthHandler(authRpc), LoggingMiddleware)

	rpcServer := rpc.NewServer("nameko", amqpConnection)

	rpcServer.RegisterMethod("multiply", Multiply)

	go func() {
		fmt.Println("Server running on :8080")
		if err := server.Listen(":8080"); err != nil {
			log.Fatalf("Error starting server: %v\n", err)
		}
	}()

	err = rpcServer.Start()

	if err != nil {
		log.Fatal(err)
	}

}
