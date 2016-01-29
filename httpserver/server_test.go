package httpserver

// This file is Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style license.
//
// Modifications by Kevin Stenerson for Reflexion Health Inc. Copyright 2015

import (
	"reflect"
	"testing"

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
	group := server.Group("/users")
	{
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
