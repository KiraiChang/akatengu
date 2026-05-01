package middleware

import (
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/pkg/jwt"
	"context"
	"net/http"
	"strings"
)

func Jwt(jwt jwt.JwtService) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}

			parts := strings.Fields(header)
			if parts[0] != "Bearer" {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			if len(parts) != 2 {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			claims, err := jwt.VerifyToken(parts[1])
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ctxkey.UserClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
