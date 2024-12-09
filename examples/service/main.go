package main

import (
	"fmt"
	"github.com/joejoe-am/namego/pkg/rpc"
	"github.com/joejoe-am/namego/pkg/web"
	"github.com/valyala/fasthttp"
)

// TODO: change package name

func main() {
	server := web.New()
	rpcServer := rpc.NewRpcServer("nameko")

	rpcClient := rpc.NewClient()
	authRpc := rpcClient.TargetService("authnzng")
	quotaRpc := rpcClient.TargetService("quota")

	response, err := authRpc.CallRpc("joe", map[string]string{})
	fmt.Println(response, err)

	response, err = quotaRpc.CallRpc("health_check", map[string]string{})
	fmt.Println(response, err)

	server.Get("/health", func(ctx *fasthttp.RequestCtx) { ctx.WriteString("OK") })
	server.Get("/auth-health", AuthHealthHandler(authRpc), LoggingMiddleware)

	rpcServer.RegisterMethod("multiply", Multiply)

	go func() {
		fmt.Println("Server running on :8080")
		if err := server.Listen(":8080"); err != nil {
			panic(fmt.Sprintf("Error starting server: %v\n", err))
		}
	}()

	rpcServer.Start()
}
