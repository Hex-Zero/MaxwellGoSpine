package middleware_test

import (
	"github.com/hex-zero/MaxwellGoSpine/internal/metrics"
	appmw "github.com/hex-zero/MaxwellGoSpine/internal/middleware"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	m := metrics.NewRegistry()
	called := false
	h := appmw.Logging(logger, m)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true; w.WriteHeader(204) }))
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if !called {
		t.Fatalf("handler not called")
	}
	if w.Code != 204 {
		t.Fatalf("expected 204")
	}
}
