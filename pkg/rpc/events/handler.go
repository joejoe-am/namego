package events

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/joejoe-am/namego/pkg/rpc"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
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
	handlers   map[string]func(body []byte) error // Map of event handlers
}

func generateQueueName(eventCfg EventConfig) (string, bool, bool) {
	var queueName string
	exclusive := false
	autoDelete := !eventCfg.ReliableDelivery

	switch eventCfg.HandlerType {
	case ServicePool:
		queueName = fmt.Sprintf(
			rpc.EventHandlerServicePoolQueueTemplate,
			eventCfg.SourceService,
			eventCfg.EventType,
			rpc.Cfg.ServiceName,
			getFunctionName(eventCfg.HandlerFunction),
		)
	case Singleton:
		queueName = fmt.Sprintf(
			rpc.EventHandlerSingletonCaseQueueTemplate,
			eventCfg.SourceService,
			eventCfg.EventType,
		)
	case Broadcast:
		if eventCfg.BroadcastID == "" {
			eventCfg.BroadcastID = uuid.New().String()
		}
		queueName = fmt.Sprintf(
			rpc.EventHandlerBroadCaseQueueTemplate,
			eventCfg.SourceService,
			eventCfg.EventType,
			rpc.Cfg.ServiceName,
			getFunctionName(eventCfg.HandlerFunction),
			eventCfg.BroadcastID,
		)
		exclusive = !eventCfg.ReliableDelivery
	}

	return queueName, exclusive, autoDelete
}

// NewEventHandler initializes a new event handler.
func NewEventHandler(cfg EventConfig) (*EventHandler, error) {
	queueName, exclusive, autoDelete := generateQueueName(cfg)

	return &EventHandler{
		config:     cfg,
		queueName:  queueName,
		exclusive:  exclusive,
		autoDelete: autoDelete,
		handlers:   make(map[string]func(body []byte) error),
	}, nil
}

func (h *EventHandler) Start(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("failed to open RabbitMQ channel: %v", err)
		return err
	}

	err = h.SetupQueue(ch)

	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		h.queueName,
		"",
		!h.config.RequeueOnError,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("failed to consume RabbitMQ events: %v", err)
		return err
	}

	for msg := range msgs {
		go h.handleMessage(msg)
	}

	return nil
}

func (h *EventHandler) SetupQueue(ch *amqp.Channel) error {
	exchange := fmt.Sprintf("%s.events", h.config.SourceService)

	err := ch.ExchangeDeclare(
		exchange,
		"topic",
		true,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("failed to declare exchange: %v", err)
		return err
	}

	queue, err := ch.QueueDeclare(
		h.queueName,
		true,
		h.autoDelete,
		h.exclusive,
		false,
		nil,
	)
	if err != nil {
		log.Printf("failed to declare queue: %v", err)
		return err
	}

	h.queue = &queue

	err = ch.QueueBind(
		h.queueName,
		h.config.EventType,
		exchange,
		false,
		nil,
	)
	if err != nil {
		log.Printf("failed to bind queue: %v", err)
		return err
	}

	return nil
}

func (h *EventHandler) handleMessage(msg amqp.Delivery) {
	handler := h.config.HandlerFunction

	err := handler(msg.Body)
	if err != nil {
		log.Printf("handler error: %v\n", err)
		_ = msg.Nack(false, h.config.RequeueOnError)
		return
	}

	_ = msg.Ack(false)
}
