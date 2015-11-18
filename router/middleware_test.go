// This file is Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style license.
//
// Modifications by Kevin Stenerson for Reflexion Health Inc. Copyright 2015
package router

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reflexionhealth/vanilla/router/routertest"
)

func TestMiddlewareGeneralCase(t *testing.T) {
	signature := ""
	router := New()
	router.Use(func(c *Context) {
		signature += "A"
		c.Continue()
		signature += "B"
	})
	router.Use(func(c *Context) {
		signature += "C"
		c.Continue()
		signature += "D"
	})
	router.GET("/", func(c *Context) {
		signature += "E"
	})
	router.NotFound(func(c *Context) {
		signature += " X "
	})
	router.NoMethod(func(c *Context) {
		signature += " XX "
	})
	// RUN
	w := routertest.PerformRequest(router, "GET", "/")

	// TEST
	assert.Equal(t, w.Code, 200)
	assert.Equal(t, signature, "ACEDB")
}

func TestMiddlewareNotFound(t *testing.T) {
	signature := ""
	router := New()
	router.Use(func(c *Context) {
		signature += "A"
		c.Continue()
		signature += "B"
	})
	router.Use(func(c *Context) {
		signature += "C"
		c.Continue()
		c.Continue() // we can call Continue (not MustContinue) as much as we want
		c.Continue() // we can call Continue (not MustContinue) as much as we want
		c.Continue() // we can call Continue (not MustContinue) as much as we want
		signature += "D"
	})
	router.NotFound(func(c *Context) {
		signature += "E"
		c.Continue()
		signature += "F"
	}, func(c *Context) {
		signature += "G"
		c.Continue()
		signature += "H"
	})
	router.NoMethod(func(c *Context) {
		signature += " X "
	})
	// RUN
	w := routertest.PerformRequest(router, "GET", "/")

	// TEST
	assert.Equal(t, w.Code, 404)
	assert.Equal(t, signature, "ACEGHFDB")
}

func TestMiddlewareNoMethodEnabled(t *testing.T) {
	signature := ""
	router := New()
	router.Use(func(c *Context) {
		signature += "A"
		c.Continue()
		signature += "B"
	})
	router.Use(func(c *Context) {
		signature += "C"
		c.Continue()
		signature += "D"
	})
	router.NoMethod(func(c *Context) {
		signature += "E"
		c.Continue()
		signature += "F"
	}, func(c *Context) {
		signature += "G"
		c.Continue()
		signature += "H"
	})
	router.NotFound(func(c *Context) {
		signature += " X "
	})
	router.POST("/", func(c *Context) {
		signature += " XX "
	})
	// RUN
	w := routertest.PerformRequest(router, "GET", "/")

	// TEST
	assert.Equal(t, w.Code, 405)
	assert.Equal(t, signature, "ACEGHFDB")
}

func TestMiddlewareWrite(t *testing.T) {
	router := New()
	router.Use(func(c *Context) {
		c.Response.Text(333, "hola\n")
	})
	router.GET("/", func(c *Context) {
		c.Response.JSON(444, map[string]string{"foo": "bar"})
	})

	w := routertest.PerformRequest(router, "GET", "/")

	assert.Equal(t, w.Code, 333)
	assert.Equal(t, w.Body.String(), "hola\n")
}
