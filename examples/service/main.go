package main

import (
	"encoding/json"
	"fmt"
	"github.com/joejoe-am/namego/pkg/http"
	"github.com/joejoe-am/namego/pkg/rpc"
	"github.com/valyala/fasthttp"
)

// TODO: change package name

func main() {
	const name = "joe"

	server := http.New()

	authRpc := rpc.ServiceRpc("authnzng")
	quotaRpc := rpc.ServiceRpc("quota")

	response, err := authRpc.CallRpc("health_check", map[string]string{})
	fmt.Println(response, err)

	response, err = authRpc.CallRpc("joe", map[string]string{})
	fmt.Println(response, err)

	response, err = quotaRpc.CallRpc("health_check", map[string]string{})
	fmt.Println(response, err)

	server.Get("/health", func(ctx *fasthttp.RequestCtx) { ctx.WriteString("OK") })

	server.Get("/auth-health", func(ctx *fasthttp.RequestCtx) {
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
	})

	select {}

	//rpcServer, err := rpc.NewRpcServer(name)
	//
	//if err != nil {
	//	panic(err)
	//}
	//
	//rpcServer.RegisterMethod("Multiply", Multiply)

	//if err := rpcServer.Start(); err != nil {
	//	log.Fatalf("Failed to start RPC server: %v", err)
	//}

	//
	//// TODO: this should be a run method that start the whole application
	//
	//stop := make(chan os.Signal, 1)
	//signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	//<-stop
	//
	//fmt.Println("Shutting down server...")
}
