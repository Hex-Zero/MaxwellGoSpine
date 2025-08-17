package middleware

import (
    "net/http"
    "time"
    "go.uber.org/zap"
    "github.com/hex-zero/MaxwellGoSpine/internal/metrics"
)

type statusWriter struct {
    http.ResponseWriter
    status int
}

func (w *statusWriter) WriteHeader(code int) { w.status = code; w.ResponseWriter.WriteHeader(code) }

func Logging(logger *zap.Logger, m *metrics.Registry) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            sw := &statusWriter{ResponseWriter: w, status: 200}
            next.ServeHTTP(sw, r)
            dur := time.Since(start)
            if m != nil {
                m.HTTPRequests.WithLabelValues(r.Method, r.URL.Path, http.StatusText(sw.status)).Inc()
                m.HTTPDuration.WithLabelValues(r.Method, r.URL.Path).Observe(dur.Seconds())
            }
            logger.Info("request",
                zap.String("request_id", GetRequestID(r.Context())),
                zap.String("method", r.Method),
                zap.String("path", r.URL.Path),
                zap.Int("status", sw.status),
                zap.Duration("duration", dur),
                zap.String("user_agent", r.UserAgent()),
                zap.String("remote_ip", r.RemoteAddr),
            )
        })
    }
}
