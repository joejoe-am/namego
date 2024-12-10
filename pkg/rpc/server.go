package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/joejoe-am/namego/configs"
	amqp "github.com/rabbitmq/amqp091-go"
	"strings"
)

const (
	RpcQueueTemplate = "rpc-%s"
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
func (s *Server) Start() error {
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
		configs.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind RPC queue: %v", err)
	}

	msgs, err := s.amqpChannel.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %v", err)
	}

	for msg := range msgs {
		go s.handleRequest(msg)
	}

	return nil
}

// handleRequest processes incoming messages and invokes the appropriate handler.
func (s *Server) handleRequest(msg amqp.Delivery) {
	routingKeyParts := strings.Split(msg.RoutingKey, ".")
	if len(routingKeyParts) != 2 {
		s.sendResponse(
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
		s.sendResponse(msg.ReplyTo, msg.CorrelationId, nil, fmt.Errorf("invalid request: %v", err))
		return
	}

	handler, exists := s.methods[methodName]

	if !exists {
		s.sendResponse(msg.ReplyTo, msg.CorrelationId, nil, fmt.Errorf("method not found: %s", methodName))
		return
	}

	// TODO: handle the kwargs

	result, err := handler(request.Args, request.Kwargs)
	s.sendResponse(msg.ReplyTo, msg.CorrelationId, result, err)
}

// sendResponse sends a response back to the caller.
func (s *Server) sendResponse(replyTo, correlationID string, result interface{}, err error) {
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

	err = s.amqpChannel.Publish(
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
