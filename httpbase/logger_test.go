package httpbase

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reflexionhealth/vanilla/router"
	"github.com/reflexionhealth/vanilla/router/routertest"
)

func TestLogger(t *testing.T) {
	buffer := new(bytes.Buffer)
	Logger.Global.SetOutput(buffer)

	server := router.New()
	server.Use(LogRequest)

	server.GET("/example", func(c *router.Context) {})
	server.POST("/example", func(c *router.Context) {})
	server.HEAD("/example", func(c *router.Context) {})
	server.OPTIONS("/example", func(c *router.Context) {})

	server.GET("/nomethod", func(c *router.Context) {})

	server.NoMethod(LogRequest, func(c *router.Context) {})
	server.NotFound(LogRequest, func(c *router.Context) {})

	routertest.PerformRequest(server, "GET", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "GET")
	assert.Contains(t, buffer.String(), "/example")

	routertest.PerformRequest(server, "POST", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "POST")
	assert.Contains(t, buffer.String(), "/example")

	routertest.PerformRequest(server, "HEAD", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "HEAD")
	assert.Contains(t, buffer.String(), "/example")

	routertest.PerformRequest(server, "OPTIONS", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "OPTIONS")
	assert.Contains(t, buffer.String(), "/example")

	routertest.PerformRequest(server, "PUT", "/nomethod")
	assert.Contains(t, buffer.String(), "405")
	assert.Contains(t, buffer.String(), "PUT")
	assert.Contains(t, buffer.String(), "/nomethod")

	routertest.PerformRequest(server, "GET", "/notfound")
	assert.Contains(t, buffer.String(), "404")
	assert.Contains(t, buffer.String(), "GET")
	assert.Contains(t, buffer.String(), "/notfound")
}
