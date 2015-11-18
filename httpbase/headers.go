package httpbase

import "github.com/reflexionhealth/vanilla/router"

const (
	HeaderCacheControl       = "Cache-Control"
	HeaderXssProtection      = "X-Xss-Protection"
	HeaderFrameOptions       = "X-Frame-Options"
	HeaderContentTypeOptions = "X-Content-Type-Options"
	HeaderServer             = "Server"

	CacheControlNeverCache = "max-age=0, private, must-revalidate"
)

// CommonHeaders sets our Server-side headers like Cache, Security, etc
func CommonHeaders(serverName string) router.HandlerFunc {
	return func(c *router.Context) {
		header := c.Response.Header()

		// CACHING
		header.Set(HeaderCacheControl, CacheControlNeverCache)

		// SECURITY
		header.Set(HeaderXssProtection, "1; mode=block")
		header.Set(HeaderFrameOptions, "SAMEORIGIN")
		header.Set(HeaderContentTypeOptions, "nosniff")

		// SERVER INFO
		header.Set(HeaderServer, serverName)

		c.Continue()
	}
}
