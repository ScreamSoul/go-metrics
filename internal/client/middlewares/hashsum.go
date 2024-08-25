package middlewares

import (
	"crypto/sha256"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func NewHashSumHeaderMiddleware(hashKey string) func(c *resty.Client, r *resty.Request) error {
	return func(c *resty.Client, r *resty.Request) error {
		bodyBytes, ok := r.Body.([]byte)
		if !ok {
			return fmt.Errorf("body is not of type []byte")
		}
		h := sha256.New()

		h.Write(bodyBytes)
		h.Write([]byte(hashKey))

		dst := h.Sum(nil)

		r.Header.Set("HashSHA256", string(dst))

		return nil
	}
}
