package middlewares

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"
)

func NewDecryptMiddleware(privateKey *rsa.PrivateKey) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if privateKey != nil {
				ciphertext, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Failed to read request body", http.StatusInternalServerError)
					return
				}

				// Decrypt the ciphertext
				plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, ciphertext)
				if err != nil {
					http.Error(w, "Decryption failed", http.StatusBadRequest)
					return
				}

				// Replace r.Body with a new reader reading from plaintext
				r.Body = io.NopCloser(bytes.NewReader(plaintext))
			}

			next.ServeHTTP(w, r)

		})
	}
}
