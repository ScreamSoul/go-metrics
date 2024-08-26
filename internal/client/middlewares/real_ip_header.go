package middlewares

import "github.com/go-resty/resty/v2"

func NewRealIPHeaderMiddleware(realIP string) func(c *resty.Client, r *resty.Request) error {
	return func(c *resty.Client, r *resty.Request) error {
		r.Header.Set("X-Real-IP", realIP)

		return nil
	}
}
