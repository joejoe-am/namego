package main

import (
	"github.com/joejoe-am/american-nameko/pkg/http"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

func main() {
	setupLogger()

	server := http.New()

	server.Get("/health", func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("OK")
	})

	log.Println("Server running on :8080")
	if err := server.Listen(":8080"); err != nil {
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
