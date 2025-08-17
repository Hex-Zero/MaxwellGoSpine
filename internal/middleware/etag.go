package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

// ETag sets an ETag header for cacheable 200 JSON responses and handles If-None-Match.
func ETag(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr := &respRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rr, r)
		if rr.status == http.StatusOK && rr.hdrContentTypeJSON() && len(rr.body) > 0 {
			sum := sha256.Sum256(rr.body)
			etag := "\"" + hex.EncodeToString(sum[:8]) + "\"" // short strong etag
			if inm := r.Header.Get("If-None-Match"); inm != "" && inm == etag {
				w.Header().Set("ETag", etag)
				w.WriteHeader(http.StatusNotModified)
				return
			}
			w.Header().Set("ETag", etag)
			// rewrite body
			if !rr.wrote {
				w.WriteHeader(rr.status)
			}
			_, _ = w.Write(rr.body)
		}
	})
}

type respRecorder struct {
	http.ResponseWriter
	body   []byte
	wrote  bool
	status int
}

func (r *respRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
	r.wrote = true
}
func (r *respRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return len(b), nil
}
func (r *respRecorder) hdrContentTypeJSON() bool {
	ct := r.Header().Get("Content-Type")
	return ct == "application/json" || ct == "application/json; charset=utf-8"
}
