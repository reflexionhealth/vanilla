package stack

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reflexionhealth/vanilla/httpserver"
	"github.com/reflexionhealth/vanilla/httpserver/request"
)

func TestLogger(t *testing.T) {
	buffer := new(bytes.Buffer)
	Logger.Global.SetOutput(buffer)

	server := httpserver.New()
	server.Use(LogRequest)

	server.GET("/example", func(c *httpserver.Context) {})
	server.POST("/example", func(c *httpserver.Context) {})
	server.HEAD("/example", func(c *httpserver.Context) {})
	server.OPTIONS("/example", func(c *httpserver.Context) {})

	server.GET("/nomethod", func(c *httpserver.Context) {})

	server.NoMethod(LogRequest, func(c *httpserver.Context) {})
	server.NotFound(LogRequest, func(c *httpserver.Context) {})

	request.PerformRequest(server, "GET", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "GET")
	assert.Contains(t, buffer.String(), "/example")

	request.PerformRequest(server, "POST", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "POST")
	assert.Contains(t, buffer.String(), "/example")

	request.PerformRequest(server, "HEAD", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "HEAD")
	assert.Contains(t, buffer.String(), "/example")

	request.PerformRequest(server, "OPTIONS", "/example")
	assert.Contains(t, buffer.String(), "200")
	assert.Contains(t, buffer.String(), "OPTIONS")
	assert.Contains(t, buffer.String(), "/example")

	request.PerformRequest(server, "PUT", "/nomethod")
	assert.Contains(t, buffer.String(), "405")
	assert.Contains(t, buffer.String(), "PUT")
	assert.Contains(t, buffer.String(), "/nomethod")

	request.PerformRequest(server, "GET", "/notfound")
	assert.Contains(t, buffer.String(), "404")
	assert.Contains(t, buffer.String(), "GET")
	assert.Contains(t, buffer.String(), "/notfound")
}
