package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

type Registry struct {
    Gatherer *prometheus.Registry
    HTTPRequests *prometheus.CounterVec
    HTTPDuration *prometheus.HistogramVec
}

func NewRegistry() *Registry {
    reg := prometheus.NewRegistry()
    r := &Registry{Gatherer: reg}
    r.HTTPRequests = promauto.With(reg).NewCounterVec(prometheus.CounterOpts{Name: "http_requests_total", Help: "Total HTTP requests"}, []string{"method", "path", "status"})
    r.HTTPDuration = promauto.With(reg).NewHistogramVec(prometheus.HistogramOpts{Name: "http_request_duration_seconds", Help: "Duration", Buckets: prometheus.DefBuckets}, []string{"method", "path"})
    return r
}
