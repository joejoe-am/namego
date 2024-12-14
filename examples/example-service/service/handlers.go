package service

import (
	"fmt"
	"time"
)

func Multiply(args interface{}, kwargs map[string]interface{}) (interface{}, error) {
	time.Sleep(10 * time.Second)
	argsList, ok := args.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid args: expected []interface{} but got %T", args)
	}

	var nums []float64
	for _, v := range argsList {
		num, ok := v.(float64) // Assert each element to float64
		if !ok {
			return nil, fmt.Errorf("invalid element in args: expected float64 but got %T", v)
		}
		nums = append(nums, num)
	}

	product := 1.0
	for _, num := range nums {
		product *= num
	}

	return product, nil
}
