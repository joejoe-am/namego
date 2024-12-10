package service

import "fmt"

func EventHandlerFunction(body []byte) error {
	fmt.Printf("Received event: %s\n", string(body))
	return nil
}
