package middleware

import (
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/pkg/jwt"
	"akatengu/internal/repos/query"
	"context"
	"database/sql"
	"errors"
	"net/http"
)

func MerchantAuth(repo query.MerchantRepo) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ctxkey.UserClaims).(*jwt.Claims)
			if !ok || claims == nil || claims.MerchantID == 0 {
				http.Error(w, "merchant not selected", http.StatusForbidden)
				return
			}

			_, err := repo.GetUserRole(r.Context(), claims.UserID, claims.MerchantID)
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), ctxkey.MerchantID, claims.MerchantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
