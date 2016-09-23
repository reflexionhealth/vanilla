package mock

// This file is Copyright 2014 Jared Morse.  All rights reserved.
// Use of this source code is governed by a MIT style license.
//
// Modifications by Kevin Stenerson for Reflexion Health Inc. Copyright 2016

import (
	"fmt"
	"net/http"
	"strings"
)

// Responses are callbacks that receive and http request and return a mocked response.
type Response func(*http.Request) (*http.Response, error)

// ConnectionFailure is a response that returns a connection failure.  This is
// the default response, and is used when no other response matches.
func ConnectionFailure(req *http.Request) (*http.Response, error) {
	return nil, fmt.Errorf(`failed to connect to "%s"`, req.URL.String())
}

// NewTransport creates a new *Transport with no .Responses.
func NewTransport() *Transport {
	return &Transport{
		Responses: make(map[string]Response),
		Requests:  make(map[string][]*http.Request),
	}
}

// Transport implements http.RoundTripper, which fulfills single http requests
// issued by an http.Client.  This implementation doesn't actually make a
// network request, instead deferring to a registered list of responses.
type Transport struct {
	Responses map[string]Response
	Requests  map[string][]*http.Request

	replaced http.RoundTripper
}

// RoundTrip receives HTTP requests and routes them to the appropriate response.
// It is required to implement the http.RoundTripper interface.  You should not
// use this directly, instead an *http.Client will call it for you.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	url := req.URL.String()

	var key string
	if strings.Contains(url, "?") {
		key = req.Method + " " + strings.Split(url, "?")[0]
	} else {
		key = req.Method + " " + strings.Split(url, "?")[0]
	}
	t.Requests[key] = append(t.Requests[key], req)

	response := t.Responses[key]
	if response != nil {
		return response(req)
	}
	return ConnectionFailure(req)
}

// Register adds a new response associated with a given HTTP method and URL.
// When a request matches, the response will be called to complete the request.
func (t *Transport) Register(method, url string, response Response) {
	t.Responses[method+" "+url] = response
}

// Reset removes all registered Responses and recorded Requests
func (t *Transport) Reset() {
	t.Responses = make(map[string]Response)
	t.Requests = make(map[string][]*http.Request)
}

// Enable replaces net/http's DefaultTransport
func (t *Transport) Enable() {
	t.replaced = http.DefaultTransport
	http.DefaultTransport = t
}

// Disable restores net/http's DefaultTransport
func (t *Transport) Disable() {
	if t.replaced == nil {
		return
	}
	if http.DefaultTransport != t {
		panic("http.DefaultTransport was changed prior to Disable()")
	}

	http.DefaultTransport = t.replaced
	t.replaced = nil
}
