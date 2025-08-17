package middleware

import (
    "net/http"
    "strings"
    "time"
)

// APIKeyAuth returns middleware enforcing presence of a valid API key.
// Keys: slice of accepted keys. Header precedence: X-API-Key then Authorization: ApiKey <key>.
type APIKeyOptions struct {
    Current []string
    Old     []string // accepted but deprecated
    Expiries map[string]int64 // unix date (start of day) expiry (exclusive)
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
                if isExpired(candidate, opts.Expiries) {
                    unauthorized(w)
                    return
                }
                next.ServeHTTP(w, r)
                return
            }
            if _, ok := old[candidate]; ok {
                if isExpired(candidate, opts.Expiries) { // even old keys can expire
                    unauthorized(w)
                    return
                }
                w.Header().Add("Warning", "299 - \"Deprecated API key in use; rotate to a current key\"")
                next.ServeHTTP(w, r)
                return
            }
            unauthorized(w)
        })
    }
}

func isExpired(key string, expiries map[string]int64) bool {
    if len(expiries) == 0 { return false }
    if ts, ok := expiries[key]; ok {
        // expiry stored as unix date boundary (00:00 UTC). Compare current UTC date.
        now := time.Now().UTC().Unix()
        if now >= ts { return true }
    }
    return false
}

func unauthorized(w http.ResponseWriter) {
    w.Header().Set("WWW-Authenticate", "ApiKey realm=api")
    http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}
