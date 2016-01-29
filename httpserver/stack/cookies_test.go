package stack

// This file is Copyright 2015 Matt Silverlock (matt@eatsleeprepeat.net).  All rights reserved.
// Use of this source code is governed by a BSD style license.
//
// Modifications by Kevin Stenerson for Reflexion Health Inc. Copyright 2015

import (
	"bytes"
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reflexionhealth/vanilla/httpserver"
)

var testKey = []byte("abcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabc")

func GetToken(rec *httptest.ResponseRecorder) string {
	res := http.Response{Header: rec.Header()}
	for _, cookie := range res.Cookies() {
		if cookie.Name == CookieXSRFToken {
			return cookie.Value
		}
	}
	return ""
}

// TestProtectCookies checks that ProtectCookies calls sets a cookie and calls continue()
func TestProtectCookies(t *testing.T) {
	server := httpserver.New()
	server.Use(ProtectCookies(testKey))
	server.GET("/", func(c *httpserver.Context) { c.Response.HEAD(300) })

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	assert.Equal(t, rec.Code, 300)
	cookie := rec.Header().Get("Set-Cookie")
	if assert.NotEqual(t, cookie, "") {
		assert.Contains(t, cookie, "HttpOnly")
		assert.Contains(t, cookie, "Secure")
	}
}

// TestMethod checks that idempotent methods return a 200 OK status and that non-idempotent
// methods return a 403 Forbidden status when a CSRF cookie is not present
func TestMethods(t *testing.T) {
	server := httpserver.New()
	server.Use(ProtectCookies(testKey))

	// test idempontent ("safe") methods
	for _, method := range safeMethods {
		server.Handle(method, "/", func(c *httpserver.Context) {})

		req, err := http.NewRequest(method, "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		assert.Equal(t, rec.Code, 200)
		assert.NotEqual(t, rec.Header().Get("Set-Cookie"), "")
	}

	// test non-idempotent methods (should return a 403 without a cookie set)
	nonIdempotent := []string{"POST", "PUT", "DELETE", "PATCH"}
	for _, method := range nonIdempotent {
		server.Handle(method, "/", func(c *httpserver.Context) {})

		req, err := http.NewRequest(method, "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		assert.Equal(t, rec.Code, 403)
		assert.NotEqual(t, rec.Header().Get("Set-Cookie"), "")
	}
}

// TestNoCookie tests for failure if the cookie containing the session does not exist on a POST request
func TestNoCookie(t *testing.T) {
	server := httpserver.New()
	server.Use(ProtectCookies(testKey))
	server.GET("/", func(c *httpserver.Context) {})
	server.POST("/", func(c *httpserver.Context) {})

	// POST the token back in the header
	req, err := http.NewRequest("POST", "http://cookiejar.tst/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	assert.Equal(t, rec.Code, 403)
	assert.Equal(t, rec.Header().Get("Request-Errors"), `["XSRF Token does not match in protected request"]`)
	assert.Equal(t, rec.Body.String(), `{"errors":["XSRF Token does not match in protected request"]}`)
}

// TestBadCookie tests for failure when a cookie header is modified (malformed)
func TestBadCookie(t *testing.T) {
	server := httpserver.New()
	server.Use(ProtectCookies(testKey))
	server.GET("/", func(c *httpserver.Context) {})
	server.POST("/", func(c *httpserver.Context) {})

	// obtain a CSRF cookie via a GET request
	req, err := http.NewRequest("GET", "http://cookiejar.tst/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	tok := GetToken(rec)

	// POST the token back in the header
	req, err = http.NewRequest("POST", "http://cookiejar.tst/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// replace the cookie prefix
	badHeader := strings.Replace(CookieRealToken+"=", rec.Header().Get("Set-Cookie"), "_badCookie", -1)
	req.Header.Set("Cookie", badHeader)
	req.Header.Set("X-XSRF-Token", tok)
	req.Header.Set("Referer", "http://cookiejar.tst/")

	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	assert.Equal(t, rec.Code, 403)
	assert.Equal(t, rec.Header().Get("Request-Errors"), `["XSRF Token does not match in protected request"]`)
	assert.Equal(t, rec.Body.String(), `{"errors":["XSRF Token does not match in protected request"]}`)
}

// TestVaryHeader checks that responses set a "Vary: Cookie" header to prevent client/proxy caching
func TestVaryHeader(t *testing.T) {
	server := httpserver.New()
	server.Use(ProtectCookies(testKey))
	server.HEAD("/", func(c *httpserver.Context) {})

	req, err := http.NewRequest("HEAD", "https://www.golang.org/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	assert.Equal(t, rec.Code, 200)
	assert.Equal(t, rec.Header().Get("Vary"), "Cookie")
}

// TestNoReferer checks that requests with no Referer header fail
func TestNoReferer(t *testing.T) {
	server := httpserver.New()
	server.Use(ProtectCookies(testKey))
	server.POST("/", func(c *httpserver.Context) {})

	req, err := http.NewRequest("POST", "https://golang.org/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	assert.Equal(t, rec.Code, 403)
	assert.Equal(t, rec.Header().Get("Request-Errors"), `["Referer is missing in protected request"]`)
	assert.Equal(t, rec.Body.String(), `{"errors":["Referer is missing in protected request"]}`)
}

// TestBadReferer checks that HTTPS requests with a Referer that do not
// match the request URL corecectly fail CSRF validation
func TestBadReferer(t *testing.T) {
	server := httpserver.New()
	server.Use(ProtectCookies(testKey))
	server.GET("/", func(c *httpserver.Context) {})
	server.POST("/", func(c *httpserver.Context) {})

	// obtain a CSRF cookie via a GET request.
	req, err := http.NewRequest("GET", "https://cookiejar.tst/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	// POST the token back in the header.
	req, err = http.NewRequest("POST", "https://cookiejar.tst/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Cookie", rec.Header().Get("Set-Cookie"))
	req.Header.Set("X-XSRF-Token", GetToken(rec))

	// set a non-matching Referer header.
	req.Header.Set("Referer", "http://golang.org/")

	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	assert.Equal(t, rec.Code, 403)
	assert.Equal(t, rec.Header().Get("Request-Errors"), `["Referer does not match Origin in protected request"]`)
	assert.Equal(t, rec.Body.String(), `{"errors":["Referer does not match Origin in protected request"]}`)
}

// TestWithReferer checks that requests with a valid Referer pass
func TestWithReferer(t *testing.T) {
	server := httpserver.New()
	server.Use(ProtectCookies(testKey))
	server.GET("/", func(c *httpserver.Context) {})
	server.POST("/", func(c *httpserver.Context) {})

	// obtain a CSRF cookie via a GET request.
	req, err := http.NewRequest("GET", "http://cookiejar.tst/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	// POST the token back in the header.
	req, err = http.NewRequest("POST", "http://cookiejar.tst/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Cookie", rec.Header().Get("Set-Cookie"))
	req.Header.Set("X-XSRF-Token", GetToken(rec))
	req.Header.Set("Referer", "http://cookiejar.tst/")

	rec = httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	assert.Equal(t, rec.Code, 200)
	assert.Equal(t, rec.Body.String(), "")
}

// TestMaskUnmaskTokens checks that a token traversing the mask -> unmask process
// is corecectly unmasked to the original 'real' token
func TestMaskUnmaskTokens(t *testing.T) {
	realToken, err := randomBytes(xsrfTokenLength)
	if assert.Nil(t, err) {
		masked := maskToken(realToken)
		if assert.Nil(t, err) {
			unmasked := unmaskToken(masked)
			assert.True(t, sameToken(unmasked, realToken))
		}
	}
}

// TestSameOrigin tests domains that should (or should not) return true for a same-origin check
func TestSameOrigin(t *testing.T) {
	type originTest struct {
		originA  string
		originB  string
		expected bool
	}

	var originTests = []originTest{
		{"https://cookiejar.tst/", "https://cookiejar.tst", true},
		{"http://golang.org/", "http://golang.org/pkg/net/http", true},
		{"https://cookiejar.tst/", "http://cookiejar.tst", false},
		{"https://cookiejar.tst:3333/", "http://cookiejar.tst:4444", false},
	}

	for _, origins := range originTests {
		a, err := url.Parse(origins.originA)
		assert.Nil(t, err)
		b, err := url.Parse(origins.originB)
		assert.Nil(t, err)
		assert.Equal(t, sameOrigin(a, b), origins.expected)
	}
}

func TestXOR(t *testing.T) {
	type tokenTest struct {
		tokenA   []byte
		tokenB   []byte
		expected []byte
	}

	tokenTests := []tokenTest{
		{[]byte("goodbye"), []byte("hello"), []byte{15, 10, 3, 8, 13}},
		{[]byte("gophers"), []byte("clojure"), []byte{4, 3, 31, 2, 16, 0, 22}},
		{nil, []byte("requestToken"), nil},
	}

	for _, tokens := range tokenTests {
		result := xorToken(tokens.tokenA, tokens.tokenB)
		if result != nil {
			assert.Equal(t, bytes.Compare(result, tokens.expected), 0)
		}
	}
}

// shortReader provides a broken implementation of io.Reader for testing.
type shortReader struct{}

func (sr shortReader) Read(p []byte) (int, error) {
	return len(p) % 2, io.ErrUnexpectedEOF
}

// TestGenerateRandomBytes tests the (extremely rare) case that crypto/rand does
// not return the expected number of bytes
func TestGenerateRandomBytes(t *testing.T) {
	original := rand.Reader
	rand.Reader = shortReader{}
	defer func() { rand.Reader = original }()

	_, err := randomBytes(xsrfTokenLength)
	assert.NotNil(t, err)
}
