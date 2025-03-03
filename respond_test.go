package goweb_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sehrgutesoftware/goweb"
	"github.com/stretchr/testify/assert"
)

func TestItSendsAJsonResponse(t *testing.T) {
	w := httptest.NewRecorder()
	goweb.Respond(w, nil, map[string]string{"hello": "world"})
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
	assert.JSONEq(t, `{"hello":"world"}`, w.Body.String())
}

func TestItSendsAnErrorResponse(t *testing.T) {
	e := goweb.NewError("test:code", "test message", http.StatusTeapot).Apply("extra info")

	w := httptest.NewRecorder()
	goweb.RespondError(w, nil, e)
	assert.Equal(t, w.Code, http.StatusTeapot)
	assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
	assert.JSONEq(t, `{"code":"test:code","message":"test message","detail":"extra info"}`, w.Body.String())
}

func TestItMasksDetailsWhenSendingAnErrorResponse(t *testing.T) {
	e := goweb.NewMaskedError("test:code", "this should be hidden", http.StatusTeapot).Apply(map[string]string{"password": "secret"})

	w := httptest.NewRecorder()
	goweb.RespondError(w, nil, e)
	assert.Equal(t, w.Code, http.StatusTeapot)
	assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
	assert.JSONEq(t, `{"code":"test:code","detail":null,"message":""}`, w.Body.String())
}
