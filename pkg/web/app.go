package web

import (
	"github.com/valyala/fasthttp"
)

// Handler returns the server handler.
type Handler func(ctx *fasthttp.RequestCtx)

type Server struct {
	server *fasthttp.Server
	routes []Route
}

// Config holds the configuration for the HTTP server
type Config struct {
	Addr string // Address to bind the server to
}

func New() *Server {
	http := &Server{
		routes: make([]Route, 0),
	}

	http.init()

	return http
}

func (s *Server) init() *Server {
	s.server = &fasthttp.Server{}

	// fasthttp server settings
	s.server.Handler = s.Handler()

	return s
}

func (s *Server) Handler() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		for _, route := range s.routes {
			if matchRoute(ctx, route) {
				// Execute middleware in sequence
				for _, mw := range route.Middleware {
					mw(ctx)
					if ctx.Response.StatusCode() != fasthttp.StatusOK {
						// Stop processing if a middleware has set an error status
						return
					}
				}

				// Call the main handler
				route.Handler(ctx)
				return
			}
		}
		// Default 404 handler
		ctx.Error("Not Found", fasthttp.StatusNotFound)
	}
}

// Add allows you to specify multiple HTTP methods to register a route.
func (s *Server) Add(methods []string, path string, handler Handler, middleware ...Handler) Router {
	s.routes = append(s.routes, Route{
		Methods:    methods,
		Path:       path,
		Handler:    handler,
		Middleware: middleware,
	})
	return s
}

// Get registers a route for GET methods that requests a representation
// of the specified resource. Requests using GET should only retrieve data.
func (s *Server) Get(path string, handler Handler, middleware ...Handler) Router {
	return s.Add([]string{MethodGet}, path, handler, middleware...)
}

// Head registers a route for HEAD methods that asks for a response identical
// to that of a GET request, but without the response body.
func (s *Server) Head(path string, handler Handler, middleware ...Handler) Router {
	return s.Add([]string{MethodHead}, path, handler, middleware...)
}

// Post registers a route for POST methods that is used to submit an entity to the
// specified resource, often causing a change in state or side effects on the server.
func (s *Server) Post(path string, handler Handler, middleware ...Handler) Router {
	return s.Add([]string{MethodPost}, path, handler, middleware...)
}

// Put registers a route for PUT methods that replaces all current representations
// of the target resource with the request payload.
func (s *Server) Put(path string, handler Handler, middleware ...Handler) Router {
	return s.Add([]string{MethodPut}, path, handler, middleware...)
}

// Patch registers a route for PATCH methods that is used to apply partial
// modifications to a resource.
func (s *Server) Patch(path string, handler Handler, middleware ...Handler) Router {
	return s.Add([]string{MethodPatch}, path, handler, middleware...)
}

// Options registers a route for OPTIONS methods that is used to describe the
// communication options for the target resource.
func (s *Server) Options(path string, handler Handler, middleware ...Handler) Router {
	return s.Add([]string{MethodOptions}, path, handler, middleware...)
}

// Delete registers a route for DELETE methods that deletes the specified resource.
func (s *Server) Delete(path string, handler Handler, middleware ...Handler) Router {
	return s.Add([]string{MethodDelete}, path, handler, middleware...)
}
