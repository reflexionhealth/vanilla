package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reflexionhealth/vanilla/httpserver"
	"github.com/reflexionhealth/vanilla/httpserver/request"
)

func TestError(t *testing.T) {
	r := httpserver.New()
	r.GET("/", func(c *httpserver.Context) { Error(&c.Response, 418, "Teapots cannot make coffee") })

	w := request.Perform(r, "GET", "/")
	assert.Equal(t, w.Code, 418)
	assert.Equal(t, w.Header().Get("Request-Errors"), `["Teapots cannot make coffee"]`)
	assert.Equal(t, w.Body.String(), `{"errors":["Teapots cannot make coffee"]}`)
}

func TestUnauthorized(t *testing.T) {
	r := httpserver.New()
	r.GET("/", func(c *httpserver.Context) { Unauthorized(&c.Response) })

	w := request.Perform(r, "GET", "/")
	assert.Equal(t, w.Code, 401)
	assert.Equal(t, w.Header().Get("Request-Errors"), `["Access is denied due to invalid credentials"]`)
	assert.Equal(t, w.Body.String(), `{"errors":["Access is denied due to invalid credentials"]}`)
}

func TestForbidden(t *testing.T) {
	r := httpserver.New()
	r.GET("/", func(c *httpserver.Context) { Forbidden(&c.Response) })

	w := request.Perform(r, "GET", "/")
	assert.Equal(t, w.Code, 403)
	assert.Equal(t, w.Header().Get("Request-Errors"), `["You don't have permissions for this operation"]`)
	assert.Equal(t, w.Body.String(), `{"errors":["You don't have permissions for this operation"]}`)
}

func TestNotFound(t *testing.T) {
	r := httpserver.New()
	r.GET("/", func(c *httpserver.Context) { RouteNotFound(&c.Response) })

	w := request.Perform(r, "GET", "/")
	assert.Equal(t, w.Code, 404)
	assert.Equal(t, w.Header().Get("Request-Errors"), `["No route for requested path"]`)
	assert.Equal(t, w.Body.String(), `{"errors":["No route for requested path"]}`)
}

func TestNoMethod(t *testing.T) {
	r := httpserver.New()
	r.GET("/", func(c *httpserver.Context) { MethodNotSupported(&c.Response) })

	w := request.Perform(r, "GET", "/")
	assert.Equal(t, w.Code, 405)
	assert.Equal(t, w.Header().Get("Request-Errors"), `["Requested path doesn't support that HTTP method"]`)
	assert.Equal(t, w.Body.String(), `{"errors":["Requested path doesn't support that HTTP method"]}`)
}
