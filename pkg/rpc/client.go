package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/joejoe-am/namego/configs"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
)

var (
	replyQueueName       string
	replyQueueID         string
	ReplyQueueOnce       sync.Once
	globalPendingReplies sync.Map // To track pending replies (correlation_id -> channel)
)

type Rpc struct {
	serviceName    string
	amqpConnection *amqp.Connection
	amqpChannel    *amqp.Channel
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
func ServiceRpc(serviceName string) *Rpc {
	conn, err := amqp.Dial(configs.RabbitMQURL)
	if err != nil {
		panic(fmt.Errorf("failed to connect to RabbitMQ: %v", err))
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(fmt.Errorf("failed to open a channel: %v", err))
	}

	err = ch.Qos(
		1,
		0,
		false,
	)

	if err != nil {
		panic(fmt.Errorf("failed to set QoS: %v", err))
	}

	ReplyQueueOnce.Do(func() {
		replyQueueName = initReplyQueue(ch)
		go consumeReplies(ch)
	})

	rpc := &Rpc{
		serviceName:    serviceName,
		amqpConnection: conn,
		amqpChannel:    ch,
	}

	return rpc
}

// initReplyQueue initializes the reply queue for the specific service.
func initReplyQueue(ch *amqp.Channel) string {
	replyQueueID = uuid.New().String()
	replyQueueName = fmt.Sprintf(RpcReplyQueueTemplate, configs.ServiceName, replyQueueID)

	replyQueue, err := ch.QueueDeclare(
		replyQueueName,
		true,
		false,
		false,
		false,
		amqp.Table{"x-expires": int32(RpcReplyQueueTtl)},
	)

	if err != nil {
		panic(fmt.Errorf("failed to declare shared reply queue: %v", err))
	}

	err = ch.QueueBind(
		replyQueue.Name,
		replyQueueID,
		configs.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		panic(fmt.Errorf("failed to bind reply queue: %v", err))
	}

	return replyQueueName
}

func consumeReplies(ch *amqp.Channel) {
	messages, err := ch.Consume(
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
		if ch, ok := globalPendingReplies.Load(msg.CorrelationId); ok {
			ch.(chan amqp.Delivery) <- msg // Send the message to the appropriate channel
			_ = msg.Ack(false)             // Acknowledge the message
		} else {
			_ = msg.Nack(false, false) // No matching handler, discard the message
		}
	}
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
	globalPendingReplies.Store(correlationID, replyChan)
	defer globalPendingReplies.Delete(correlationID)

	// Publish the RPC request
	err = r.amqpChannel.Publish(
		configs.ExchangeName,
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
