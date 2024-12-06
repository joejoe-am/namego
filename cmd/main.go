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

	select {} // fix it later

	//authRpc, err := rpc.NewRpc("auth")
	//
	//response, err := authRpc.CallRpc("health_check", map[string]interface{}{})
	//
	//if err != nil {
	//	return
	//}
	//fmt.Println(response)
	//
	//if err != nil {
	//	log.Fatal(err)
	//}
}

// Configure logging
func setupLogger() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.InfoLevel)
}
