package main

import (
	"fmt"
	"github.com/joejoe-am/namego/pkg/rpc"
)

// TODO: change package name

func main() {
	const name = "joe"

	authRpc, err := rpc.ServiceRpc("authnzng")
	if err != nil {
		panic(err)
	}

	quotaRpc, err := rpc.ServiceRpc("quota")

	if err != nil {
		panic(err)
	}

	response, err := authRpc.CallRpc("health_check", map[string]string{})
	fmt.Println(response, err)

	response, err = quotaRpc.CallRpc("health_check", map[string]string{})
	fmt.Println(response, err)

	response, err = authRpc.CallRpc("joe", map[string]string{})
	fmt.Println(response, err)

	//select {}

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

	//server := service.New()
	//
	//server.Get("/health", func(ctx *fasthttp.RequestCtx) { ctx.WriteString("OK") })
	//server.Get("/m2", MultipleTwo)
	//
	//// TODO: this should be a run method that start the whole application
	//
	//stop := make(chan os.Signal, 1)
	//signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	//<-stop
	//
	//fmt.Println("Shutting down server...")
}
