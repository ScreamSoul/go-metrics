package middlewares

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
)

func NewHashSumHeaderMiddleware(hashKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyHash := r.Header.Get("HashSHA256")

			if bodyHash == "" || hashKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			h := sha256.New()

			buf := make([]byte, 4096) // 4KB buffer

			if _, err := io.CopyBuffer(h, r.Body, buf); err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				return
			}

			h.Write([]byte(hashKey))

			dst := h.Sum(nil)

			if bodyHash != fmt.Sprintf("%x", dst) {
				http.Error(w, "The data is corrupted", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)

		})
	}
}
