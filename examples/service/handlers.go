package main

import (
	"encoding/json"
	"fmt"
	"github.com/joejoe-am/namego/pkg/rpc"
	"github.com/joejoe-am/namego/pkg/web"
	"github.com/valyala/fasthttp"
)

func AuthHealthHandler(authRpc *rpc.Rpc) web.Handler {
	return func(ctx *fasthttp.RequestCtx) {
		response, err := authRpc.CallRpc("health_check", map[string]string{})
		if err != nil {
			// Handle RPC error and respond with HTTP 500 status
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetContentType("application/json")
			ctx.WriteString(fmt.Sprintf(`{"error": "%s"}`, err.Error()))
			return
		}

		// Respond with the RPC result as JSON
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		if response.Result != nil {
			if jsonResponse, err := json.Marshal(response.Result); err == nil {
				ctx.Write(jsonResponse)
			} else {
				// Handle JSON marshalling error
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				ctx.WriteString(fmt.Sprintf(`{"error": "failed to encode response: %s"}`, err.Error()))
			}
		} else {
			// Handle case where response.Result is nil
			ctx.WriteString(`{"status": "no result"}`)
		}
	}
}

// Multiply is an example handler for your RPC server.
func Multiply(args interface{}, kwargs map[string]interface{}) (interface{}, error) {
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
