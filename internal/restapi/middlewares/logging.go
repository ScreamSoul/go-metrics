package middlewares

import (
	"net/http"
	"time"

	"github.com/screamsoul/go-metrics-tpl/pkg/logging"
	"go.uber.org/zap"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.GetLogger()

		start := time.Now()

		logger.Info("Request received",
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
		)

		lw := &loggingResponseWriter{ResponseWriter: w}

		next.ServeHTTP(lw, r)

		logger.Info("Response sent",
			zap.Int("status", lw.statusCode),
			zap.Int("size", lw.size),
			zap.Duration("duration", time.Since(start)),
		)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lw.ResponseWriter.Write(b)
	lw.size += size
	return size, err
}
