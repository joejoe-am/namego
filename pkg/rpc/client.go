package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/joejoe-am/namego/configs"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Rpc struct {
	serviceName    string
	amqpConnection *amqp.Connection
	amqpChannel    *amqp.Channel
	replyQueueName string
	replyQueueID   string
}

type RPCResponse struct {
	Result interface{} `json:"result"`
	Err    error       `json:"error"`
}

// ServiceRpc initializes and returns an Rpc instance for a specific service.
func ServiceRpc(serviceName string) (*Rpc, error) {
	conn, err := amqp.Dial(configs.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}

	rpc := &Rpc{
		serviceName:    serviceName,
		amqpConnection: conn,
		amqpChannel:    ch,
	}

	return rpc, nil
}

// initReplyQueue initializes the reply queue for the specific service.
func (r *Rpc) initReplyQueue() error {
	r.replyQueueID = uuid.New().String()
	r.replyQueueName = fmt.Sprintf(RpcReplyQueueTemplate, configs.ServiceName, r.replyQueueID)

	replyQueue, err := r.amqpChannel.QueueDeclare(
		r.replyQueueName,
		true,
		false,
		false,
		false,
		amqp.Table{"x-expires": int32(RpcReplyQueueTtl)},
	)
	fmt.Printf("Declaring queue %s", r.replyQueueName)

	if err != nil {
		return fmt.Errorf("failed to declare reply queue: %v", err)
	}

	err = r.amqpChannel.QueueBind(
		replyQueue.Name,
		r.replyQueueID,
		configs.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind reply queue: %v", err)
	}
	return nil
}

// CallRpc performs the RPC call for the specific service.
func (r *Rpc) CallRpc(methodName string, args interface{}) (*RPCResponse, error) {
	if err := r.initReplyQueue(); err != nil {
		r.Close()
		return nil, err
	}

	correlationID := uuid.New().String()
	routingKey := fmt.Sprintf("%s.%s", r.serviceName, methodName)

	payload := map[string]interface{}{
		"args":   args,
		"kwargs": map[string]interface{}{},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal arguments: %v", err)
	}

	// Publish the RPC request
	err = r.amqpChannel.Publish(
		configs.ExchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			ReplyTo:       r.replyQueueID,
			Body:          body,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to publish message: %v", err)
	}

	messages, err := r.amqpChannel.Consume(
		r.replyQueueName,
		"",
		true,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume messages: %v", err)
	}

	// Wait for the reply with the correct correlation ID
	for msg := range messages {
		if msg.CorrelationId == correlationID {
			var response RPCResponse
			if err := json.Unmarshal(msg.Body, &response); err != nil {
				return nil, fmt.Errorf("failed to decode response: %v", err)
			}
			return &response, nil
		}
	}

	return nil, fmt.Errorf("no response received for correlation ID %s", correlationID)
}

// Close gracefully closes the AMQP channel and connection.
func (r *Rpc) Close() {
	if r.amqpChannel != nil {
		err := r.amqpChannel.Close()
		if err != nil {
			return
		}
	}
	if r.amqpConnection != nil {
		err := r.amqpConnection.Close()
		if err != nil {
			return
		}
	}
}
