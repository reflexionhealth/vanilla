package request

// This file is Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style license.
//
// Modifications by Kevin Stenerson for Reflexion Health Inc. Copyright 2015

import (
	"net/http"
	"net/http/httptest"
)

// New wraps 'http.NewRequest' so it isn't necessary to import 'net/http' everywhere
func New(method, path string) *http.Request {
	req, _ := http.NewRequest(method, path, nil)
	return req
}

func Handle(h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

// Shorthand for Handle(handler, New(method, path))
func Perform(h http.Handler, method, path string) *httptest.ResponseRecorder {
	return Handle(h, New(method, path))
}
