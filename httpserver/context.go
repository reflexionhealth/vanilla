package httpserver

// This file is copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style license.
//
// Modifications by Kevin Stenerson for Reflexion Health Inc. Copyright 2015

import (
	"fmt"
	"net/http"
	"strings"
)

// Context manages the control flow of middleware
type Context struct {
	Request  *http.Request
	Response Response
	Params   Params                 // Params from the path (eg. /thing/:id)
	Locals   map[string]interface{} // Local values set by middleware

	handlers     HandlersChain
	handlerIndex int8
}

// Continue should be used only inside middleware.
// It executes the next Handler in the chain.
func (c *Context) Continue() {
	c.handlerIndex++
	if c.handlerIndex < int8(len(c.handlers)) {
		c.handlers[c.handlerIndex](c)
	} else {
		fmt.Errorf("Called Continue() but there are no more handlers to execute")
	}
}

// MustContinue should be used only inside middleware.
// It is like Continue, but it skips bounds checking
func (c *Context) MustContinue() {
	c.handlerIndex++
	c.handlers[c.handlerIndex](c)
}

// Clear resets the context so it can be used by another request
func (c *Context) Clear(res http.ResponseWriter) {
	c.Request = nil
	c.Response.Clear(res)
	c.Params = c.Params[0:0]
	c.Locals = nil

	c.handlers = nil
	c.handlerIndex = -1
}

// ClientIP implements a best effort algorithm to return the real client IP, it parses
// X-Real-IP and X-Forwarded-For in order to work properly with reverse-proxies such us: nginx or haproxy.
func (c *Context) ClientIP() string {
	// Try X-Real-Ip
	clientIP := strings.TrimSpace(c.Request.Header.Get("X-Real-Ip"))
	if len(clientIP) > 0 {
		return clientIP
	}

	// Try X-Forwarded-For
	clientIP = c.Request.Header.Get("X-Forwarded-For")
	if index := strings.IndexByte(clientIP, ','); index >= 0 {
		clientIP = clientIP[0:index]
	}
	clientIP = strings.TrimSpace(clientIP)
	if len(clientIP) > 0 {
		return clientIP
	}

	// Try net.http.Request.RemoteAddr
	return strings.TrimSpace(c.Request.RemoteAddr)
}

// SetLocal is used to store a new key/value pair exclusively for this context.
// It also lazy initializes c.Locals if it was not used previously.
func (c *Context) SetLocal(key string, value interface{}) {
	if c.Locals == nil {
		c.Locals = make(map[string]interface{})
	}
	c.Locals[key] = value
}

// GetLocal returns the value for the given key
func (c *Context) GetLocal(key string) (value interface{}, exists bool) {
	if c.Locals != nil {
		value, exists = c.Locals[key]
		return
	}
	return nil, false
}

// Returns the value for the given key if it exists, otherwise it panics.
func (c *Context) MustGetLocal(key string) interface{} {
	if value, exists := c.GetLocal(key); exists {
		return value
	}
	panic("Local \"" + key + "\" does not exist")
}
