package middlewares

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func NewEncryptMiddleware(publicKey *rsa.PublicKey) func(c *resty.Client, r *resty.Request) error {
	return func(c *resty.Client, r *resty.Request) error {
		bodyBytes, ok := r.Body.([]byte)
		if !ok {
			return fmt.Errorf("body is not of type []byte")
		}
		ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, bodyBytes)
		if err != nil {
			return fmt.Errorf("fail encrypt massage; %w", err)
		}

		r.SetBody(ciphertext)
		return nil
	}
}
