package middleware_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    appmw "github.com/hex-zero/MaxwellGoSpine/internal/middleware"
)

func TestAPIKeyAuth(t *testing.T) {
    mw := appmw.APIKeyAuth([]string{"secret1", "secret2"})
    called := false
    handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))

    // No key
    rr := httptest.NewRecorder()
    req, _ := http.NewRequest(http.MethodGet, "/", nil)
    handler.ServeHTTP(rr, req)
    if rr.Code != http.StatusUnauthorized || called {
        t.Fatalf("expected 401 and not called, got %d called=%v", rr.Code, called)
    }

    // Wrong key
    rr = httptest.NewRecorder()
    req, _ = http.NewRequest(http.MethodGet, "/", nil)
    req.Header.Set("X-API-Key", "bad")
    handler.ServeHTTP(rr, req)
    if rr.Code != http.StatusUnauthorized {
        t.Fatalf("expected 401 for wrong key, got %d", rr.Code)
    }

    // Correct key
    rr = httptest.NewRecorder()
    req, _ = http.NewRequest(http.MethodGet, "/", nil)
    req.Header.Set("X-API-Key", "secret2")
    handler.ServeHTTP(rr, req)
    if rr.Code != http.StatusOK || !called {
        t.Fatalf("expected 200 and handler execution, got %d called=%v", rr.Code, called)
    }
}

func TestAPIKeyAuthWithDeprecated(t *testing.T) {
    mw := appmw.APIKeyAuthWithOpts(appmw.APIKeyOptions{Current: []string{"newkey"}, Old: []string{"oldkey"}})
    called := false
    handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true; w.WriteHeader(http.StatusOK) }))

    // Deprecated key should pass and include Warning header
    rr := httptest.NewRecorder()
    req, _ := http.NewRequest(http.MethodGet, "/", nil)
    req.Header.Set("X-API-Key", "oldkey")
    handler.ServeHTTP(rr, req)
    if rr.Code != http.StatusOK || !called {
        t.Fatalf("deprecated key should allow access; got %d called=%v", rr.Code, called)
    }
    if rr.Header().Get("Warning") == "" {
        t.Fatalf("expected Warning header for deprecated key")
    }
}

func TestAPIKeyAuthExpiry(t *testing.T) {
    tomorrow := time.Now().UTC().Add(24 * time.Hour)
    yesterday := time.Now().UTC().Add(-24 * time.Hour)
    exp := map[string]int64{}
    day := func(t time.Time) int64 { d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC); return d.Unix() }
    exp["valid"] = day(tomorrow)
    exp["expired"] = day(yesterday)
    mw := appmw.APIKeyAuthWithOpts(appmw.APIKeyOptions{Current: []string{"valid", "expired"}, Expiries: exp})
    base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
    h := mw(base)

    // valid key
    rr := httptest.NewRecorder()
    req, _ := http.NewRequest(http.MethodGet, "/", nil)
    req.Header.Set("X-API-Key", "valid")
    h.ServeHTTP(rr, req)
    if rr.Code != http.StatusOK { t.Fatalf("expected 200 for valid key, got %d", rr.Code) }

    // expired key
    rr = httptest.NewRecorder()
    req, _ = http.NewRequest(http.MethodGet, "/", nil)
    req.Header.Set("X-API-Key", "expired")
    h.ServeHTTP(rr, req)
    if rr.Code != http.StatusUnauthorized { t.Fatalf("expected 401 for expired key, got %d", rr.Code) }
}
