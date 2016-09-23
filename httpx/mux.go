// Copyright 2013 Julien Schmidt. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license.
//
// Modifications by Kevin Stenerson for Reflexion Health Inc. Copyright 2016

package httpx

import "net/http"

// Mux is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes.  Mux is based off Julien Schmidt's
// httprouter, but altered to use the "context" package added in Go 1.7, and to
// be more consistent with the builtin net/http Mux's interface.
//
// See https://godoc.org/github.com/julienschmidt/httprouter for more details.
//
// There are a few notable differences:
//
//  - Use httpx.NewMux() instead of httprouter.New()
//  - Use http.Handler or http.HandlerFunc instead of httprouter.Handle
//  - Access the path parameters via a Context with httpx.GetParams(ctx)
//
type Mux struct {
	trees map[string]*node

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handler is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handler can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// If enabled, the router automatically replies to OPTIONS requests.
	// Custom OPTIONS handlers take priority over automatic replies.
	HandleOPTIONS bool

	// Configurable http.Handler which is called when no matching route is
	// found. If it is not set, http.NotFound is used.
	NotFound http.Handler

	// Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	// The "Allow" header with allowed request methods is set before the handler
	// is called.
	MethodNotAllowed http.Handler

	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics.
	PanicHandler func(http.ResponseWriter, *http.Request, interface{})
}

// Make sure the Mux conforms with the http.Handler interface
var _ http.Handler = NewMux()

// New returns a new initialized Mux.
// Path auto-correction, including trailing slashes, is enabled by default.
func NewMux() *Mux {
	return &Mux{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
	}
}

// GET is a shortcut for router.HandleFunc("GET", path, handler)
func (r *Mux) GET(path string, handler http.HandlerFunc) {
	r.HandleFunc("GET", path, handler)
}

// HEAD is a shortcut for router.HandleFunc("HEAD", path, handler)
func (r *Mux) HEAD(path string, handler http.HandlerFunc) {
	r.HandleFunc("HEAD", path, handler)
}

// OPTIONS is a shortcut for router.HandleFunc("OPTIONS", path, handler)
func (r *Mux) OPTIONS(path string, handler http.HandlerFunc) {
	r.HandleFunc("OPTIONS", path, handler)
}

// POST is a shortcut for router.HandleFunc("POST", path, handler)
func (r *Mux) POST(path string, handler http.HandlerFunc) {
	r.HandleFunc("POST", path, handler)
}

// PUT is a shortcut for router.HandleFunc("PUT", path, handler)
func (r *Mux) PUT(path string, handler http.HandlerFunc) {
	r.HandleFunc("PUT", path, handler)
}

// PATCH is a shortcut for router.HandleFunc("PATCH", path, handler)
func (r *Mux) PATCH(path string, handler http.HandlerFunc) {
	r.HandleFunc("PATCH", path, handler)
}

// DELETE is a shortcut for router.HandleFunc("DELETE", path, handler)
func (r *Mux) DELETE(path string, handler http.HandlerFunc) {
	r.HandleFunc("DELETE", path, handler)
}

// Handle registers a new request handler with the given path and method.
func (r *Mux) Handle(method, path string, handler http.Handler) {
	if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root
	}

	root.addRoute(path, handler)
}

// HandleFunc registers a new request handler with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (r *Mux) HandleFunc(method, path string, handler http.HandlerFunc) {
	r.Handle(method, path, http.HandlerFunc(handler))
}

// ServeFiles serves files from the given file system root.
// The path must end with "/*filepath", files are then served from the local
// path /defined/root/dir/*filepath.
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Mux's NotFound handler.
// To use the operating system's file system implementation,
// use http.Dir:
//     router.ServeFiles("/src/*filepath", http.Dir("/var/www"))
func (r *Mux) ServeFiles(path string, root http.FileSystem) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}

	fileServer := http.FileServer(root)

	r.GET(path, func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		params := GetParams(ctx)
		req.URL.Path = params.ByName("filepath")
		fileServer.ServeHTTP(w, req)
	})
}

func (r *Mux) recv(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler(w, req, rcv)
	}
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handler function and the path parameter
// values. Otherwise the third return value indicates whether a redirection to
// the same path with an extra / without the trailing slash should be performed.
func (r *Mux) Lookup(method, path string) (http.Handler, Params, bool) {
	if root := r.trees[method]; root != nil {
		return root.getValue(path)
	}
	return nil, nil, false
}

func (r *Mux) allowed(path, reqMethod string) (allow string) {
	if path == "*" { // server-wide
		for method := range r.trees {
			if method == "OPTIONS" {
				continue
			}

			// add request method to list of allowed methods
			if len(allow) == 0 {
				allow = method
			} else {
				allow += ", " + method
			}
		}
	} else { // specific path
		for method := range r.trees {
			// Skip the requested method - we already tried this one
			if method == reqMethod || method == "OPTIONS" {
				continue
			}

			handler, _, _ := r.trees[method].getValue(path)
			if handler != nil {
				// add request method to list of allowed methods
				if len(allow) == 0 {
					allow = method
				} else {
					allow += ", " + method
				}
			}
		}
	}
	if len(allow) > 0 {
		allow += ", OPTIONS"
	}
	return
}

// ServeHTTP makes the router implement the http.Handler interface.
func (r *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.PanicHandler != nil {
		defer r.recv(w, req)
	}

	path := req.URL.Path

	if root := r.trees[req.Method]; root != nil {
		if handler, ps, tsr := root.getValue(path); handler != nil {
			ctx := ps.Put(req.Context())
			req = req.WithContext(ctx)
			handler.ServeHTTP(w, req)
			return
		} else if req.Method != "CONNECT" && path != "/" {
			code := 301 // Permanent redirect, request with GET method
			if req.Method != "GET" {
				// Temporary redirect, request with same method
				// As of Go 1.3, Go does not support status code 308.
				code = 307
			}

			if tsr && r.RedirectTrailingSlash {
				if len(path) > 1 && path[len(path)-1] == '/' {
					req.URL.Path = path[:len(path)-1]
				} else {
					req.URL.Path = path + "/"
				}
				http.Redirect(w, req, req.URL.String(), code)
				return
			}

			// Try to fix the request path
			if r.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					r.RedirectTrailingSlash,
				)
				if found {
					req.URL.Path = string(fixedPath)
					http.Redirect(w, req, req.URL.String(), code)
					return
				}
			}
		}
	}

	if req.Method == "OPTIONS" {
		// Handle OPTIONS requests
		if r.HandleOPTIONS {
			if allow := r.allowed(path, req.Method); len(allow) > 0 {
				w.Header().Set("Allow", allow)
				return
			}
		}
	} else {
		// Handle 405
		if r.HandleMethodNotAllowed {
			if allow := r.allowed(path, req.Method); len(allow) > 0 {
				w.Header().Set("Allow", allow)
				if r.MethodNotAllowed != nil {
					r.MethodNotAllowed.ServeHTTP(w, req)
				} else {
					http.Error(w,
						http.StatusText(http.StatusMethodNotAllowed),
						http.StatusMethodNotAllowed,
					)
				}
				return
			}
		}
	}

	// Handle 404
	if r.NotFound != nil {
		r.NotFound.ServeHTTP(w, req)
	} else {
		http.NotFound(w, req)
	}
}
