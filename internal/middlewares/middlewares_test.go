package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/screamsoul/go-metrics-tpl/pkg/ipmask"
	"github.com/stretchr/testify/assert"
)

func TestGzipDecompressMiddleware(t *testing.T) {
	middleware := GzipDecompressMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body) // Use io.ReadAll instead of ioutil.ReadAll
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(body)
		assert.NoError(t, err)
	}))

	tests := []struct {
		name                 string
		contentEncoding      string
		expectedResponseBody string
		encodingFunc         func(body *bytes.Buffer, data string)
	}{
		{
			name:                 "Gzip Encoded",
			contentEncoding:      "gzip",
			expectedResponseBody: "Hello, World!",
			encodingFunc: func(body *bytes.Buffer, data string) {
				var err error
				gw := gzip.NewWriter(body)
				_, err = gw.Write([]byte(data))
				assert.NoError(t, err)
				err = gw.Close()
				assert.NoError(t, err)
			},
		},
		{
			name:                 "Not Gzip Encoded",
			contentEncoding:      "",
			expectedResponseBody: "Hello, World!",
			encodingFunc: func(body *bytes.Buffer, data string) {
				_, err := body.Write([]byte(data))
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer

			tt.encodingFunc(&body, tt.expectedResponseBody)

			req, err := http.NewRequest("POST", "/", &body)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Encoding", tt.contentEncoding)

			// Record the response
			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			// Check the status code
			assert.Equal(t, http.StatusOK, rr.Code)

			// Check the response body
			assert.Equal(t, tt.expectedResponseBody, rr.Body.String())
		})
	}
}

func TestGzipCompressMiddleware(t *testing.T) {
	middleware := GzipCompressMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Hello, World!"))
		assert.NoError(t, err)
	}))

	tests := []struct {
		name                 string
		acceptEncodingHeader string
		contentTypeHeader    string
		expectGzip           bool
	}{
		{
			name:                 "Gzip Supported and Content Type Supported",
			acceptEncodingHeader: "gzip",
			contentTypeHeader:    "application/json",
			expectGzip:           true,
		},
		{
			name:                 "Gzip Supported but Content Type Not Supported",
			acceptEncodingHeader: "gzip",
			contentTypeHeader:    "image/png",
			expectGzip:           false,
		},
		{
			name:                 "Gzip Not Supported",
			acceptEncodingHeader: "",
			contentTypeHeader:    "application/json",
			expectGzip:           false,
		},
		{
			name:                 "Gzip Supported and Content Type Not Specified",
			acceptEncodingHeader: "gzip",
			contentTypeHeader:    "",
			expectGzip:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request
			req, err := http.NewRequest("GET", "/", nil)
			assert.NoError(t, err)

			// Set headers
			req.Header.Set("Accept-Encoding", tt.acceptEncodingHeader)
			req.Header.Set("Content-Type", tt.contentTypeHeader)

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Serve the request
			middleware.ServeHTTP(rr, req)

			// Check if the response is gzipped
			if tt.expectGzip {
				assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"))

				// Decompress the response body
				gr, err := gzip.NewReader(rr.Body)
				assert.NoError(t, err)
				defer func() {
					closeErr := gr.Close()
					assert.NoError(t, closeErr)
				}()

				body, err := io.ReadAll(gr)
				assert.NoError(t, err)
				assert.Equal(t, "Hello, World!", string(body))
			} else {
				assert.Empty(t, rr.Header().Get("Content-Encoding"))
				assert.Equal(t, "Hello, World!", rr.Body.String())
			}
		})
	}
}

func TestNewHashSumHeaderMiddleware(t *testing.T) {
	// Define test cases
	testCase := []struct {
		name           string
		hashKey        string
		requestBody    string
		headerHash     string
		expectedStatus int
	}{
		{
			name:           "Valid hash",
			hashKey:        "testKey",
			requestBody:    "testBody",
			headerHash:     "33899393cccd71ea35f6340d8a70b2e1910d4de0f2c1c5c0befea38b27aecfca", // SHA256("testBodytestKey")
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid hash",
			hashKey:        "testKey",
			requestBody:    "testBody",
			headerHash:     "invalidHash",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing hash key",
			hashKey:        "",
			requestBody:    "testBody",
			headerHash:     "",
			expectedStatus: http.StatusOK, // Assuming the middleware allows the request to proceed if the hash key is missing.
		},
		{
			name:           "Missing header hash",
			hashKey:        "testKey",
			requestBody:    "testBody",
			headerHash:     "",
			expectedStatus: http.StatusOK, // Assuming the middleware allows the request to proceed if the header hash is missing.
		},
	}

	// Run each test case
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			// Create a request with the test case's body and header
			req, err := http.NewRequest("GET", "/", bytes.NewBufferString(tc.requestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("HashSHA256", tc.headerHash)

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Create the middleware with the test case's hash key
			middleware := NewHashSumHeaderMiddleware(tc.hashKey)

			// Wrap the ResponseRecorder in the middleware
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			// Serve the request
			handler.ServeHTTP(rr, req)

			// Assert the expected status code
			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

func TestMiddlewareAllowsRequestWithinCIDR(t *testing.T) {
	cidrip := ipmask.CIDRIP{
		Network: &net.IPNet{
			IP:   net.ParseIP("192.168.1.0"),
			Mask: net.CIDRMask(24, 32),
		},
	}
	middleware := NewTrustedIpMiddleware(cidrip)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "192.168.1.10:12345"

	rr := httptest.NewRecorder()
	handler := middleware(nextHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestMiddlewareHandlesMalformedIP(t *testing.T) {
	cidrip := ipmask.CIDRIP{
		Network: &net.IPNet{
			IP:   net.ParseIP("192.168.1.0"),
			Mask: net.CIDRMask(24, 32),
		},
	}
	middleware := NewTrustedIpMiddleware(cidrip)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "127.0.0.1"

	rr := httptest.NewRecorder()
	handler := middleware(nextHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusForbidden)
	}
}
