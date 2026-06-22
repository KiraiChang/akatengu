package middleware

import (
	"akatengu/internal/kernel/event"
	"akatengu/internal/pkg/ctxkey"
	"context"
	"net/http"
)

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = event.NewUUIDv7()
		}

		ctx := context.WithValue(r.Context(), ctxkey.RequestID, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
