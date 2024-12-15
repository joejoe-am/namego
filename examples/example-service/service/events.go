package service

import (
	"encoding/json"
	"fmt"
	"github.com/joejoe-am/namego/pkg/rpc/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

func EventHandlerFunction(body []byte) error {
	fmt.Printf("Received event: %s\n", string(body))
	return nil
}

func DispatchEventExampleFunction(conn *amqp.Connection, serviceName string) {
	eventData := map[string]interface{}{
		"id":   "12345",
		"name": "example",
	}
	payload, err := json.Marshal(eventData)
	if err != nil {
		fmt.Printf("Failed to marshal event data: %v", err)
	}

	err = events.Dispatch(conn, serviceName, "TEST_EVENT_HANDLER", payload)
	if err != nil {
		fmt.Printf("Failed to dispatch event: %v\n", err)
	} else {
		fmt.Println("Event successfully dispatched.")
	}
}
