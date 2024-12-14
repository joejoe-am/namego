package rpc

import (
	"fmt"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type HandlerType string
type EventHandlerType func(body []byte) error

const (
	ServicePool HandlerType = "SERVICE_POOL"
	Singleton   HandlerType = "SINGLETON"
	Broadcast   HandlerType = "BROADCAST"
)

type EventConfig struct {
	SourceService    string
	EventType        string
	HandlerType      HandlerType
	ReliableDelivery bool
	RequeueOnError   bool
	BroadcastID      string // Used for Broadcast queues
	HandlerFunction  EventHandlerType
}

type EventHandler struct {
	config     EventConfig
	queueName  string
	exclusive  bool
	autoDelete bool
	queue      *amqp.Queue
	connection *amqp.Connection
	channel    *amqp.Channel
	handlers   map[string]func(body []byte) error // Map of event handlers
}

func generateQueueName(eventCfg EventConfig) (string, bool, bool) {
	var queueName string
	exclusive := false
	autoDelete := !eventCfg.ReliableDelivery

	switch eventCfg.HandlerType {
	case ServicePool:
		queueName = fmt.Sprintf(
			EventHandlerServicePoolQueueTemplate,
			eventCfg.SourceService,
			eventCfg.EventType,
			cfg.ServiceName,
			GetFunctionName(eventCfg.HandlerFunction),
		)
	case Singleton:
		queueName = fmt.Sprintf(
			EventHandlerSingletonCaseQueueTemplate,
			eventCfg.SourceService,
			eventCfg.EventType,
		)
	case Broadcast:
		if eventCfg.BroadcastID == "" {
			eventCfg.BroadcastID = uuid.New().String()
		}
		queueName = fmt.Sprintf(
			EventHandlerBroadCaseQueueTemplate,
			eventCfg.SourceService,
			eventCfg.EventType,
			cfg.ServiceName,
			GetFunctionName(eventCfg.HandlerFunction),
			eventCfg.BroadcastID,
		)
		exclusive = !eventCfg.ReliableDelivery
	}

	return queueName, exclusive, autoDelete
}

// NewEventHandler initializes a new event handler.
func NewEventHandler(cfg EventConfig, conn *amqp.Connection) (*EventHandler, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open RabbitMQ channel: %v", err)
	}

	queueName, exclusive, autoDelete := generateQueueName(cfg)

	return &EventHandler{
		config:     cfg,
		queueName:  queueName,
		exclusive:  exclusive,
		autoDelete: autoDelete,
		connection: conn,
		channel:    channel,
		handlers:   make(map[string]func(body []byte) error),
	}, nil
}

func (h *EventHandler) SetupQueue() error {
	exchange := fmt.Sprintf("%s.events", h.config.SourceService)

	err := h.channel.ExchangeDeclare(
		exchange,
		"topic",
		true,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %v", err)
	}

	queue, err := h.channel.QueueDeclare(
		h.queueName,
		true,
		h.autoDelete,
		h.exclusive,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}

	h.queue = &queue

	err = h.channel.QueueBind(
		h.queueName,
		h.config.EventType,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %v", err)
	}

	return nil
}

func (h *EventHandler) Start() error {
	msgs, err := h.channel.Consume(
		h.queueName,
		"",
		!h.config.RequeueOnError,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %v", err)
	}

	for msg := range msgs {
		go h.handleMessage(msg)
	}

	return nil
}

func (h *EventHandler) handleMessage(msg amqp.Delivery) {
	handler := h.config.HandlerFunction

	err := handler(msg.Body)
	if err != nil {
		fmt.Printf("handler error: %v\n", err)
		_ = msg.Nack(false, h.config.RequeueOnError)
		return
	}

	_ = msg.Ack(false)
}
