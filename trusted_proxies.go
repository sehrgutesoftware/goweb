package goweb

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"slices"
	"strings"
)

// TrustedProxies are the reverse proxies allowed to report a request's
// originating client via X-Forwarded-For or X-Real-IP.
type TrustedProxies []netip.Prefix

// ParseTrustedProxies parses CIDR prefixes and bare IP addresses, which
// become single-host prefixes. Empty entries are skipped.
func ParseTrustedProxies(in []string) (TrustedProxies, error) {
	out := make(TrustedProxies, 0, len(in))
	for _, raw := range in {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}

		if strings.Contains(s, "/") {
			p, err := netip.ParsePrefix(s)
			if err != nil {
				return nil, fmt.Errorf("parse trusted proxy %q: %w", s, err)
			}
			out = append(out, p)
			continue
		}

		addr, err := netip.ParseAddr(s)
		if err != nil {
			return nil, fmt.Errorf("parse trusted proxy %q: %w", s, err)
		}
		out = append(out, netip.PrefixFrom(addr, addr.BitLen()))
	}

	return out, nil
}

// TrustsPeer reports whether the request's direct peer, the host of
// [http.Request.RemoteAddr], is a trusted proxy.
func (t TrustedProxies) TrustsPeer(r *http.Request) bool {
	return t.contains(peerHost(r.RemoteAddr))
}

// ClientIP resolves the originating client IP of a request.
//
// When the direct peer is not a trusted proxy, its address (the host of
// [http.Request.RemoteAddr]) is returned and forwarded headers are ignored.
//
// When it is, ClientIP walks X-Forwarded-For right-to-left and returns the
// right-most address that is not trusted. If every address in the chain is
// trusted, the left-most is returned. Without X-Forwarded-For, X-Real-IP is
// consulted. Without either, the peer address is returned.
//
// The empty string is returned only when [http.Request.RemoteAddr] is empty
// and no usable forwarded header is present.
func (t TrustedProxies) ClientIP(r *http.Request) string {
	peer := peerHost(r.RemoteAddr)
	if !t.contains(peer) {
		return peer
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for _, part := range slices.Backward(parts) {
			ip := strings.TrimSpace(part)
			if ip != "" && !t.contains(ip) {
				return ip
			}
		}
		for _, part := range parts {
			if ip := strings.TrimSpace(part); ip != "" {
				return ip
			}
		}
	}

	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}

	return peer
}

func (t TrustedProxies) contains(ip string) bool {
	if len(t) == 0 || ip == "" {
		return false
	}
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false
	}

	addr = addr.Unmap()
	for _, p := range t {
		if p.Contains(addr) {
			return true
		}
	}

	return false
}

// peerHost returns the host of a "host:port" remote address. Values
// without a port are returned as-is.
func peerHost(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}

	return host
}
