package goweb

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP returns the original client IP address of the request using headers
// added by reverse proxies. It falls back to the request's RemoteAddr if no
// relevant headers are present in the request.
func ClientIP(r *http.Request) (string, error) {
	for _, header := range []string{
		"X-Forwarded-For",
		"X-Real-IP",
	} {
		// In a chain of reverse proxies, each proxy should append its client's
		// IP address to the header. The first address in this comma-separated
		// list is the original client IP.
		parts := strings.Split(r.Header.Get(header), ",")
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip, nil
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	return host, err
}
