package client_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

	"github.com/screamsoul/go-metrics-tpl/internal/client"
	"github.com/screamsoul/go-metrics-tpl/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name           string
		config         client.Config
		expectedServer string
		expectedUpdate string
	}{
		{
			name: "Default Config",
			config: client.Config{
				Server: client.Server{
					ListenServerHost: "localhost:8080",
					CompressRequest:  true,
					BackoffIntervals: []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
					BackoffRetries:   true,
				},
				Client: client.Client{
					ReportInterval: 10,
					PollInterval:   2,
					LogLevel:       "INFO",
				},
			},
			expectedServer: "http://localhost:8080",
			expectedUpdate: "http://localhost:8080/updates/",
		},
		// Add more test cases as needed
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedServer, tc.config.GetServerURL())
			assert.Equal(t, tc.expectedUpdate, tc.config.GetUpdateMetricURL())
		})
	}
}

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

			cfg, err := client.NewConfig()
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedBackoffIntervals, cfg.Server.BackoffIntervals)
			assert.Equal(t, tt.expectedBackoffRetries, cfg.Server.BackoffRetries)

			for k := range tt.envVars {
				t.Setenv(k, "")
			}
		})
	}
}

func TestUnmarshalText__Success(t *testing.T) {
	// Generate a new RSA private key for testing purposes
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	// Marshal the public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)
	pemBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	pemData := pem.EncodeToMemory(pemBlock)

	// Create a temporary file and write the PEM data to it
	tmpFile, err := os.CreateTemp("", "rsa-public-key-*.pem")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove(tmpFile.Name()))
	}()
	_, err = tmpFile.Write(pemData)
	assert.NoError(t, err)
	utils.CloseForse(tmpFile)

	// Test UnmarshalText
	cryptoPublicKey := client.CryptoPublicKey{}
	err = cryptoPublicKey.UnmarshalText([]byte(tmpFile.Name()))
	assert.NoError(t, err)

	// Verify the public key was correctly decoded
	assert.NotNil(t, cryptoPublicKey.Key)
	assert.Equal(t, &privateKey.PublicKey, cryptoPublicKey.Key)
}

func TestUnmarshalText__NotPublicKey(t *testing.T) {
	// Generate a new RSA private key for testing purposes
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	// Marshal the public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)
	pemBlock := &pem.Block{
		Type:  "RSA FAKE KEY",
		Bytes: publicKeyBytes,
	}
	pemData := pem.EncodeToMemory(pemBlock)

	// Create a temporary file and write the PEM data to it
	tmpFile, err := os.CreateTemp("", "rsa-public-key-*.pem")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove(tmpFile.Name()))
	}()
	_, err = tmpFile.Write(pemData)
	assert.NoError(t, err)
	utils.CloseForse(tmpFile)

	// Test UnmarshalText
	cryptoPublicKey := client.CryptoPublicKey{}
	err = cryptoPublicKey.UnmarshalText([]byte(tmpFile.Name()))
	assert.Error(t, err)
}

func TestUnmarshalText__InvalidFile(t *testing.T) {
	cryptoPublicKey := client.CryptoPublicKey{}
	err := cryptoPublicKey.UnmarshalText([]byte("/fake_file"))
	assert.Error(t, err)
}

func TestUnmarshalText__NotParsePublicKey(t *testing.T) {

	pemBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: []byte("fake rsa"),
	}
	pemData := pem.EncodeToMemory(pemBlock)

	// Create a temporary file and write the PEM data to it
	tmpFile, err := os.CreateTemp("", "rsa-public-key-*.pem")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove(tmpFile.Name()))
	}()
	_, err = tmpFile.Write(pemData)
	assert.NoError(t, err)
	utils.CloseForse(tmpFile)

	// Test UnmarshalText
	cryptoPublicKey := client.CryptoPublicKey{}
	err = cryptoPublicKey.UnmarshalText([]byte(tmpFile.Name()))
	assert.Error(t, err)
}

func TestConfigFile(t *testing.T) {
	os.Args = os.Args[:1]

	file, err := os.CreateTemp("", "config_*.json")

	require.NoError(t, err)

	_, err = file.Write([]byte(`
{
	"address": "localhost:1234",
	"report_interval": 1,
	"poll_interval": 1, 
	"crypto_key": "/path/to/key.pem"
}`))
	require.NoError(t, err)

	require.NoError(t, os.Setenv("CONFIG", file.Name()))
	defer func() {
		assert.NoError(t, os.Unsetenv("CONFIG"))
	}()

	cfg, err := client.NewConfig()

	require.NoError(t, err)

	assert.Equal(t, "localhost:1234", cfg.Server.ListenServerHost)

}
