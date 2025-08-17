package middleware

import (
    "net/http"
    "strings"
)

// APIKeyAuth returns middleware enforcing presence of a valid API key.
// Keys: slice of accepted keys. Header precedence: X-API-Key then Authorization: ApiKey <key>.
type APIKeyOptions struct {
    Current []string
    Old     []string // accepted but deprecated
}

func APIKeyAuth(keys []string) func(http.Handler) http.Handler { // backward compat
    return APIKeyAuthWithOpts(APIKeyOptions{Current: keys})
}

func APIKeyAuthWithOpts(opts APIKeyOptions) func(http.Handler) http.Handler {
    if len(opts.Current) == 0 && len(opts.Old) == 0 {
        return func(next http.Handler) http.Handler { return next }
    }
    current := make(map[string]struct{}, len(opts.Current))
    for _, k := range opts.Current { current[k] = struct{}{} }
    old := make(map[string]struct{}, len(opts.Old))
    for _, k := range opts.Old { old[k] = struct{}{} }
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            var candidate string
            if v := r.Header.Get("X-API-Key"); v != "" { candidate = v }
            if candidate == "" {
                auth := r.Header.Get("Authorization")
                if strings.HasPrefix(strings.ToLower(auth), "apikey ") {
                    candidate = strings.TrimSpace(auth[7:])
                }
            }
            if candidate == "" {
                unauthorized(w)
                return
            }
            if _, ok := current[candidate]; ok {
                next.ServeHTTP(w, r)
                return
            }
            if _, ok := old[candidate]; ok {
                w.Header().Add("Warning", "299 - \"Deprecated API key in use; rotate to a current key\"")
                next.ServeHTTP(w, r)
                return
            }
            unauthorized(w)
        })
    }
}

func unauthorized(w http.ResponseWriter) {
    w.Header().Set("WWW-Authenticate", "ApiKey realm=api")
    http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}
