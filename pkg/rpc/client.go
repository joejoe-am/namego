package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
)

var (
	amqpChannel    *amqp.Channel
	replyQueueName string
	replyQueueID   string
	pendingReplies sync.Map // To track pending replies (correlation_id -> channel)
)

// Client handles RPC communication with a target service.
type Client struct {
	targetService string
}

// Response represents the result of an RPC call.
type Response struct {
	Result interface{} `json:"result"`
	Error  *struct {
		ExcType   string                 `json:"exc_type"`
		ExcPath   string                 `json:"exc_path"`
		ExcArgs   []interface{}          `json:"exc_args"`
		ExcKwargs map[string]interface{} `json:"exc_kwargs"`
		Value     string                 `json:"value"`
	} `json:"error"`
}

func NewClient(serviceName string) *Client {
	return &Client{
		targetService: serviceName,
	}
}

func InitClient(amqpConnection *amqp.Connection) error {
	var err error

	amqpChannel, err = amqpConnection.Channel()
	if err != nil {
		return fmt.Errorf("error creating amqp channel: %v", err)
	}

	err = amqpChannel.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		return fmt.Errorf("error setting Qos: %s", err)
	}

	err = setupReplyQueue()
	if err != nil {
		return err
	}

	go consumeReplies()

	return nil
}

// CallRpc performs the RPC call for the specific service.
func (c *Client) CallRpc(methodName string, args interface{}) (*Response, error) {
	correlationID := uuid.New().String()
	routingKey := fmt.Sprintf("%s.%s", c.targetService, methodName)

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
	pendingReplies.Store(correlationID, replyChan)
	defer pendingReplies.Delete(correlationID)

	// Publish the RPC request
	err = amqpChannel.Publish(
		cfg.ExchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			ReplyTo:       replyQueueID,
			Body:          body,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to publish message: %v", err)
	}

	// Wait for the reply
	select {
	case msg := <-replyChan:
		var response Response
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

		return &Response{Result: response.Result}, nil
	}
}

// Sets up the reply queue for receiving RPC responses.
func setupReplyQueue() error {
	replyQueueID = uuid.New().String()
	replyQueueName = fmt.Sprintf(RpcReplyQueueTemplate, cfg.ServiceName, replyQueueID)

	replyQueue, err := amqpChannel.QueueDeclare(
		replyQueueName,
		true,
		false,
		false,
		false,
		amqp.Table{"x-expires": int32(RpcReplyQueueTtl)},
	)
	if err != nil {
		return fmt.Errorf("failed to declare reply queue: %w", err)
	}

	err = amqpChannel.QueueBind(
		replyQueue.Name,
		replyQueueID,
		cfg.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind reply queue: %w", err)
	}

	return nil
}

// consumeReplies Consumes messages from the reply queue and routes them to the appropriate handlers.
func consumeReplies() {
	messages, err := amqpChannel.Consume(
		replyQueueName,
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
		if ch, ok := pendingReplies.Load(msg.CorrelationId); ok {
			ch.(chan amqp.Delivery) <- msg
			_ = msg.Ack(false)
		} else {
			_ = msg.Nack(false, false)
		}
	}
}
