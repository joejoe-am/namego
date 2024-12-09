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

	authRpc := rpc.ServiceRpc("authnzng")
	quotaRpc := rpc.ServiceRpc("quota")

	response, err := authRpc.CallRpc("joe", map[string]string{})
	fmt.Println(response, err)

	response, err = quotaRpc.CallRpc("health_check", map[string]string{})
	fmt.Println(response, err)

	server.Get("/health", func(ctx *fasthttp.RequestCtx) { ctx.WriteString("OK") })
	server.Get("/auth-health", AuthHealthHandler(authRpc), LoggingMiddleware)

	rpcServer.RegisterMethod("multiply", Multiply)

	rpcServer.Start()

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
