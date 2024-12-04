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
	if !contains(route.Methods, string(ctx.Method())) {
		return false
	}
	if route.Path != string(ctx.Path()) {
		return false
	}
	return true
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}
