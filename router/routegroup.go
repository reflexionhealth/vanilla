package router

// This file is Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style license.
//
// Modifications by Kevin Stenerson for Reflexion Health Inc. Copyright 2015

import (
	"math"
	"path"
	"regexp"
)

// RouteGroup is used internally to configure a router, a RouteGroup is
// associated with a prefix and an array of handlers (middleware)
type RouteGroup struct {
	Handlers HandlersChain
	basePath string
	router   *Router
	root     bool
}

// Use adds middleware to the group, see example code in github.
func (group *RouteGroup) Use(middleware ...HandlerFunc) RouteHandler {
	group.Handlers = append(group.Handlers, middleware...)
	return group.returnObj()
}

// Group creates a new router group. You should add all the routes that have common middlwares or the same path prefix.
// For example, all the routes that use a common middlware for authorization could be grouped.
func (group *RouteGroup) Group(relativePath string, handlers ...HandlerFunc) *RouteGroup {
	return &RouteGroup{
		Handlers: group.appendHandlers(handlers),
		basePath: group.absolutePath(relativePath),
		router:   group.router,
	}
}

func (group *RouteGroup) BasePath() string {
	return group.basePath
}

func (group *RouteGroup) handle(httpMethod, relativePath string, handlers HandlersChain) RouteHandler {
	absolutePath := group.absolutePath(relativePath)
	handlers = group.appendHandlers(handlers)
	group.router.addRoute(httpMethod, absolutePath, handlers)
	return group.returnObj()
}

// Handle registers a new request handle and middleware with the given path and method.
// The last handler should be the real handler, the other ones should be middleware that can and should be shared among different routes.
// See the example code in github.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (group *RouteGroup) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) RouteHandler {
	if matches, err := regexp.MatchString("^[A-Z]+$", httpMethod); !matches || err != nil {
		panic("http method " + httpMethod + " is not valid")
	}
	return group.handle(httpMethod, relativePath, handlers)
}

// Get is a shortcut for router.Handle("GET", path, handle)
func (group *RouteGroup) GET(relativePath string, handlers ...HandlerFunc) RouteHandler {
	return group.handle("GET", relativePath, handlers)
}

// Post is a shortcut for router.Handle("POST", path, handle)
func (group *RouteGroup) POST(relativePath string, handlers ...HandlerFunc) RouteHandler {
	return group.handle("POST", relativePath, handlers)
}

// Delete is a shortcut for router.Handle("DELETE", path, handle)
func (group *RouteGroup) DELETE(relativePath string, handlers ...HandlerFunc) RouteHandler {
	return group.handle("DELETE", relativePath, handlers)
}

// Patch is a shortcut for router.Handle("PATCH", path, handle)
func (group *RouteGroup) PATCH(relativePath string, handlers ...HandlerFunc) RouteHandler {
	return group.handle("PATCH", relativePath, handlers)
}

// Put is a shortcut for router.Handle("PUT", path, handle)
func (group *RouteGroup) PUT(relativePath string, handlers ...HandlerFunc) RouteHandler {
	return group.handle("PUT", relativePath, handlers)
}

// Options is a shortcut for router.Handle("OPTIONS", path, handle)
func (group *RouteGroup) OPTIONS(relativePath string, handlers ...HandlerFunc) RouteHandler {
	return group.handle("OPTIONS", relativePath, handlers)
}

// Head is a shortcut for router.Handle("HEAD", path, handle)
func (group *RouteGroup) HEAD(relativePath string, handlers ...HandlerFunc) RouteHandler {
	return group.handle("HEAD", relativePath, handlers)
}

// Any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE
func (group *RouteGroup) Any(relativePath string, handlers ...HandlerFunc) RouteHandler {
	group.handle("GET", relativePath, handlers)
	group.handle("POST", relativePath, handlers)
	group.handle("PUT", relativePath, handlers)
	group.handle("PATCH", relativePath, handlers)
	group.handle("HEAD", relativePath, handlers)
	group.handle("OPTIONS", relativePath, handlers)
	group.handle("DELETE", relativePath, handlers)
	group.handle("CONNECT", relativePath, handlers)
	group.handle("TRACE", relativePath, handlers)
	return group.returnObj()
}

const maxHandlers int8 = math.MaxInt8 / 2

func combineHandlers(a, b HandlersChain) HandlersChain {
	finalSize := len(a) + len(b)
	if finalSize >= int(maxHandlers) {
		panic("too many handlers")
	}
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, a)
	copy(mergedHandlers[len(a):], b)
	return mergedHandlers
}

func (group *RouteGroup) appendHandlers(handlers HandlersChain) HandlersChain {
	return combineHandlers(group.Handlers, handlers)
}

func lastChar(str string) uint8 {
	return str[len(str)-1]
}

func (group *RouteGroup) absolutePath(relativePath string) string {
	absolutePath := group.basePath
	if len(relativePath) == 0 {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	appendSlash := lastChar(relativePath) == '/' && lastChar(finalPath) != '/'
	if appendSlash {
		return finalPath + "/"
	} else {
		return finalPath
	}
}

func (group *RouteGroup) returnObj() RouteHandler {
	if group.root {
		return group.router
	}
	return group
}
