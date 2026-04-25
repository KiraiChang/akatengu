package response

import (
	"encoding/json"
	"net/http"
)

// OK writes a bare JSON success response with status 200.
func OK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// WriteError writes an RFC 9457 Problem Details response.
// title should be a short, human-readable summary of the problem type (e.g. "Bad Request").
// detail should describe this specific occurrence (e.g. the underlying error message).
func WriteError(w http.ResponseWriter, r *http.Request, status int, title, detail string) {
	p := Problem{
		Type:     "about:blank",
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: r.RequestURI,
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(p)
}