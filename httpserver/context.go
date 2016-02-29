package httpserver

// This file is copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style license.
//
// Modifications by Kevin Stenerson for Reflexion Health Inc. Copyright 2015

import (
	"net/http"
	"strings"
)

// Context manages the control flow of middleware
type Context struct {
	Request  *http.Request
	Response Response
	Params   Params                 // Params from the path (eg. /thing/:id)
	Locals   map[string]interface{} // Local values set by middleware
	Debug    bool

	handlers         HandlersChain
	nextHandler      HandlerFunc
	nextHandlerIndex int8
}

// ContinueRequest asks the server to call the next handler for this request
// after the current handler function has returned.  This should be preferred
// over PerformRequest().Now() because it keeps the program stack simpler.
func (c *Context) ContinueRequest() {
	if c.nextHandler == nil && c.nextHandlerIndex < int8(len(c.handlers)) {
		c.nextHandler = c.handlers[c.nextHandlerIndex]
		c.nextHandlerIndex += 1
	}
}

// PerformRequest is used to handle the request immediately.
// It can be used instead of ContinueRequest when it must perform logic after
// the normal request handler has run.  If neither method is called, then the
// server does not run any more request handlers.
//
// PerformRequest should be used to initiate the request for the first time;
// if there are no handlers when PerformRequest is called, Go will panic with
// an index out of range runtime error.
func (c *Context) PerformRequest() {
	c.nextHandler = c.handlers[c.nextHandlerIndex]
	c.nextHandlerIndex += 1
	for c.nextHandler != nil {
		handler := c.nextHandler
		c.nextHandler = nil
		handler(c)
	}
}

// Clear resets the context so it can be used by another request
func (c *Context) Clear(res http.ResponseWriter) {
	c.Request = nil
	c.Response.Clear(res)
	c.Params = c.Params[0:0]
	c.Locals = nil

	c.handlers = nil
	c.nextHandler = nil
	c.nextHandlerIndex = 0
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
