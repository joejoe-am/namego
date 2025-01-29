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
Use `NewClient` to establish connections with Nameko services via RabbitMQ:

``` go
import (
    "github.com/joejoe-am/namego/pkg/rpc"
)

amqpConnection := InitRabbitMQ("amqp://guest:guest@localhost:5672/")

err := rpc.InitClient(amqpConnection)

svcRpc := rpc.NewClient("other_service")
```

#### Example: Setting up an RPC Server
Use `rpc.NewServer` to create an RPC server and register methods for remote calls:

``` go
import (
    "github.com/joejoe-am/namego/pkg/rpc"
)

rpcServer := rpc.NewServer("my_service_name", amqpConnection)
rpcServer.RegisterMethod("example_method", ExampleMethodFunction)

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

``` go
import (
    "github.com/joejoe-am/namego/pkg/web"
)

server := web.New()
server.Get("/health", gateway.HealthHandler, LoggingMiddleware)

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

``` go
import (
    "encoding/json"
    "fmt"
    "github.com/joejoe-am/namego/pkg/rpc/events"
)

eventData := map[string]interface{}{
    "id":   "12345",
    "name": "example",
}

payload, err := json.Marshal(eventData)

if err != nil {
    fmt.Printf("Failed to marshal event data: %v", err)
}

err = events.Dispatch(amqpConnection, "my_service_name", "TEST_EVENT_HANDLER", payload)

if err != nil {
    fmt.Printf("Failed to dispatch event: %v\n", err)
} else {
    fmt.Println("Event successfully dispatched.")
}
```

## Configuration

The package relies on configuration provided by the `configs` module. Example configurations include:

- `RabbitMQURL`: URL for RabbitMQ connection.
- `ServiceName`: The name of the service dispatching events.

### Config File Path

By default, the package looks for a `config.yaml` file in the root of your project directory. If the file is not found, it searches parent directories until it reaches the system root.

#### Custom Config Path
To override the default path, you can set an environment variable:

```bash
export CONFIG_PATH=/path/to/your/config.yaml
```

This allows flexibility in specifying configuration locations.

## Contributing

Contributions are welcome! Feel free to open an issue or submit a pull request on GitHub.

### TODO

- Refactor package structure and naming conventions.
- Enhance documentation.

## Acknowledgments

- [Nameko](https://github.com/nameko/nameko) - Python framework for microservices.
- Inspiration for seamless integration between Go and Nameko projects.

