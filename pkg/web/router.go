package web

import (
	"github.com/valyala/fasthttp"
	"strings"
)

type Router interface {
	Add(methods []string, path string, handler Handler, middleware ...Handler) Router

	Get(path string, handler Handler, middleware ...Handler) Router
	Post(path string, handler Handler, middleware ...Handler) Router
	Put(path string, handler Handler, middleware ...Handler) Router
	Delete(path string, handler Handler, middleware ...Handler) Router
	Patch(path string, handler Handler, middleware ...Handler) Router
}

type Route struct {
	Methods    []string
	Path       string
	Handler    Handler
	Middleware []Handler // Middleware handlers
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
