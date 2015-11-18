package router

// This file is Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style license.
//
// Modifications by Kevin Stenerson for Reflexion Health Inc. Copyright 2015

import (
	"net"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"sync"
)

type RouteHandler interface {
	Use(...HandlerFunc) RouteHandler

	Handle(string, string, ...HandlerFunc) RouteHandler
	Any(string, ...HandlerFunc) RouteHandler

	GET(string, ...HandlerFunc) RouteHandler
	POST(string, ...HandlerFunc) RouteHandler
	DELETE(string, ...HandlerFunc) RouteHandler
	PATCH(string, ...HandlerFunc) RouteHandler
	PUT(string, ...HandlerFunc) RouteHandler
	OPTIONS(string, ...HandlerFunc) RouteHandler
	HEAD(string, ...HandlerFunc) RouteHandler
}

// Router supports configure middleware and routing for a handler
// Create an instance of Router, by using New()
type Router struct {
	RouteGroup
	contextPool sync.Pool
	methodTrees routeTrees

	notFoundHandlers HandlersChain
	noMethodHandlers HandlersChain
}

// New returns a new blank Router instance without any middleware attached
func New() *Router {
	r := &Router{
		RouteGroup: RouteGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		methodTrees: make(routeTrees, 0, 9),
	}
	r.RouteGroup.router = r
	r.contextPool.New = func() interface{} { return &Context{} }
	return r
}

type RouteInfo struct {
	Method  string
	Path    string
	Handler string
}

type HandlerFunc func(c *Context)

type HandlersChain []HandlerFunc

// Last returns the last handler in the chain. ie. the last handler is the main one.
func (chain HandlersChain) Last() HandlerFunc {
	length := len(chain)
	if length > 0 {
		return chain[length-1]
	}
	return nil
}

func (r *Router) addRoute(method, path string, handlers HandlersChain) {
	if path[0] != '/' {
		panic("path must begin with '/'")
	}
	if method == "" {
		panic("HTTP method can not be empty")
	}
	if len(handlers) == 0 {
		panic("there must be at least one handler")
	}

	root := r.methodTrees.get(method)
	if root == nil {
		root = new(node)
		r.methodTrees = append(r.methodTrees, routeTree{
			method: method,
			root:   root,
		})
	}
	root.addRoute(path, handlers)
}

// NotFound registers a handler chain for requests with a path that does not exist
func (r *Router) NotFound(handlers ...HandlerFunc) {
	r.notFoundHandlers = combineHandlers(r.Handlers, handlers)
}

// NoMethod registers a handler chain for requests with a method that isn't allowed for a path
func (r *Router) NoMethod(handlers ...HandlerFunc) {
	r.noMethodHandlers = combineHandlers(r.Handlers, handlers)
}

// Routes returns a slice of registered routes, including some useful information, such as:
// the http method, path and the handler name.
func (r *Router) Routes() (routes []RouteInfo) {
	for _, tree := range r.methodTrees {
		routes = iterate("", tree.method, routes, tree.root)
	}
	return routes
}

func iterate(path, method string, routes []RouteInfo, root *node) []RouteInfo {
	path += root.path
	if len(root.handlers) > 0 {
		routes = append(routes, RouteInfo{
			Method:  method,
			Path:    path,
			Handler: runtime.FuncForPC(reflect.ValueOf(root.handlers.Last()).Pointer()).Name(),
		})
	}
	for _, child := range root.children {
		routes = iterate(path, method, routes, child)
	}
	return routes
}

// Run attaches the router to a http.Server and starts listening and serving HTTP requests.
// It is a shortcut for http.ListenAndServe(addr, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (r *Router) Run(addr ...string) error {
	var address string
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); len(port) > 0 {
			address = ":" + port
		} else {
			address = ":8080"
		}
	case 1:
		address = addr[0]
	default:
		panic("too many arguments for resolveAddress")
	}

	return http.ListenAndServe(address, r)
}

// RunTLS attaches the router to a http.Server and starts listening and serving HTTPS (secure) requests.
// It is a shortcut for http.ListenAndServeTLS(addr, certFile, keyFile, router)
// Note: this method will block the calling goroutine undefinitelly unless an error happens.
func (r *Router) RunTLS(addr string, certFile string, keyFile string) (err error) {
	err = http.ListenAndServeTLS(addr, certFile, keyFile, r)
	return
}

// RunUnix attaches the router to a http.Server and starts listening and serving HTTP requests
// through the specified unix socket (ie. a file).
// Note: this method will block the calling goroutine undefinitelly unless an error happens.
func (r *Router) RunUnix(file string) (err error) {
	os.Remove(file)
	listener, err := net.Listen("unix", file)
	if err != nil {
		return
	}
	defer listener.Close()
	err = http.Serve(listener, r)
	return
}

// Conforms to the http.Handler interface.
func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	c := r.contextPool.Get().(*Context)
	c.Clear(res)
	c.Request = req

	r.handleHTTPRequest(c)

	r.contextPool.Put(c)
}

func (r *Router) handleHTTPRequest(c *Context) {
	method := c.Request.Method
	path := c.Request.URL.Path

	// Find root of the tree for the given HTTP method
	for _, tree := range r.methodTrees {
		if tree.method == method {
			var handlers HandlersChain
			handlers, params := tree.root.getValue(path, c.Params)
			if handlers != nil {
				c.handlers = handlers
				c.Params = params
				c.MustContinue() // Execute the handler chain
				if !c.Response.Rendered() {
					c.Response.HEAD(200)
				}
				return
			}
			break
		}
	}

	// Handle method not allowed
	if len(r.notFoundHandlers) > 0 {
		for _, tree := range r.methodTrees {
			if tree.method != method {
				handlers, _ := tree.root.getValue(path, nil)
				if handlers != nil {
					c.handlers = r.noMethodHandlers
					c.Params = c.Params[0:0]
					c.Response.status = 405
					c.Continue() // Execute the handler chain
					if !c.Response.Rendered() {
						c.Response.Text(405, "No Method")
					}
					return
				}
			}
		}
	}

	if len(r.notFoundHandlers) > 0 {
		c.handlers = r.notFoundHandlers
		c.Params = c.Params[0:0]
		c.Response.status = 404
		c.Continue() // Execute the handler chain
	}

	if !c.Response.Rendered() {
		c.Response.Text(404, "Not Found")
	}
}
