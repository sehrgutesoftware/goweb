package goweb_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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

func TestItMasksDetailsWhenPassingAGenericError(t *testing.T) {
	e := errors.New("this should be hidden")

	w := httptest.NewRecorder()
	goweb.RespondError(w, nil, e)
	assert.Equal(t, w.Code, goweb.ErrGeneric.StatusCode())
	assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
	assert.JSONEq(t, fmt.Sprintf(`{"code":"%s","detail":null,"message":""}`, goweb.ErrGeneric.ErrorCode()), w.Body.String())
}

func TestItLogsMaskedErrorsWithTheRequestContext(t *testing.T) {
	logs := recordLogs(t)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), ctxKey{}, "req-1"))

	goweb.RespondError(httptest.NewRecorder(), r, errors.New("boom"))

	if assert.Len(t, logs.records, 1) {
		assert.Equal(t, slog.LevelError, logs.records[0].Level)
		assert.Equal(t, "req-1", logs.ctxs[0].Value(ctxKey{}))
		assert.Equal(t, []string{"error"}, attrKeys(logs.records[0]))
	}
}

func TestItDoesNotLogUnmaskedErrors(t *testing.T) {
	logs := recordLogs(t)

	goweb.RespondError(httptest.NewRecorder(), nil, goweb.NewError("test:code", "test message", http.StatusTeapot))

	assert.Empty(t, logs.records)
}

func TestItLogsWriteFailuresAtDebug(t *testing.T) {
	logs := recordLogs(t)
	w := failingWriter{httptest.NewRecorder()}

	err := goweb.Respond(w, nil, map[string]string{"hello": "world"})

	assert.Error(t, err)
	if assert.Len(t, logs.records, 1) {
		assert.Equal(t, slog.LevelDebug, logs.records[0].Level)
		assert.Equal(t, []string{"error"}, attrKeys(logs.records[0]))
	}
}

func TestItLogsEncodingFailuresAtError(t *testing.T) {
	logs := recordLogs(t)
	w := httptest.NewRecorder()

	err := goweb.Respond(w, nil, make(chan int))

	assert.Error(t, err)
	assert.Empty(t, w.Body.String())
	if assert.Len(t, logs.records, 1) {
		assert.Equal(t, slog.LevelError, logs.records[0].Level)
	}
}

type ctxKey struct{}

type failingWriter struct {
	*httptest.ResponseRecorder
}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("broken pipe")
}

type recordingHandler struct {
	records []slog.Record
	ctxs    []context.Context
}

func (h *recordingHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *recordingHandler) WithAttrs([]slog.Attr) slog.Handler       { return h }
func (h *recordingHandler) WithGroup(string) slog.Handler            { return h }

func (h *recordingHandler) Handle(ctx context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	h.ctxs = append(h.ctxs, ctx)
	return nil
}

func recordLogs(t *testing.T) *recordingHandler {
	h := &recordingHandler{}
	prev := slog.Default()
	slog.SetDefault(slog.New(h))
	t.Cleanup(func() { slog.SetDefault(prev) })

	return h
}

func attrKeys(r slog.Record) []string {
	var keys []string
	r.Attrs(func(a slog.Attr) bool {
		keys = append(keys, a.Key)
		return true
	})

	return keys
}
