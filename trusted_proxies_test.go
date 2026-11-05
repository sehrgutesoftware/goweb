package goweb_test

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/sehrgutesoftware/goweb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItParsesTrustedProxies(t *testing.T) {
	got, err := goweb.ParseTrustedProxies([]string{"10.0.0.0/8", " 2001:db8::/32 ", "10.1.2.3", "2001:db8::1", "", "  "})

	require.NoError(t, err)
	assert.Equal(t, goweb.TrustedProxies{
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("2001:db8::/32"),
		netip.MustParsePrefix("10.1.2.3/32"),
		netip.MustParsePrefix("2001:db8::1/128"),
	}, got)
}

func TestItParsesNoTrustedProxiesFromEmptyInput(t *testing.T) {
	got, err := goweb.ParseTrustedProxies(nil)

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestItRejectsInvalidTrustedProxies(t *testing.T) {
	for _, in := range []string{"not-a-cidr/24", "not-an-ip", "10.0.0.0/33"} {
		t.Run(in, func(t *testing.T) {
			got, err := goweb.ParseTrustedProxies([]string{"10.0.0.0/8", in})

			assert.ErrorContains(t, err, in)
			assert.Nil(t, got)
		})
	}
}

func TestItTellsWhetherThePeerIsATrustedProxy(t *testing.T) {
	trusted := goweb.TrustedProxies{netip.MustParsePrefix("10.0.0.0/8")}

	tests := []struct {
		name       string
		trusted    goweb.TrustedProxies
		remoteAddr string
		want       bool
	}{
		{"it trusts a peer within a prefix", trusted, "10.0.0.1:1234", true},
		{"it trusts a peer without a port", trusted, "10.0.0.1", true},
		{"it trusts an IPv4-mapped IPv6 peer", trusted, "[::ffff:10.0.0.1]:1234", true},
		{"it distrusts a peer outside every prefix", trusted, "203.0.113.5:1234", false},
		{"it distrusts every peer when no proxies are trusted", nil, "10.0.0.1:1234", false},
		{"it distrusts an empty remote address", trusted, "", false},
		{"it distrusts a malformed remote address", trusted, "garbage", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tt.remoteAddr

			assert.Equal(t, tt.want, tt.trusted.TrustsPeer(r))
		})
	}
}

func TestItResolvesTheClientIP(t *testing.T) {
	trusted := goweb.TrustedProxies{
		netip.MustParsePrefix("127.0.0.1/32"),
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("::1/128"),
	}

	tests := []struct {
		name       string
		trusted    goweb.TrustedProxies
		remoteAddr string
		xff        string
		xri        string
		want       string
	}{
		{
			name:       "it returns the peer when no proxies are trusted",
			remoteAddr: "203.0.113.5:1234",
			xff:        "1.2.3.4",
			want:       "203.0.113.5",
		},
		{
			name:       "it ignores forwarded headers from an untrusted peer",
			trusted:    trusted,
			remoteAddr: "203.0.113.5:1234",
			xff:        "1.2.3.4",
			xri:        "5.6.7.8",
			want:       "203.0.113.5",
		},
		{
			name:       "it returns a trusted peer without forwarded headers",
			trusted:    trusted,
			remoteAddr: "10.0.0.1:1234",
			want:       "10.0.0.1",
		},
		{
			name:       "it returns the forwarded address from a trusted peer",
			trusted:    trusted,
			remoteAddr: "10.0.0.1:1234",
			xff:        "203.0.113.5",
			want:       "203.0.113.5",
		},
		{
			name:       "it walks X-Forwarded-For right-to-left skipping trusted hops",
			trusted:    trusted,
			remoteAddr: "10.0.0.1:1234",
			xff:        "198.51.100.9, 203.0.113.5, 10.0.0.2, 10.0.0.3",
			want:       "203.0.113.5",
		},
		{
			name:       "it returns the left-most entry when every hop is trusted",
			trusted:    trusted,
			remoteAddr: "10.0.0.1:1234",
			xff:        "10.0.0.4, 10.0.0.2, 10.0.0.3",
			want:       "10.0.0.4",
		},
		{
			name:       "it falls back to X-Real-IP without X-Forwarded-For",
			trusted:    trusted,
			remoteAddr: "127.0.0.1:1234",
			xri:        "203.0.113.7",
			want:       "203.0.113.7",
		},
		{
			name:       "it prefers X-Forwarded-For over X-Real-IP",
			trusted:    trusted,
			remoteAddr: "127.0.0.1:1234",
			xff:        "203.0.113.5",
			xri:        "9.9.9.9",
			want:       "203.0.113.5",
		},
		{
			name:       "it handles IPv6 peers",
			trusted:    trusted,
			remoteAddr: "[::1]:1234",
			xff:        "2001:db8::1",
			want:       "2001:db8::1",
		},
		{
			name:       "it returns the peer when X-Forwarded-For is blank",
			trusted:    trusted,
			remoteAddr: "10.0.0.1:1234",
			xff:        " , ,  ",
			want:       "10.0.0.1",
		},
		{
			name:       "it handles a remote address without a port",
			trusted:    trusted,
			remoteAddr: "10.0.0.1",
			xff:        "203.0.113.5",
			want:       "203.0.113.5",
		},
		{
			name:       "it returns empty without a remote address or headers",
			trusted:    trusted,
			remoteAddr: "",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				r.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				r.Header.Set("X-Real-IP", tt.xri)
			}

			assert.Equal(t, tt.want, tt.trusted.ClientIP(r))
		})
	}
}
