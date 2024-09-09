package middlewares

import (
	"compress/gzip"
	"net/http"
	"slices"
	"strings"

	"github.com/screamsoul/go-metrics-tpl/pkg/utils"
)

type gzipResponseWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *gzipResponseWriter {
	return &gzipResponseWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (grw *gzipResponseWriter) Write(p []byte) (int, error) {
	return grw.zw.Write(p)
}
func (grw *gzipResponseWriter) Header() http.Header {
	return grw.w.Header()
}

func (grw *gzipResponseWriter) WriteHeader(statusCode int) {
	grw.w.WriteHeader(statusCode)
}

func (grw *gzipResponseWriter) Close() error {
	return grw.zw.Close()
}

func GzipCompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w
		supportsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
		gzipContentType := slices.Contains(
			[]string{"application/json", "text/html", ""},
			r.Header.Get("Content-Type"),
		)
		if supportsGzip && gzipContentType {

			gw := newCompressWriter(w)
			ow = gw
			gw.w.Header().Set("Content-Encoding", "gzip")

			defer utils.CloseForse(gw)
		}

		next.ServeHTTP(ow, r)

	})
}
