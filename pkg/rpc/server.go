package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/joejoe-am/namego/configs"
	amqp "github.com/rabbitmq/amqp091-go"
	"strings"
)

type Server struct {
	serviceName    string
	amqpConnection *amqp.Connection
	amqpChannel    *amqp.Channel
	methods        map[string]func(args interface{}, kwargs map[string]interface{}) (interface{}, error)
}

// NewRpcServer initializes and returns a Server instance.
func NewRpcServer(serviceName string) *Server {
	conn, err := amqp.Dial(configs.RabbitMQURL)
	if err != nil {
		panic(fmt.Errorf("failed to connect to RabbitMQ: %v", err))
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(fmt.Errorf("failed to open a channel: %v", err))
	}

	return &Server{
		serviceName:    serviceName,
		amqpConnection: conn,
		amqpChannel:    ch,
		methods:        make(map[string]func(args interface{}, kwargs map[string]interface{}) (interface{}, error)),
	}
}

// RegisterMethod registers an RPC method with the server.
func (r *Server) RegisterMethod(methodName string, handler func(args interface{}, kwargs map[string]interface{}) (interface{}, error)) {
	r.methods[methodName] = handler
}

// Start begins listening for RPC requests on the service's queue.
func (r *Server) Start() {
	queueName := fmt.Sprintf(RpcQueueTemplate, r.serviceName)
	routingKey := fmt.Sprintf("%s.*", r.serviceName)

	_, err := r.amqpChannel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(fmt.Errorf("failed to declare RPC queue: %v", err))
	}

	err = r.amqpChannel.QueueBind(
		queueName,
		routingKey,
		configs.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		panic(fmt.Errorf("failed to bind RPC queue: %v", err))
	}

	msgs, err := r.amqpChannel.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(fmt.Errorf("failed to consume messages: %v", err))
	}

	for msg := range msgs {
		go r.handleRequest(msg)
	}
}

// handleRequest processes incoming messages and invokes the appropriate handler.
func (r *Server) handleRequest(msg amqp.Delivery) {
	routingKeyParts := strings.Split(msg.RoutingKey, ".")
	if len(routingKeyParts) != 2 {
		r.sendResponse(
			msg.ReplyTo,
			msg.CorrelationId,
			nil, fmt.Errorf("invalid routing key: %s", msg.RoutingKey),
		)
		return
	}
	methodName := routingKeyParts[1] // The second part is the method name

	var request struct {
		Args   interface{}            `json:"args"`
		Kwargs map[string]interface{} `json:"kwargs"`
	}

	if err := json.Unmarshal(msg.Body, &request); err != nil {
		r.sendResponse(msg.ReplyTo, msg.CorrelationId, nil, fmt.Errorf("invalid request: %v", err))
		return
	}

	handler, exists := r.methods[methodName]

	if !exists {
		r.sendResponse(msg.ReplyTo, msg.CorrelationId, nil, fmt.Errorf("method not found: %s", methodName))
		return
	}

	// TODO: handle the kwargs

	result, err := handler(request.Args, request.Kwargs)
	r.sendResponse(msg.ReplyTo, msg.CorrelationId, result, err)
}

// sendResponse sends a response back to the caller.
func (r *Server) sendResponse(replyTo, correlationID string, result interface{}, err error) {
	var response Response

	if err != nil {
		response.Error = &struct {
			ExcType   string                 `json:"exc_type"`
			ExcPath   string                 `json:"exc_path"`
			ExcArgs   []interface{}          `json:"exc_args"`
			ExcKwargs map[string]interface{} `json:"exc_kwargs"`
			Value     string                 `json:"value"`
		}{
			ExcType:   "",
			ExcPath:   "",
			ExcArgs:   []interface{}{err.Error()},
			ExcKwargs: map[string]interface{}{},
			Value:     err.Error(),
		}
	} else {
		response.Result = result
	}

	body, err := json.Marshal(response)

	if err != nil {
		fmt.Printf("failed to serialize response: %v\n", err)
		return
	}

	err = r.amqpChannel.Publish(
		configs.ExchangeName,
		replyTo,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			Body:          body,
		},
	)
	if err != nil {
		fmt.Printf("failed to send response: %v\n", err)
	}
}

// Close gracefully shuts down the server.
func (r *Server) Close() {
	if r.amqpChannel != nil {
		_ = r.amqpChannel.Close()
	}
	if r.amqpConnection != nil {
		_ = r.amqpConnection.Close()
	}
}
