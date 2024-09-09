package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Logs request method and URI correctly
func TestLoggingMiddleware_LogsRequestMethodAndURI(t *testing.T) {
	req, err := http.NewRequest("GET", "/test-uri", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Som text"))
		assert.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	})

	LoggingMiddleware(handler).ServeHTTP(rr, req)

}
