package httpserver

// This file is Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style license.
//
// Modifications by Kevin Stenerson for Reflexion Health Inc. Copyright 2015

import (
	"testing"

	"github.com/reflexionhealth/vanilla/httpserver/request"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareGeneralCase(t *testing.T) {
	signature := ""
	server := New()
	server.Use(func(c *Context) {
		signature += "A"
		c.PerformRequest()
		signature += "B"
	})
	server.Use(func(c *Context) {
		signature += "C"
		c.ContinueRequest()
		signature += "D"
	})
	server.Use(func(c *Context) {
		signature += "E"
		c.PerformRequest()
		signature += "F"
	})
	server.GET("/", func(c *Context) {
		signature += "G"
	})
	server.NotFound(func(c *Context) {
		signature += " X "
	})
	server.NoMethod(func(c *Context) {
		signature += " XX "
	})

	// RUN
	w := request.Perform(server, "GET", "/")

	// TEST
	assert.Equal(t, w.Code, 200)
	assert.Equal(t, signature, "ACDEGFB")
}

func TestMiddlewareNotFound(t *testing.T) {
	signature := ""
	server := New()
	server.Use(func(c *Context) {
		signature += "A"
		c.ContinueRequest()
		signature += "B"
	})
	server.Use(func(c *Context) {
		signature += "C"
		c.ContinueRequest()
		c.ContinueRequest() // we can call ContinueRequest (not PerformRequest) as much as we want
		c.ContinueRequest() // we can call ContinueRequest (not PerformRequest) as much as we want
		c.ContinueRequest() // we can call ContinueRequest (not PerformRequest) as much as we want
		signature += "D"
	})
	server.NotFound(func(c *Context) {
		signature += "E"
		c.PerformRequest()
		signature += "F"
	}, func(c *Context) {
		signature += "G"
		c.ContinueRequest()
		signature += "H"
	})
	server.NoMethod(func(c *Context) {
		signature += " X "
	})

	// RUN
	w := request.Perform(server, "GET", "/")

	// TEST
	assert.Equal(t, w.Code, 404)
	assert.Equal(t, signature, "ABCDEGHF")
}

func TestMiddlewareNoMethod(t *testing.T) {
	signature := ""
	server := New()
	server.Use(func(c *Context) {
		signature += "A"
		c.ContinueRequest()
		signature += "B"
	})
	server.Use(func(c *Context) {
		signature += "C"
		c.ContinueRequest()
		signature += "D"
	})
	server.NoMethod(func(c *Context) {
		signature += "E"
		c.PerformRequest()
		signature += "F"
	}, func(c *Context) {
		signature += "G"
		c.ContinueRequest()
		signature += "H"
	})
	server.NotFound(func(c *Context) {
		signature += " X "
	})
	server.POST("/", func(c *Context) {
		signature += " XX "
	})

	// RUN
	w := request.Perform(server, "GET", "/")

	// TEST
	assert.Equal(t, w.Code, 405)
	assert.Equal(t, signature, "ABCDEGHF")
}

func TestMiddlewareUnavailable(t *testing.T) {
	var signature string
	server := New()
	server.Use(func(c *Context) {
		signature += "A"
		c.ContinueRequest()
		signature += "B"
	})
	server.Use(func(c *Context) {
		signature += "C"
		c.ContinueRequest()
		signature += "D"
	})
	server.Unavailable(func(c *Context) {
		signature += "E("
		c.PerformRequest()
		signature += ")F"
	}, func(c *Context) {
		signature += "G"
	})
	server.NoMethod(func(c *Context) { signature += " X " })
	server.NotFound(func(c *Context) { signature += " Y " })
	server.GET("/", func(c *Context) { signature += "(Z)" })

	// initially available
	signature = ""
	w := request.Perform(server, "GET", "/")
	assert.Equal(t, w.Code, 200)
	assert.Equal(t, signature, "ABCD(Z)")

	// made unavailable
	signature = ""
	server.SetAvailable(false)
	w = request.Perform(server, "GET", "/")
	assert.Equal(t, w.Code, 503)
	assert.Equal(t, signature, "ABCDE(G)F")

	// made available
	signature = ""
	server.SetAvailable(true)
	w = request.Perform(server, "GET", "/")
	assert.Equal(t, w.Code, 200)
	assert.Equal(t, signature, "ABCD(Z)")
}

func TestMiddlewareWrite(t *testing.T) {
	server := New()
	server.Use(func(c *Context) {
		c.Response.Text(333, "hola\n")
	})
	server.GET("/", func(c *Context) {
		c.Response.JSON(444, map[string]string{"foo": "bar"})
	})

	w := request.Perform(server, "GET", "/")

	assert.Equal(t, w.Code, 333)
	assert.Equal(t, w.Body.String(), "hola\n")
}
