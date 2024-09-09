package handlers_test

import (
	"fmt"
	"net/http"

	"github.com/screamsoul/go-metrics-tpl/internal/repositories/memory"
	"github.com/screamsoul/go-metrics-tpl/internal/restapi/handlers"
)

// MyResponseWriter implements http.ResponseWriter
type MyResponseWriter struct{}

// Header returns an empty header map since we're not using headers in this example
func (m *MyResponseWriter) Header() http.Header {
	fmt.Println("Header method called")
	return http.Header{}
}

// Write writes the byte slice to the console instead of sending it over HTTP
func (m *MyResponseWriter) Write(p []byte) (int, error) {
	fmt.Println("Write method called:", string(p))
	return len(p), nil
}

// WriteHeader simply prints the status code to the console
func (m *MyResponseWriter) WriteHeader(code int) {
	fmt.Printf("WriteHeader method called with status code: %d\n", code)
}

func Example() {

	// Creating an instance of any repository
	// implementing the repositories.Matrix Storage interface,
	// let's take memory.NewMemStorage as an example.
	mStorage := memory.NewMemStorage()

	// Creating an instance of MetricServer
	var metricServer = handlers.NewMetricServer(
		mStorage,
	)

	// Сreating http request.
	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		panic(err)
	}
	writer := &MyResponseWriter{}

	// Act
	metricServer.PingStorage(writer, req)

	// Output:
	//
}
