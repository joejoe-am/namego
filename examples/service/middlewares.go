package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
)

func LoggingMiddleware(ctx *fasthttp.RequestCtx) {
	fmt.Printf("Request: %s %s\n", string(ctx.Method()), string(ctx.Path()))
}
