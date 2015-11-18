package httpbase

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reflexionhealth/vanilla/router"
)

// TestCommonHeaders checks the headers set by CommonHeaders
func TestCommonHeaders(t *testing.T) {
	server := router.New()
	server.Use(CommonHeaders("Testify"))
	server.GET("/", func(c *router.Context) { c.Response.HEAD(300) })

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	assert.Equal(t, rec.Code, 300)
	assert.Equal(t, rec.Header().Get("Server"), "Testify")
	assert.NotEqual(t, rec.Header().Get("Cache-Control"), "")
	assert.NotEqual(t, rec.Header().Get("X-Xss-Protection"), "")
	assert.NotEqual(t, rec.Header().Get("X-Frame-Options"), "")
	assert.NotEqual(t, rec.Header().Get("X-Content-Type-Options"), "")
}
