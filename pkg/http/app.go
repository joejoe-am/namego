package http

import (
	"github.com/valyala/fasthttp"
)

const (
	MethodGet    = "GET"
	MethodPost   = "POST"
	MethodPut    = "PUT"
	MethodDelete = "DELETE"
)

type Router interface {
	Add(methods []string, path string, handler Handler) Router
	Get(path string, handler Handler) Router
}

// Handler returns the server handler.
type Handler func(ctx *fasthttp.RequestCtx)

type Http struct {
	server *fasthttp.Server
	routes []Route
}

// Config holds the configuration for the HTTP server
type Config struct {
	Addr string // Address to bind the server to
}

func New(config ...Config) *Http {
	http := &Http{
		routes: make([]Route, 0),
	}

	if len(config) > 0 && config[0].Addr != "" {
		http.server = &fasthttp.Server{
			Handler: http.Handler(),
		}
	} else {
		http.server = &fasthttp.Server{
			Handler: http.Handler(),
		}
	}

	return http
}

func (http *Http) Handler() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		for _, route := range http.routes {
			if matchRoute(ctx, route) {
				route.Handler(ctx)
				return
			}
		}
		// Default 404 handler
		ctx.Error("Not Found", fasthttp.StatusNotFound)
	}
}

func (http *Http) Add(methods []string, path string, handler Handler) Router {
	http.routes = append(http.routes, Route{
		Methods: methods,
		Path:    path,
		Handler: handler,
	})
	return http
}

func (http *Http) Get(path string, handler Handler) Router {
	return http.Add([]string{MethodGet}, path, handler)
}

func (http *Http) Post(path string, handler Handler) Router {
	return http.Add([]string{MethodPost}, path, handler)
}

func (http *Http) Put(path string, handler Handler) Router {
	return http.Add([]string{MethodPut}, path, handler)
}

func (http *Http) Delete(path string, handler Handler) Router {
	return http.Add([]string{MethodDelete}, path, handler)
}
