package middlewares

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDecryptMiddleware(t *testing.T) {
	// Generate RSA keys
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	publicKey := &privateKey.PublicKey

	// Encrypt some plaintext
	plaintext := []byte("test plaintext")
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt plaintext: %v", err)
	}

	// Create a mock request with the ciphertext as body
	req, err := http.NewRequest("GET", "/", io.NopCloser(bytes.NewReader(ciphertext)))
	if err != nil {
		t.Fatalf("Failed to create mock request: %v", err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Instantiate and apply the middleware
	middleware := NewDecryptMiddleware(privateKey)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handler that just echoes the body back
		_, err := io.Copy(w, r.Body)
		assert.NoError(t, err)
	})

	// Pass the request through the middleware
	middleware(handler).ServeHTTP(rr, req)

	// Check if the body has been correctly decrypted and echoed back
	assert.Equal(t, string(plaintext), rr.Body.String())
}
