package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func Gzip(level int) func(http.Handler) http.Handler {
	if level < gzip.HuffmanOnly {
		level = gzip.BestSpeed
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}
			gz, err := gzip.NewWriterLevel(w, level)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			defer gz.Close()
			w.Header().Set("Content-Encoding", "gzip")
			gw := &gzipResponseWriter{ResponseWriter: w, Writer: gz}
			next.ServeHTTP(gw, r)
		})
	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) { return w.Writer.Write(b) }
