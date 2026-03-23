package compress

import (
	"compress/gzip"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

var supportedContentTypes = []string{"application/json", "text/html", "text/plain"}

type gzipResponseWriter struct {
	gin.ResponseWriter
	zw *gzip.Writer
}

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
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			c.Request.Body = zr
			defer zr.Close()
		}

		acceptEncoding := c.Request.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			if c.Writer.Status() < 300 {
				c.Writer.Header().Set("Content-Encoding", "gzip")
			}
			zw := gzip.NewWriter(c.Writer)
			c.Writer = &gzipResponseWriter{zw: zw, ResponseWriter: c.Writer}
			defer zw.Close()
		}

		c.Next()
	}
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.zw.Write(b)
}

func (w *gzipResponseWriter) WriteString(s string) (int, error) {
	return w.zw.Write([]byte(s))
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Flush() {
	w.zw.Flush()
	w.ResponseWriter.Flush()
}

func (w *gzipResponseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.CloseNotify()
}
