package httpx

import "net/http"

// Chain is a helper for chaining middleware handlers together for easier
// management.
type Chain []func(http.Handler) http.Handler

// Use appends a handler to the middleware chain.
func (c *Chain) Use(handler func(http.Handler) http.Handler) {
	*c = append(*c, handler)
}

// Add appends multiple middleware handlers to the middleware chain.
func (c *Chain) Add(handlers ...func(http.Handler) http.Handler) {
	for _, handler := range handlers {
		c.Use(handler)
	}
}

// With creates a new middleware chain from an existing chain, extending it with
// additional middleware.
func (c *Chain) With(handlers ...func(http.Handler) http.Handler) *Chain {
	chain := make(Chain, len(*c))
	copy(chain, *c)
	chain.Add(handlers...)
	return &chain
}

// Handler wraps the provided final handler with all the middleware appended to
// the chain and returns a http.Handler instance.
func (c Chain) Handler(handler http.Handler) http.Handler {
	for i := len(c) - 1; i >= 0; i-- {
		handler = c[i](handler)
	}
	return handler
}

// HandlerFunc wraps the provided final handler function with all the middleware
// appended to the chain and returns a http.Handler instance.
func (c Chain) HandlerFunc(handler http.HandlerFunc) http.Handler {
	return c.Handler(http.HandlerFunc(handler))
}
