package gateway

import (
	"encoding/json"
	"fmt"
	"github.com/joejoe-am/namego/pkg/rpc"
	"github.com/joejoe-am/namego/pkg/web"
	"github.com/valyala/fasthttp"
)

func HealthHandler(ctx *fasthttp.RequestCtx) {
	ctx.WriteString("OK")
}

func AuthHealthHandler(authRpc *rpc.Client) web.Handler {
	return func(ctx *fasthttp.RequestCtx) {
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
	}
}
