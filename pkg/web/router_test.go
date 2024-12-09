package web

import (
	"github.com/valyala/fasthttp"
	"testing"
)

func TestMatchRoute(t *testing.T) {
	tests := []struct {
		name           string
		route          Route
		requestMethod  string
		requestPath    string
		expectedResult bool
	}{
		{
			name: "Exact match - GET /health",
			route: Route{
				Methods: []string{"GET", "POST"},
				Path:    "/health",
			},
			requestMethod:  "GET",
			requestPath:    "/health",
			expectedResult: true,
		},
		{
			name: "Method mismatch - POST /health",
			route: Route{
				Methods: []string{"GET"},
				Path:    "/health",
			},
			requestMethod:  "POST",
			requestPath:    "/health",
			expectedResult: false,
		},
		{
			name: "Path mismatch - GET /status",
			route: Route{
				Methods: []string{"GET"},
				Path:    "/health",
			},
			requestMethod:  "GET",
			requestPath:    "/status",
			expectedResult: false,
		},
		{
			name: "Case insensitive match - GET /Health",
			route: Route{
				Methods: []string{"get"},
				Path:    "/health",
			},
			requestMethod:  "GET",
			requestPath:    "/Health",
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Mock the request context
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.SetRequestURI(test.requestPath)
			ctx.Request.Header.SetMethod(test.requestMethod)

			// Perform the matchRoute test
			result := matchRoute(ctx, test.route)

			// Validate the result
			if result != test.expectedResult {
				t.Errorf("expected %v, got %v", test.expectedResult, result)
			}
		})
	}

}
