# NameGo

**NameGo** is a Go package designed to seamlessly integrate Golang projects with Python Nameko services. It provides features for setting up RPC clients and servers, managing HTTP APIs, and working with RabbitMQ for messaging. This package simplifies the communication between Go and Nameko services, making it easy to build and maintain distributed systems.

## Installation

To install the package, use:

```bash
go get github.com/joejoe-am/namego
```

## Features

### 1. RPC Client and Server

**Feature:** Enable RPC communication between Go and Nameko services.

#### Example: Setting up an RPC Client
Use `SetupRPCClients` to establish connections with Nameko services via RabbitMQ:

```go
amqpConnection := InitRabbitMQ("amqp://guest:guest@localhost:5672/")

authRpc, quotaRpc, err := SetupRPCClients(amqpConnection)
if err != nil {
    log.Fatalf("failed to initialize RPC clients: %v", err)
}

response, err := authRpc.CallRpc("health_check", map[string]string{})
fmt.Println(response, err)
```

#### Example: Setting up an RPC Server
Use `rpc.NewServer` to create an RPC server and register methods for remote calls:

```go
rpcServer := rpc.NewServer("nameko", amqpConnection)
rpcServer.RegisterMethod("multiply", service.Multiply)

ctx := context.Background()
go func() {
    if err := rpcServer.Start(ctx); err != nil {
        log.Printf("RPC server error: %v", err)
    }
}()
```

### 2. HTTP Server

**Feature:** Serve HTTP endpoints to expose APIs or test integrations.

#### Example: Setting up an HTTP Server
Use `web.New` to create an HTTP server and define routes:

```go
server := web.New()
server.Get("/health", gateway.HealthHandler)
server.Get("/auth-health", gateway.AuthHealthHandler(authRpc), gateway.LoggingMiddleware)

go func() {
    fmt.Println("Starting HTTP server on :8080")
    if err := server.Listen(":8080"); err != nil {
        log.Printf("Web server error: %v", err)
    }
}()
```

### 3. Event Handling

**Feature:** Dispatch and handle events through RabbitMQ.

#### Example: Dispatching Events
Dispatch events using a custom function and RabbitMQ connection:

```go
service.DispatchEventExampleFunction(amqpConnection, "my-service")
```

#### Example: Handling Events (Commented Out in Template)
Use `events.NewEventHandler` to configure and start event handlers:

```go
handlerConfig := events.EventConfig{
    SourceService:    "auth",
    EventType:        "USER_CREATED",
    HandlerType:      events.ServicePool,
    ReliableDelivery: true,
    HandlerFunction:  service.EventHandlerFunction,
}

eventHandler, err := events.NewEventHandler(handlerConfig)
if err != nil {
    log.Fatalf("failed to create event handler: %v", err)
}

if err := eventHandler.Start(amqpConnection); err != nil {
    log.Fatalf("failed to start event handler: %v", err)
}
```

### 4. Middleware Support

**Feature:** Add middleware for HTTP routes easily.

#### Example: Adding Middleware
Pass middleware functions while defining routes:

```go
server.Get("/auth-health", gateway.AuthHealthHandler(authRpc), gateway.LoggingMiddleware)
```

Here, `gateway.LoggingMiddleware` is a custom middleware that can log requests or add other pre-processing functionality.

## Configuration

The package relies on configuration provided by the `configs` module. Example configurations include:

- `RabbitMQURL`: URL for RabbitMQ connection.
- `ServiceName`: The name of the service dispatching events.

## Contributing

Contributions are welcome! Feel free to open an issue or submit a pull request on GitHub.

### TODO

- Refactor package structure and naming conventions.
- Add more examples for event handling.
- Enhance documentation.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Acknowledgments

- [Nameko](https://github.com/nameko/nameko) - Python framework for microservices.
- Inspiration for seamless integration between Go and Nameko projects.

