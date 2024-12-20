package rpc

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
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
func (s *Server) Start() error {
	var err error

	queueName := fmt.Sprintf(RpcQueueTemplate, s.serviceName)
	routingKey := fmt.Sprintf("%s.*", s.serviceName)

	s.amqpChannel, err = s.amqpConnection.Channel()

	if err != nil {
		log.Printf("error creating amqp channel: %v", err)
		return err
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
		log.Printf("failed to declare RPC queue: %v", err)
		return err
	}

	err = s.amqpChannel.QueueBind(
		queueName,
		routingKey,
		Cfg.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		log.Printf("failed to bind RPC queue: %v", err)
		return err
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
		log.Printf("failed to consume messages: %v", err)
		return err
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
			log.Printf("panic occurred: %v\n", r)
			// TODO: this could make a loop to requeue, NACK a message only 3 times for example
			err := msg.Nack(false, false)
			if err != nil {
				log.Printf("failed to nack message: %v", err)
			}
		}
	}()

	err := s.processMessage(msg)
	if err != nil {
		log.Printf("error processing message: %v\n", err)
	}

	err = msg.Ack(false)

	if err != nil {
		log.Printf("error acknowledging message: %v\n", err)
	}
}

// processMessage decodes and handles the message logic.
func (s *Server) processMessage(msg amqp.Delivery) error {
	routingKeyParts := strings.Split(msg.RoutingKey, ".")

	if len(routingKeyParts) != 2 {
		return s.sendResponse(msg, nil, fmt.Errorf("invalid routing key: %s", msg.RoutingKey))
	}

	methodName := routingKeyParts[1]

	var request struct {
		Args   interface{}            `json:"args"`
		Kwargs map[string]interface{} `json:"kwargs"`
	}
	if err := json.Unmarshal(msg.Body, &request); err != nil {
		return s.sendResponse(msg, nil, fmt.Errorf("invalid request: %v", err))
	}

	handler, exists := s.methods[methodName]
	if !exists {
		return s.sendResponse(msg, nil, fmt.Errorf("method not found: %s", methodName))
	}

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
		log.Printf("failed to serialize response: %v", marshalErr)
		return marshalErr
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
		log.Printf("failed to send response: %v", publishErr)
		return err
	}

	return nil
}
