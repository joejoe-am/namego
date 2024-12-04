package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type Router struct{}

func (r *Router) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/health":
		handleHealthCheck(ctx)
	default:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetBody([]byte("404 - Not Found"))
		log.Warnf("Unhandled route: %s", ctx.Path())
	}
}

func handleHealthCheck(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody([]byte("OK"))
	log.Info("Health check endpoint hit")
}

func main() {
	setupLogger()

	myRouter := &Router{}

	log.Info("Starting server on :8080")
	if err := fasthttp.ListenAndServe(":8080", myRouter.HandleFastHTTP); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

// Configure logging
func setupLogger() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.InfoLevel)
}
