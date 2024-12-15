package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"strings"
)

type Server struct {
	serviceName    string
	amqpConnection *amqp.Connection
	amqpChannel    *amqp.Channel
	methods        map[string]func(args interface{}, kwargs map[string]interface{}) (interface{}, error)
}

// NewServer initializes and returns a Server instance.
func NewServer(serviceName string, amqpConnection *amqp.Connection) *Server {
	return &Server{
		serviceName:    serviceName,
		amqpConnection: amqpConnection,
		methods:        make(map[string]func(args interface{}, kwargs map[string]interface{}) (interface{}, error)),
	}
}

// RegisterMethod registers an RPC method with the server.
func (s *Server) RegisterMethod(methodName string, handler func(args interface{}, kwargs map[string]interface{}) (interface{}, error)) {
	s.methods[methodName] = handler
}

// Start begins listening for RPC requests on the service's queue.
func (s *Server) Start(ctx context.Context) error {
	var err error

	queueName := fmt.Sprintf(RpcQueueTemplate, s.serviceName)
	routingKey := fmt.Sprintf("%s.*", s.serviceName)

	s.amqpChannel, err = s.amqpConnection.Channel()

	if err != nil {
		return fmt.Errorf("error creating amqp channel: %v", err)
	}

	_, err = s.amqpChannel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to declare RPC queue: %v", err)
	}

	err = s.amqpChannel.QueueBind(
		queueName,
		routingKey,
		Cfg.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind RPC queue: %v", err)
	}

	err = s.amqpChannel.Qos(
		2,
		0,
		false,
	)

	msgs, err := s.amqpChannel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to consume messages: %v", err)
	}

	workerPool := make(chan struct{}, 2)

	// TODO: handle worker pool better (worker pool pattern)

	for msg := range msgs {
		workerPool <- struct{}{}
		go func(m amqp.Delivery) {
			defer func() { <-workerPool }()
			s.handleRequest(m)
		}(msg)
	}

	return nil
}

// handleRequest processes incoming messages and invokes the appropriate handler.
func (s *Server) handleRequest(msg amqp.Delivery) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("panic occurred: %v\n", r)
			msg.Nack(false, true) // Requeue the message in case of panic
		}
	}()

	err := s.processMessage(msg)
	if err != nil {
		fmt.Printf("error processing message: %v\n", err)
	}

	msg.Ack(false) // Acknowledge the message on success
}

// processMessage decodes and handles the message logic.
func (s *Server) processMessage(msg amqp.Delivery) error {
	// Validate routing key structure
	routingKeyParts := strings.Split(msg.RoutingKey, ".")
	if len(routingKeyParts) != 2 {
		return s.sendResponse(msg, nil, fmt.Errorf("invalid routing key: %s", msg.RoutingKey))
	}

	methodName := routingKeyParts[1] // Extract method name

	// Parse the request body
	var request struct {
		Args   interface{}            `json:"args"`
		Kwargs map[string]interface{} `json:"kwargs"`
	}
	if err := json.Unmarshal(msg.Body, &request); err != nil {
		return s.sendResponse(msg, nil, fmt.Errorf("invalid request: %v", err))
	}

	// Check if the handler exists
	handler, exists := s.methods[methodName]
	if !exists {
		return s.sendResponse(msg, nil, fmt.Errorf("method not found: %s", methodName))
	}

	// Invoke the handler
	result, err := handler(request.Args, request.Kwargs)
	return s.sendResponse(msg, result, err)
}

// sendResponse constructs and sends a success response.
func (s *Server) sendResponse(msg amqp.Delivery, result interface{}, err error) error {
	response := Response{}
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

	body, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		return fmt.Errorf("failed to serialize response: %v", marshalErr)
	}

	publishErr := s.amqpChannel.Publish(
		Cfg.ExchangeName,
		msg.ReplyTo,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: msg.CorrelationId,
			Body:          body,
		},
	)
	if publishErr != nil {
		return fmt.Errorf("failed to send response: %v", publishErr)
	}

	return nil
}
