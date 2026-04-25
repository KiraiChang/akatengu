package response

// Problem is the RFC 9457 Problem Details envelope for error responses.
// Clients can detect an error by checking Content-Type: application/problem+json.
type Problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}