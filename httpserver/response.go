package httpserver

// This file is Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style license.
//
// Modifications by Kevin Stenerson for Reflexion Health Inc. Copyright 2015

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
)

const (
	HeaderContentType = "Content-Type"
	ContentTypeHTML   = "text/html; charset=utf-8"
	ContentTypeJSON   = "application/json; charset=utf-8"
	ContentTypeText   = "text/plain; charset=utf-8"
)

type Response struct {
	http.ResponseWriter

	status   int
	rendered bool
}

func (r *Response) Status() int {
	return r.status
}

func (r *Response) Rendered() bool {
	return r.rendered
}

func (r *Response) HTML(status int, html string) (err error) {
	r.Render(status, ContentTypeHTML)
	_, err = io.WriteString(r.ResponseWriter, html)
	return
}

func (r *Response) JSON(status int, obj interface{}) (err error) {
	r.Render(status, ContentTypeJSON)
	switch val := obj.(type) {
	case string:
		_, err = io.WriteString(r.ResponseWriter, val)
	default:
		err = json.NewEncoder(r.ResponseWriter).Encode(obj)
	}
	return
}

func (r *Response) Text(status int, text string) (err error) {
	r.Render(status, ContentTypeText)
	_, err = io.WriteString(r.ResponseWriter, text)
	return
}

func (r *Response) HEAD(status int) {
	r.Render(status, "")
}

func (r *Response) Render(status int, contentType string) {
	if r.rendered {
		panic("render (aka. HTML, JSON, Text) should only be called once")
	}

	if len(contentType) > 0 {
		r.ResponseWriter.Header().Set(HeaderContentType, contentType)
	}

	// NOTE: write header is sets `r.rendered` = true
	r.WriteHeader(status)
}

func (r *Response) Clear(writer http.ResponseWriter) {
	r.ResponseWriter = writer
	r.rendered = false
	r.status = 200
}

// Override http.ResponseWriter's WriteHeader method
func (r *Response) WriteHeader(status int) {
	r.status = status
	r.rendered = true
	r.ResponseWriter.WriteHeader(status)
}

// Implements the http.Hijacker interface
func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.rendered = true
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

// Implements the http.CloseNotify interface
func (r *Response) CloseNotify() <-chan bool {
	return r.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Implements the http.Flush interface
func (r *Response) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
