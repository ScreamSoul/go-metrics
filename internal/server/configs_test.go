package server_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

	"github.com/screamsoul/go-metrics-tpl/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackoffIntervalConfig(t *testing.T) {
	testTable := []struct {
		name                     string
		envVars                  map[string]string
		expectedBackoffIntervals []time.Duration
		expectedBackoffRetries   bool
	}{
		{
			name:                     "Default configuration",
			envVars:                  map[string]string{},
			expectedBackoffIntervals: []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
			expectedBackoffRetries:   true,
		},
		{
			name: "Custom backoff intervals",
			envVars: map[string]string{
				"BACKOFF_INTERVALS": "1s,2s,3s",
			},
			expectedBackoffIntervals: []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second},
			expectedBackoffRetries:   true,
		},
		{
			name: "Disable backoff retries",
			envVars: map[string]string{
				"BACKOFF_RETRIES": "false",
			},
			expectedBackoffIntervals: nil,
			expectedBackoffRetries:   false,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = os.Args[:1]
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			cfg, err := server.NewConfig()
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedBackoffIntervals, cfg.Postgres.BackoffIntervals)
			assert.Equal(t, tt.expectedBackoffRetries, cfg.Postgres.BackoffRetries)

			for k := range tt.envVars {
				t.Setenv(k, "")
			}
		})
	}
}

func TestUnmarshalText(t *testing.T) {
	// Generate a new RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	// Encode the private key to PEM format
	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})

	// Create a temporary file to write the PEM-encoded private key
	tmpfile, err := os.CreateTemp("", "testkey*.pem")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove(tmpfile.Name())) // Clean up
	}()

	_, err = tmpfile.Write(privPEM)
	assert.NoError(t, err)
	err = tmpfile.Close()
	assert.NoError(t, err)

	// Instantiate CryptoPublicKey and call UnmarshalText
	cryptoPubKey := &server.CryptoPublicKey{}
	err = cryptoPubKey.UnmarshalText([]byte(tmpfile.Name()))
	require.NoError(t, err)

	// Verify the parsed key matches the original
	assert.Equal(t, privateKey.N, cryptoPubKey.Key.N)
}

func TestUnmarshalText__IncorrectFile(t *testing.T) {
	// Instantiate CryptoPublicKey and call UnmarshalText
	cryptoPubKey := &server.CryptoPublicKey{}
	err := cryptoPubKey.UnmarshalText([]byte("/fake_file"))
	require.Error(t, err)
}

func TestUnmarshalText__FailKeyHead(t *testing.T) {
	// Generate a new RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	// Encode the private key to PEM format
	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA FAKE HEAD",
		Bytes: privBytes,
	})

	// Create a temporary file to write the PEM-encoded private key
	tmpfile, err := os.CreateTemp("", "testkey*.pem")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove(tmpfile.Name())) // Clean up
	}()

	_, err = tmpfile.Write(privPEM)
	assert.NoError(t, err)
	err = tmpfile.Close()
	assert.NoError(t, err)

	// Instantiate CryptoPublicKey and call UnmarshalText
	cryptoPubKey := &server.CryptoPublicKey{}
	err = cryptoPubKey.UnmarshalText([]byte(tmpfile.Name()))
	require.Error(t, err)

}

func TestUnmarshalText__FailParseKey(t *testing.T) {

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: []byte("fake key"),
	})

	// Create a temporary file to write the PEM-encoded private key
	tmpfile, err := os.CreateTemp("", "testkey*.pem")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove(tmpfile.Name())) // Clean up
	}()

	_, err = tmpfile.Write(privPEM)
	assert.NoError(t, err)
	err = tmpfile.Close()
	assert.NoError(t, err)

	// Instantiate CryptoPublicKey and call UnmarshalText
	cryptoPubKey := &server.CryptoPublicKey{}
	err = cryptoPubKey.UnmarshalText([]byte(tmpfile.Name()))
	require.Error(t, err)
}
