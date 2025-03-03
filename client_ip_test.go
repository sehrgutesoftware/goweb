package goweb_test

import (
	"net/http/httptest"
	"testing"

	"github.com/sehrgutesoftware/goweb"
	"github.com/stretchr/testify/assert"
)

func TestItExtractsTheClientIPFromTheRemoteAddr(t *testing.T) {
}

func TestItExtractsTheClientIPFromProxyHeaders(t *testing.T) {
	// X-Forwarded-For is highest priority
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "42.42.42.42")
	r.Header.Set("X-Real-IP", "21.21.21.21")
	r.RemoteAddr = "0.0.0.0:123"
	ip, err := goweb.ClientIP(r)
	assert.NoError(t, err)
	assert.Equal(t, "42.42.42.42", ip)

	// X-Real-IP is second priority
	r = httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Real-IP", "21.21.21.21")
	r.RemoteAddr = "0.0.0.0:123"
	ip, err = goweb.ClientIP(r)
	assert.NoError(t, err)
	assert.Equal(t, "21.21.21.21", ip)

	// Fall back to RemoteAddr
	r = httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "0.0.0.0:123"
	ip, err = goweb.ClientIP(r)
	assert.NoError(t, err)
	assert.Equal(t, "0.0.0.0", ip)

	// First element in the list is the original client
	r = httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "42.42.42.42,31.31.31.31,20.20.20.20")
	ip, err = goweb.ClientIP(r)
	assert.NoError(t, err)
	assert.Equal(t, "42.42.42.42", ip)
}
