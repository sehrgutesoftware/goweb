package route_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sehrgutesoftware/goweb/route"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItSetsThePatternOfTheMatchedRoute(t *testing.T) {
	var got string
	router := build(t, route.Group("/api", []*route.Route{
		route.Func(http.MethodGet, "/users/:id", func(w http.ResponseWriter, r *http.Request) {
			got = r.Pattern
		}),
	}))

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/users/42", nil))

	assert.Equal(t, "GET /api/users/:id", got)
}

func TestItSetsThePatternBeforeMiddlewareRuns(t *testing.T) {
	var got string
	mw := route.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got = r.Pattern
			next.ServeHTTP(w, r)
		})
	})
	router := build(t, route.Group("/", []*route.Route{
		route.Func(http.MethodPost, "/files/*name", func(http.ResponseWriter, *http.Request) {}),
	}).Middleware(mw))

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/files/a/b.txt", nil))

	assert.Equal(t, "POST /files/*name", got)
}

func TestItLeavesThePatternEmptyForUnmatchedRequests(t *testing.T) {
	got := "unset"
	router, err := route.Group("/", []*route.Route{
		route.Func(http.MethodGet, "/users/:id", func(http.ResponseWriter, *http.Request) {}),
	}).Build()
	require.NoError(t, err)
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Pattern
	})

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/nope", nil))

	assert.Empty(t, got)
}

func build(t *testing.T, r *route.Route) http.Handler {
	t.Helper()
	router, err := r.Build()
	require.NoError(t, err)

	return router
}
