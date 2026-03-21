package compress

import (
	"compress/gzip"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

var supportedContentTypes = []string{"application/json", "text/html", "text/plain"}

func RequestEncoder() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !slices.Contains(supportedContentTypes, c.Request.Header.Get("Content-Type")) {
			c.Next()
			return
		}
		contentEncoding := c.Request.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			zr, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				return
			}
			c.Request.Body = zr
			defer zr.Close()
		}
		c.Next()
		acceptEncoding := c.Request.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			if c.Writer.Status() < 300 {
				c.Writer.Header().Set("Content-Encoding", "gzip")
			}
			zw := gzip.NewWriter(c.Writer)
			defer zw.Close()
		}

	}
}
