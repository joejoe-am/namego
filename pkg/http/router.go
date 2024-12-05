package http

import (
	"github.com/valyala/fasthttp"
	"strings"
)

type Route struct {
	Methods []string
	Path    string
	Handler Handler
}

// matchRoute checks if the incoming request matches a registered route
func matchRoute(ctx *fasthttp.RequestCtx, route Route) bool {
	// Check if the HTTP method matches (case-insensitive)
	for _, method := range route.Methods {
		if strings.EqualFold(method, string(ctx.Method())) {
			// Check if the path matches
			if route.Path == string(ctx.Path()) {
				return true
			}
		}
	}
	return false
}
