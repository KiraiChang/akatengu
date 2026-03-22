package ctxkey

type key string

const (
	RequestID  key = "request_id"
	UserClaims key = "user_claims"
)
