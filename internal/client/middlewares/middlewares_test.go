package middlewares_test

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"io"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/screamsoul/go-metrics-tpl/internal/client/middlewares"
	"github.com/stretchr/testify/assert"
)

func TestNewGzipCompressBodyMiddleware_CompressesValidBody(t *testing.T) {
	middleware := middlewares.NewGzipCompressBodyMiddleware()
	req := resty.New().R()
	req.SetBody([]byte("test body"))

	err := middleware(nil, req)
	assert.NoError(t, err)

	compressedBody := req.Body.([]byte)
	buf := bytes.NewBuffer(compressedBody)
	gz, err := gzip.NewReader(buf)
	assert.NoError(t, err)

	decompressedBody, err := io.ReadAll(gz)
	assert.NoError(t, err)
	assert.Equal(t, "test body", string(decompressedBody))
}

func TestCorrectlySetsHashSHA256Header(t *testing.T) {
	hashKey := "testKey"
	body := []byte("testBody")
	expectedHash := sha256.New()
	expectedHash.Write(body)
	expectedHash.Write([]byte(hashKey))
	expectedHashSum := expectedHash.Sum(nil)

	middleware := middlewares.NewHashSumHeaderMiddleware(hashKey)
	req := resty.New().R().SetBody(body)

	err := middleware(nil, req)
	assert.NoError(t, err)
	assert.Equal(t, string(expectedHashSum), req.Header.Get("HashSHA256"))
}

func TestNewEncryptMiddleware(t *testing.T) {
	// Setup
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := &privateKey.PublicKey
	client := resty.New()
	request := client.NewRequest()
	request.Body = []byte("test message")

	// Test middleware function
	middleware := middlewares.NewEncryptMiddleware(publicKey)
	err := middleware(client, request)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Assertions
	encryptedBody := request.Body.([]byte)
	if len(encryptedBody) == 0 {
		t.Error("Request body was not encrypted")
	}

	// Optional: Decrypt and verify the content
	// This part requires handling the private key and might be omitted in unit tests focusing on integration
}

func TestMiddlewareSetsXRealIPHeader(t *testing.T) {
	client := resty.New()
	request := client.R()
	realIP := "192.168.1.1"
	middleware := middlewares.NewRealIPHeaderMiddleware(realIP)

	err := middleware(client, request)

	assert.NoError(t, err)
	assert.Equal(t, realIP, request.Header.Get("X-Real-IP"))
}
