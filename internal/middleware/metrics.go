package middleware

import (
    "net/http"
    "time"
    "github.com/hex-zero/MaxwellGoSpine/internal/metrics"
)

func Metrics(m *metrics.Registry) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if m == nil { next.ServeHTTP(w, r); return }
            start := time.Now()
            next.ServeHTTP(w, r)
            m.HTTPDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
        })
    }
}
