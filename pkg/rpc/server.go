package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/joejoe-am/namego/configs"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RpcServer struct {
	serviceName    string
	amqpConnection *amqp.Connection
	amqpChannel    *amqp.Channel
	methods        map[string]func(args interface{}) (interface{}, error)
}

// NewRpcServer initializes and returns an RpcServer instance.
func NewRpcServer(serviceName string) (*RpcServer, error) {
	conn, err := amqp.Dial(configs.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}

	return &RpcServer{
		serviceName:    serviceName,
		amqpConnection: conn,
		amqpChannel:    ch,
		methods:        make(map[string]func(args interface{}) (interface{}, error)),
	}, nil
}

// RegisterMethod registers an RPC method with the server.
func (r *RpcServer) RegisterMethod(methodName string, handler func(args interface{}) (interface{}, error)) {
	r.methods[methodName] = handler
}

// Start begins listening for RPC requests on the service's queue.
func (r *RpcServer) Start() error {
	queueName := fmt.Sprintf(RpcQueueTemplate, r.serviceName)

	_, err := r.amqpChannel.QueueDeclare(
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
		return fmt.Errorf("failed to consume messages: %v", err)
	}

	go func() {
		for msg := range msgs {
			go r.handleRequest(msg)
		}
	}()

	return nil
}

// handleRequest processes incoming messages and invokes the appropriate handler.
func (r *RpcServer) handleRequest(msg amqp.Delivery) {
	var request struct {
		MethodName string      `json:"method_name"`
		Args       interface{} `json:"args"`
	}

	if err := json.Unmarshal(msg.Body, &request); err != nil {
		r.sendResponse(msg.ReplyTo, msg.CorrelationId, nil, fmt.Errorf("invalid request: %v", err))
		return
	}

	handler, exists := r.methods[request.MethodName]

	if !exists {
		r.sendResponse(msg.ReplyTo, msg.CorrelationId, nil, fmt.Errorf("method not found: %s", request.MethodName))
		return
	}

	result, err := handler(request.Args)
	r.sendResponse(msg.ReplyTo, msg.CorrelationId, result, err)
}

// sendResponse sends a response back to the caller.
func (r *RpcServer) sendResponse(replyTo, correlationID string, result interface{}, err error) {
	response := RPCResponse{
		Result: result,
		Err:    err,
	}

	body, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("failed to serialize response: %v\n", err)
		return
	}

	err = r.amqpChannel.Publish(
		"",
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
func (r *RpcServer) Close() {
	if r.amqpChannel != nil {
		_ = r.amqpChannel.Close()
	}
	if r.amqpConnection != nil {
		_ = r.amqpConnection.Close()
	}
}
