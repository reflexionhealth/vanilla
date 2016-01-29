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
		c.Continue()
		signature += "B"
	})
	server.Use(func(c *Context) {
		signature += "C"
		c.Continue()
		signature += "D"
	})
	server.GET("/", func(c *Context) {
		signature += "E"
	})
	server.NotFound(func(c *Context) {
		signature += " X "
	})
	server.NoMethod(func(c *Context) {
		signature += " XX "
	})
	// RUN
	w := request.PerformRequest(server, "GET", "/")

	// TEST
	assert.Equal(t, w.Code, 200)
	assert.Equal(t, signature, "ACEDB")
}

func TestMiddlewareNotFound(t *testing.T) {
	signature := ""
	server := New()
	server.Use(func(c *Context) {
		signature += "A"
		c.Continue()
		signature += "B"
	})
	server.Use(func(c *Context) {
		signature += "C"
		c.Continue()
		c.Continue() // we can call Continue (not MustContinue) as much as we want
		c.Continue() // we can call Continue (not MustContinue) as much as we want
		c.Continue() // we can call Continue (not MustContinue) as much as we want
		signature += "D"
	})
	server.NotFound(func(c *Context) {
		signature += "E"
		c.Continue()
		signature += "F"
	}, func(c *Context) {
		signature += "G"
		c.Continue()
		signature += "H"
	})
	server.NoMethod(func(c *Context) {
		signature += " X "
	})
	// RUN
	w := request.PerformRequest(server, "GET", "/")

	// TEST
	assert.Equal(t, w.Code, 404)
	assert.Equal(t, signature, "ACEGHFDB")
}

func TestMiddlewareNoMethodEnabled(t *testing.T) {
	signature := ""
	server := New()
	server.Use(func(c *Context) {
		signature += "A"
		c.Continue()
		signature += "B"
	})
	server.Use(func(c *Context) {
		signature += "C"
		c.Continue()
		signature += "D"
	})
	server.NoMethod(func(c *Context) {
		signature += "E"
		c.Continue()
		signature += "F"
	}, func(c *Context) {
		signature += "G"
		c.Continue()
		signature += "H"
	})
	server.NotFound(func(c *Context) {
		signature += " X "
	})
	server.POST("/", func(c *Context) {
		signature += " XX "
	})
	// RUN
	w := request.PerformRequest(server, "GET", "/")

	// TEST
	assert.Equal(t, w.Code, 405)
	assert.Equal(t, signature, "ACEGHFDB")
}

func TestMiddlewareWrite(t *testing.T) {
	server := New()
	server.Use(func(c *Context) {
		c.Response.Text(333, "hola\n")
	})
	server.GET("/", func(c *Context) {
		c.Response.JSON(444, map[string]string{"foo": "bar"})
	})

	w := request.PerformRequest(server, "GET", "/")

	assert.Equal(t, w.Code, 333)
	assert.Equal(t, w.Body.String(), "hola\n")
}
