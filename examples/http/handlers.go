package main

import (
	"github.com/valyala/fasthttp"
	"strconv"
)

func MultipleTwo(ctx *fasthttp.RequestCtx) {
	number, _ := strconv.Atoi(string(ctx.QueryArgs().Peek("number")))
	number += 2
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString(strconv.Itoa(number))
}
