package main

import (
	"fmt"
	"github.com/joejoe-am/namego/pkg/http"
	"github.com/valyala/fasthttp"
	"os"
	"os/signal"
	"syscall"
)

// TODO: change package name

func main() {
	server := http.New()

	server.Get("/health", func(ctx *fasthttp.RequestCtx) { ctx.WriteString("OK") })
	server.Get("/m2", MultipleTwo)

	// TODO: this should be a run method that start the whole application

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("Shutting down server...")
}
