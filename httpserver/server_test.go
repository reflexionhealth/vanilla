package httpserver

// This file is Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style license.
//
// Modifications by Kevin Stenerson for Reflexion Health Inc. Copyright 2015

import (
	"reflect"
	"testing"

	"github.com/reflexionhealth/vanilla/httpserver/request"
	"github.com/stretchr/testify/assert"
)

func TestCreateEngine(t *testing.T) {
	server := New()
	assert.Equal(t, "/", server.basePath)
	assert.Equal(t, server.server, server)
	assert.Empty(t, server.Handlers)
}

func TestAddRoute(t *testing.T) {
	server := New()
	server.addRoute("GET", "/", HandlersChain{func(_ *Context) {}})

	assert.Len(t, server.methodTrees, 1)
	assert.NotNil(t, server.methodTrees.get("GET"))
	assert.Nil(t, server.methodTrees.get("POST"))

	server.addRoute("POST", "/", HandlersChain{func(_ *Context) {}})

	assert.Len(t, server.methodTrees, 2)
	assert.NotNil(t, server.methodTrees.get("GET"))
	assert.NotNil(t, server.methodTrees.get("POST"))

	server.addRoute("POST", "/post", HandlersChain{func(_ *Context) {}})
	assert.Len(t, server.methodTrees, 2)
}

func TestAddRouteFails(t *testing.T) {
	server := New()
	assert.Panics(t, func() { server.addRoute("", "/", HandlersChain{func(_ *Context) {}}) })
	assert.Panics(t, func() { server.addRoute("GET", "a", HandlersChain{func(_ *Context) {}}) })
	assert.Panics(t, func() { server.addRoute("GET", "/", HandlersChain{}) })

	server.addRoute("POST", "/post", HandlersChain{func(_ *Context) {}})
	assert.Panics(t, func() {
		server.addRoute("POST", "/post", HandlersChain{func(_ *Context) {}})
	})
}

func compareFunc(t *testing.T, a, b interface{}) {
	sf1 := reflect.ValueOf(a)
	sf2 := reflect.ValueOf(b)
	if sf1.Pointer() != sf2.Pointer() {
		t.Error("different functions")
	}
}

func TestListOfRoutes(t *testing.T) {
	server := New()
	server.GET("/", handler_test1)

	{
		group := server.Group("/users")
		group.GET("/", handler_test2)
		group.GET("/:id", handler_test1)
		group.POST("/:id", handler_test2)
	}

	list := server.Routes()

	assert.Len(t, list, 4)
	assert.Contains(t, list, RouteInfo{
		Method:  "GET",
		Path:    "/",
		Handler: "github.com/reflexionhealth/vanilla/httpserver.handler_test1",
	})
	assert.Contains(t, list, RouteInfo{
		Method:  "GET",
		Path:    "/users/",
		Handler: "github.com/reflexionhealth/vanilla/httpserver.handler_test2",
	})
	assert.Contains(t, list, RouteInfo{
		Method:  "GET",
		Path:    "/users/:id",
		Handler: "github.com/reflexionhealth/vanilla/httpserver.handler_test1",
	})
	assert.Contains(t, list, RouteInfo{
		Method:  "POST",
		Path:    "/users/:id",
		Handler: "github.com/reflexionhealth/vanilla/httpserver.handler_test2",
	})
}

func handler_test1(c *Context) {}
func handler_test2(c *Context) {}

func TestAllowedMethods(t *testing.T) {
	server := New()
	server.OPTIONS("/*any", func(c *Context) {})
	server.GET("/", func(c *Context) {})
	server.PUT("/comments", func(c *Context) {})

	{
		group := server.Group("/users")
		group.GET("", handler_test2)
		group.GET("/:id", handler_test1)
		group.POST("/:id", handler_test2)
	}

	var methods []string
	methods = server.AllowedMethods("/")
	assert.Equal(t, []string{"GET", "OPTIONS"}, methods)
	methods = server.AllowedMethods("/comments")
	assert.Equal(t, []string{"PUT", "OPTIONS"}, methods)
	methods = server.AllowedMethods("/users")
	assert.Equal(t, []string{"GET", "OPTIONS"}, methods)
	methods = server.AllowedMethods("/users/:id")
	assert.Equal(t, []string{"GET", "POST", "OPTIONS"}, methods)
	methods = server.AllowedMethods("/unknown")
	assert.Equal(t, []string(nil), methods)
}

func TestHandleOptions(t *testing.T) {
	server := New()
	server.GET("/", func(c *Context) {})
	server.GET("/items", func(c *Context) {})
	server.POST("/items", func(c *Context) {})

	examples := []struct {
		Route  string
		Allow  string
		Status int
	}{
		{"/unknown", "", 404},
		{"/", "GET, OPTIONS", 200},
		{"/items", "GET, POST, OPTIONS", 200},
	}

	for _, ex := range examples {
		req := request.New("OPTIONS", ex.Route)
		res := request.Handle(server, req)
		assert.Equal(t, ex.Status, res.Code, ex.Route)
		assert.Equal(t, ex.Allow, res.Header().Get("Allow"), ex.Route)
	}
}
