package middlewares

import (
	"net/http"
	"strings"

	"github.com/screamsoul/go-metrics-tpl/pkg/ipmask"
)

// getIPAddress extracts the client's IP address
func getIPAddress(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")

	if ip == "" || ip == "unknown" {
		ip = r.RemoteAddr
	}

	// Remove port number if present
	if strings.Contains(ip, ":") {
		ip = ip[:strings.Index(ip, ":")]
	}

	return ip
}

func NewTrustedIPMiddleware(cidrip ipmask.CIDRIP) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cidrip.Network != nil {
				ipAddress := getIPAddress(r)

				if !cidrip.CheckIPIncluded(ipAddress) {
					http.Error(w, "", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
