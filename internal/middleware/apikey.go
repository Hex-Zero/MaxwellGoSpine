package middleware

import (
    "net/http"
    "strings"
)

// APIKeyAuth returns middleware enforcing presence of a valid API key.
// Keys: slice of accepted keys. Header precedence: X-API-Key then Authorization: ApiKey <key>.
func APIKeyAuth(keys []string) func(http.Handler) http.Handler {
    if len(keys) == 0 { // no auth enforced
        return func(next http.Handler) http.Handler { return next }
    }
    keySet := make(map[string]struct{}, len(keys))
    for _, k := range keys { keySet[k] = struct{}{} }
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            var candidate string
            if v := r.Header.Get("X-API-Key"); v != "" { candidate = v }
            if candidate == "" { // fallback Authorization
                auth := r.Header.Get("Authorization")
                if strings.HasPrefix(strings.ToLower(auth), "apikey ") {
                    candidate = strings.TrimSpace(auth[7:])
                }
            }
            if candidate == "" {
                unauthorized(w)
                return
            }
            if _, ok := keySet[candidate]; !ok {
                unauthorized(w)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

func unauthorized(w http.ResponseWriter) {
    w.Header().Set("WWW-Authenticate", "ApiKey realm=api")
    http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}
