// This file is Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style license.
//
// Modifications by Kevin Stenerson for Reflexion Health Inc. Copyright 2015
package router

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateEngine(t *testing.T) {
	router := New()
	assert.Equal(t, "/", router.basePath)
	assert.Equal(t, router.router, router)
	assert.Empty(t, router.Handlers)
}

func TestAddRoute(t *testing.T) {
	router := New()
	router.addRoute("GET", "/", HandlersChain{func(_ *Context) {}})

	assert.Len(t, router.methodTrees, 1)
	assert.NotNil(t, router.methodTrees.get("GET"))
	assert.Nil(t, router.methodTrees.get("POST"))

	router.addRoute("POST", "/", HandlersChain{func(_ *Context) {}})

	assert.Len(t, router.methodTrees, 2)
	assert.NotNil(t, router.methodTrees.get("GET"))
	assert.NotNil(t, router.methodTrees.get("POST"))

	router.addRoute("POST", "/post", HandlersChain{func(_ *Context) {}})
	assert.Len(t, router.methodTrees, 2)
}

func TestAddRouteFails(t *testing.T) {
	router := New()
	assert.Panics(t, func() { router.addRoute("", "/", HandlersChain{func(_ *Context) {}}) })
	assert.Panics(t, func() { router.addRoute("GET", "a", HandlersChain{func(_ *Context) {}}) })
	assert.Panics(t, func() { router.addRoute("GET", "/", HandlersChain{}) })

	router.addRoute("POST", "/post", HandlersChain{func(_ *Context) {}})
	assert.Panics(t, func() {
		router.addRoute("POST", "/post", HandlersChain{func(_ *Context) {}})
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
	router := New()
	router.GET("/", handler_test1)
	group := router.Group("/users")
	{
		group.GET("/", handler_test2)
		group.GET("/:id", handler_test1)
		group.POST("/:id", handler_test2)
	}

	list := router.Routes()

	assert.Len(t, list, 4)
	assert.Contains(t, list, RouteInfo{
		Method:  "GET",
		Path:    "/",
		Handler: "github.com/reflexionhealth/vanilla/router.handler_test1",
	})
	assert.Contains(t, list, RouteInfo{
		Method:  "GET",
		Path:    "/users/",
		Handler: "github.com/reflexionhealth/vanilla/router.handler_test2",
	})
	assert.Contains(t, list, RouteInfo{
		Method:  "GET",
		Path:    "/users/:id",
		Handler: "github.com/reflexionhealth/vanilla/router.handler_test1",
	})
	assert.Contains(t, list, RouteInfo{
		Method:  "POST",
		Path:    "/users/:id",
		Handler: "github.com/reflexionhealth/vanilla/router.handler_test2",
	})
}

func handler_test1(c *Context) {}
func handler_test2(c *Context) {}
