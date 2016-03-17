package httpserver

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
	"sync/atomic"
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

	File(string, string) RouteHandler
	Directory(string, string) RouteHandler
}

// Server supports configure middleware and routing for a handler
// Create an instance of Server, by using New()
type Server struct {
	RouteGroup
	contextPool sync.Pool
	methodTrees routeTrees

	notFoundHandlers    HandlersChain
	noMethodHandlers    HandlersChain
	unavailableHandlers HandlersChain
	unavailable         int32 // bool used with atomic Load/Store

	DebugEnabled bool
}

// New returns a new blank Server instance without any middleware attached
func New() *Server {
	s := &Server{
		RouteGroup: RouteGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		methodTrees: make(routeTrees, 0, 9),
	}
	s.RouteGroup.server = s
	s.contextPool.New = func() interface{} { return &Context{} }
	return s
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

func (s *Server) addRoute(method, path string, handlers HandlersChain) {
	if path[0] != '/' {
		panic("path must begin with '/'")
	}
	if method == "" {
		panic("HTTP method can not be empty")
	}
	if len(handlers) == 0 {
		panic("there must be at least one handler")
	}

	root := s.methodTrees.get(method)
	if root == nil {
		root = new(node)
		s.methodTrees = append(s.methodTrees, routeTree{
			method: method,
			root:   root,
		})
	}
	root.addRoute(path, handlers)
}

// NotFound registers a handler chain for requests with a path that does not exist
func (s *Server) NotFound(handlers ...HandlerFunc) {
	s.notFoundHandlers = combineHandlers(s.Handlers, handlers)
	s.notFoundHandlers = combineHandlers(s.Handlers, handlers)
}

// NoMethod registers a handler chain for requests with a method that isn't allowed for a path
func (s *Server) NoMethod(handlers ...HandlerFunc) {
	s.noMethodHandlers = combineHandlers(s.Handlers, handlers)
}

// Unavailable registers a handler chain for requests received while the server is marked unavailable
func (s *Server) Unavailable(handlers ...HandlerFunc) {
	s.unavailableHandlers = combineHandlers(s.Handlers, handlers)
}

// IsAvailable returns whether the server is available or not.  If the server is not available,
// the unavailable handler will be called instead of using the normal routing rules.
func (s *Server) IsAvailable() bool {
	return atomic.LoadInt32(&s.unavailable) == 0
}

// SetAvailable toggles whether the server is available or not.  If the server is not available,
// the unavailable handler will be called instead of using the normal routing rules.
func (s *Server) SetAvailable(available bool) {
	if available {
		atomic.StoreInt32(&s.unavailable, 0)
	} else {
		atomic.StoreInt32(&s.unavailable, 1)
	}
}

// Routes returns a slice of registered routes, including some useful information, such as:
// the http method, path and the handler name.
func (s *Server) Routes() (routes []RouteInfo) {
	for _, tree := range s.methodTrees {
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

// Run attaches the server to a http.Server and starts listening and serving HTTP requests.
// It is a shortcut for http.ListenAndServe(addr, server)
func (s *Server) Run(addr ...string) error {
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

	return http.ListenAndServe(address, s)
}

// RunTLS attaches the server to a http.Server and starts listening and serving HTTPS (secure) requests.
// It is a shortcut for http.ListenAndServeTLS(addr, certFile, keyFile, server)
func (s *Server) RunTLS(addr string, certFile string, keyFile string) (err error) {
	err = http.ListenAndServeTLS(addr, certFile, keyFile, s)
	return
}

// RunUnix attaches the server to a http.Server and starts listening and serving HTTP requests
// through the specified unix socket (ie. a file).
func (s *Server) RunUnix(file string) (err error) {
	os.Remove(file)
	listener, err := net.Listen("unix", file)
	if err != nil {
		return
	}
	defer listener.Close()
	err = http.Serve(listener, s)
	return
}

// Conforms to the http.Handler interface.
func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	c := s.contextPool.Get().(*Context)
	c.Clear(res)
	c.Request = req
	c.Debug = s.DebugEnabled

	s.handleHTTPRequest(c)

	s.contextPool.Put(c)
}

func (s *Server) handleHTTPRequest(c *Context) {
	method := c.Request.Method
	path := c.Request.URL.Path

	if !s.IsAvailable() {
		c.handlers = s.unavailableHandlers
		c.Params = c.Params[0:0]
		c.Response.status = 503
		c.PerformRequest()
		if !c.Response.Rendered() {
			c.Response.Text(503, "Service Unavailable")
		}
		return
	}

	// Find root of the tree for the given HTTP method
	for _, tree := range s.methodTrees {
		if tree.method == method {
			var handlers HandlersChain
			handlers, params := tree.root.getValue(path, c.Params)
			if handlers != nil {
				c.handlers = handlers
				c.Params = params
				c.PerformRequest()
				if !c.Response.Rendered() {
					c.Response.HEAD(200)
				}
				return
			}
			break
		}
	}

	// Handle method not allowed
	if len(s.notFoundHandlers) > 0 {
		for _, tree := range s.methodTrees {
			if tree.method != method {
				handlers, _ := tree.root.getValue(path, nil)
				if handlers != nil {
					c.handlers = s.noMethodHandlers
					c.Params = c.Params[0:0]
					c.Response.status = 405
					c.PerformRequest()
					if !c.Response.Rendered() {
						c.Response.Text(405, "No Method")
					}
					return
				}
			}
		}
	}

	if len(s.notFoundHandlers) > 0 {
		c.handlers = s.notFoundHandlers
		c.Params = c.Params[0:0]
		c.Response.status = 404
		c.PerformRequest()
	}

	if !c.Response.Rendered() {
		c.Response.Text(404, "Not Found")
	}
}
