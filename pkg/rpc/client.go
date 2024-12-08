package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/joejoe-am/namego/configs"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
)

type Rpc struct {
	serviceName    string
	amqpConnection *amqp.Connection
	amqpChannel    *amqp.Channel
	replyQueueName string
	replyQueueID   string
	pendingReplies sync.Map // To track pending replies (correlation_id -> channel)
}

type RPCResponse struct {
	Result interface{} `json:"result"`
	Error  *struct {
		ExcType   string                 `json:"exc_type"`
		ExcPath   string                 `json:"exc_path"`
		ExcArgs   []interface{}          `json:"exc_args"`
		ExcKwargs map[string]interface{} `json:"exc_kwargs"`
		Value     string                 `json:"value"`
	} `json:"error"`
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

	err = ch.Qos(
		1,
		0,
		false,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %v", err)
	}

	rpc := &Rpc{
		serviceName:    serviceName,
		amqpConnection: conn,
		amqpChannel:    ch,
	}

	if err := rpc.initReplyQueue(); err != nil {
		rpc.Close()
		return nil, err
	}

	// Start a single goroutine to consume messages from the reply queue
	go rpc.consumeReplies()

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

	// Create a channel to wait for the response
	replyChan := make(chan amqp.Delivery, 1)
	r.pendingReplies.Store(correlationID, replyChan)
	defer r.pendingReplies.Delete(correlationID)

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

	// Wait for the reply
	select {
	case msg := <-replyChan:
		var response RPCResponse
		if err = json.Unmarshal(msg.Body, &response); err != nil {
			return nil, fmt.Errorf("failed to decode response: %v", err)
		}

		if response.Error != nil {
			// Convert the error to a Go error type
			return nil, fmt.Errorf(
				"RPC Error: %s (type: %s, path: %s, args: %v, kwargs: %v)",
				response.Error.Value,
				response.Error.ExcType,
				response.Error.ExcPath,
				response.Error.ExcArgs,
				response.Error.ExcKwargs,
			)
		}

		return &RPCResponse{Result: response.Result}, nil
	}
}

func (r *Rpc) consumeReplies() {
	messages, err := r.amqpChannel.Consume(
		r.replyQueueName,
		"",
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Printf("failed to consume messages: %v\n", err)
		return
	}

	for msg := range messages {
		if ch, ok := r.pendingReplies.Load(msg.CorrelationId); ok {
			ch.(chan amqp.Delivery) <- msg // Send the message to the appropriate channel
			_ = msg.Ack(false)             // Acknowledge the message
		} else {
			_ = msg.Nack(false, false) // No matching handler, discard the message
		}
	}
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
