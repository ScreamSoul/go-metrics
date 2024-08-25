package middlewares

import (
	"bytes"
	"compress/gzip"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func NewGzipCompressBodyMiddleware() func(c *resty.Client, r *resty.Request) error {
	return func(c *resty.Client, r *resty.Request) error {
		// Checking if there is already a Content-Encoding header
		if r.Header.Get("Content-Encoding") != "" {
			return nil
		}

		bodyBytes, ok := r.Body.([]byte)
		if !ok {
			return fmt.Errorf("body is not of type []byte")
		}

		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write(bodyBytes); err != nil {
			return err
		}
		if err := gz.Close(); err != nil {
			return err
		}

		r.Body = buf.Bytes()

		r.Header.Set("Content-Encoding", "gzip")

		return nil
	}

}
