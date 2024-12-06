package http

import (
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"sync"
)

type Router interface {
	Add(methods []string, path string, handler Handler) Router
	Get(path string, handler Handler) Router
}

// Handler returns the server handler.
type Handler func(ctx *fasthttp.RequestCtx)

type Http struct {
	server *fasthttp.Server
	mutex  sync.Mutex
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

	http.init()

	// handle the thread kill
	// think of a better way.

	go func() {
		log.Println("Server running on :8080")
		if err := http.Listen(":8080"); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	return http
}

func (http *Http) init() *Http {
	// lock application
	http.mutex.Lock()

	// create fasthttp server
	http.server = &fasthttp.Server{}

	// fasthttp server settings
	http.server.Handler = http.Handler()

	// unlock application
	http.mutex.Unlock()
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

// Add allows you to specify multiple HTTP methods to register a route.
func (http *Http) Add(methods []string, path string, handler Handler) Router {
	http.routes = append(http.routes, Route{
		Methods: methods,
		Path:    path,
		Handler: handler,
	})
	return http
}

// Get registers a route for GET methods that requests a representation
// of the specified resource. Requests using GET should only retrieve data.
func (http *Http) Get(path string, handler Handler) Router {
	return http.Add([]string{MethodGet}, path, handler)
}

// Head registers a route for HEAD methods that asks for a response identical
// to that of a GET request, but without the response body.
func (http *Http) Head(path string, handler Handler) Router {
	return http.Add([]string{MethodHead}, path, handler)
}

// Post registers a route for POST methods that is used to submit an entity to the
// specified resource, often causing a change in state or side effects on the server.
func (http *Http) Post(path string, handler Handler) Router {
	return http.Add([]string{MethodPost}, path, handler)
}

// Put registers a route for PUT methods that replaces all current representations
// of the target resource with the request payload.
func (http *Http) Put(path string, handler Handler) Router {
	return http.Add([]string{MethodPut}, path, handler)
}

// Delete registers a route for DELETE methods that deletes the specified resource.
func (http *Http) Delete(path string, handler Handler) Router {
	return http.Add([]string{MethodDelete}, path, handler)
}
