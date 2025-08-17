package render

import (
    "encoding/json"
    "net/http"
)

func JSON(w http.ResponseWriter, _ *http.Request, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}

type ProblemDetails struct {
    Type   string `json:"type,omitempty"`
    Title  string `json:"title"`
    Status int    `json:"status"`
    Detail string `json:"detail,omitempty"`
}

func Problem(w http.ResponseWriter, _ *http.Request, status int, title, detail string) {
    w.Header().Set("Content-Type", "application/problem+json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(ProblemDetails{Title: title, Status: status, Detail: detail})
}
