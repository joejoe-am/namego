package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
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
		log.Printf("error creating amqp channel: %v", err)
		return err
	}

	err = amqpChannel.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		log.Printf("error setting qos: %v", err)
		return err
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
		fmt.Printf("error marshalling payload: %v", err)
		return nil, err
	}

	// Create a channel to wait for the response
	replyChan := make(chan amqp.Delivery, 1)
	pendingReplies.Store(correlationID, replyChan)
	defer pendingReplies.Delete(correlationID)

	// Publish the RPC request
	err = amqpChannel.Publish(
		Cfg.ExchangeName,
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
		log.Printf("failed to publish message: %v", err)
		return nil, err
	}

	// Wait for the reply
	select {
	case msg := <-replyChan:
		var response Response
		if err = json.Unmarshal(msg.Body, &response); err != nil {
			log.Printf("failed to decode response: %v", err)
			return nil, err
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
	replyQueueName = fmt.Sprintf(RpcReplyQueueTemplate, Cfg.ServiceName, replyQueueID)

	replyQueue, err := amqpChannel.QueueDeclare(
		replyQueueName,
		true,
		false,
		false,
		false,
		amqp.Table{"x-expires": int32(RpcReplyQueueTtl)},
	)
	if err != nil {
		log.Printf("failed to declare reply queue: %s", err)
		return err
	}

	err = amqpChannel.QueueBind(
		replyQueue.Name,
		replyQueueID,
		Cfg.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		log.Printf("failed to bind reply queue: %s", err)
		return err
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
			// TODO: this could make a loop to requeue, NACK a message only 3 times for example
			_ = msg.Nack(false, false)
		}
	}
}
