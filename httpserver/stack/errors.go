package stack

import (
	"encoding/json"
	"github.com/reflexionhealth/vanilla/httpserver"
)

const (
	unauthorizedMsg = "Access is denied due to invalid credentials"
	forbiddenMsg    = "You don't have permissions for this operation"
	notFoundMsg     = "No route for requested path"
	noMethodMsg     = "Requested path doesn't support that HTTP method"
)

// If HeaderRequestErrors is set, errors will additionally be sent in that header
var HeaderRequestErrors = "Request-Errors"

var (
	unauthorizedHeader = mustMakeErrorHeader(unauthorizedMsg)
	unauthorizedBody   = mustMakeErrorBody(unauthorizedMsg)
	forbiddenHeader    = mustMakeErrorHeader(forbiddenMsg)
	forbiddenBody      = mustMakeErrorBody(forbiddenMsg)
	notFoundHeader     = mustMakeErrorHeader(notFoundMsg)
	notFoundBody       = mustMakeErrorBody(notFoundMsg)
	noMethodHeader     = mustMakeErrorHeader(noMethodMsg)
	noMethodBody       = mustMakeErrorBody(noMethodMsg)
)

type RequestErrors struct {
	Errors []string `json:"errors"`
}

func mustMakeErrorHeader(errmsg string) string {
	headerBytes, err := json.Marshal([]string{errmsg})
	if err != nil {
		panic("unable to make error header")
	}

	return string(headerBytes)
}

func mustMakeErrorBody(errmsg string) string {
	bodyBytes, err := json.Marshal(RequestErrors{[]string{errmsg}})
	if err != nil {
		panic("unable to make error body")
	}

	return string(bodyBytes)
}

// Error sets the header to the value of HeaderRequestErrors (eg. "Request-Errors")
// and renders the errors in a json body
func Error(r *httpserver.Response, status int, errmsg string) {
	if len(HeaderRequestErrors) > 0 {
		header := mustMakeErrorHeader(errmsg)
		r.Header().Set(HeaderRequestErrors, header)
	}

	body := mustMakeErrorBody(errmsg)
	r.JSON(status, body)
}

func StaticError(r *httpserver.Response, status int, header string, body string) {
	if len(HeaderRequestErrors) > 0 {
		r.Header().Set(HeaderRequestErrors, header)
	}

	r.JSON(status, body)
}

func Unauthorized(r *httpserver.Response) {
	StaticError(r, 401, unauthorizedHeader, unauthorizedBody)
}

func Forbidden(r *httpserver.Response) {
	StaticError(r, 403, forbiddenHeader, forbiddenBody)
}

func RouteNotFound(r *httpserver.Response) {
	StaticError(r, 404, notFoundHeader, notFoundBody)
}

func MethodNotSupported(r *httpserver.Response) {
	StaticError(r, 405, noMethodHeader, noMethodBody)
}
