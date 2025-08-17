package middleware

import (
	"net/http"
	"strings"
)

func CORS(allowed []string) func(http.Handler) http.Handler {
	allowAll := len(allowed) == 0
	allowedSet := map[string]struct{}{}
	for _, o := range allowed {
		allowedSet[strings.TrimSpace(o)] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && (allowAll || contains(allowedSet, origin)) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Request-ID")
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func contains(set map[string]struct{}, v string) bool { _, ok := set[v]; return ok }
