package ctxkey

import (
	"context"
	"fmt"

	jwtpkg "akatengu/internal/pkg/jwt"
)

type key string

const (
	RequestID  key = "request_id"
	UserClaims key = "user_claims"
	MerchantID key = "merchant_id"
)

func GetMerchantID(ctx context.Context) (int64, error) {
	v, ok := ctx.Value(MerchantID).(int64)
	if !ok || v == 0 {
		return 0, fmt.Errorf("merchant_id not in context — missing MerchantAuth middleware")
	}
	return v, nil
}

func GetUserName(ctx context.Context) string {
	claims, ok := ctx.Value(UserClaims).(*jwtpkg.Claims)
	if !ok || claims == nil {
		return ""
	}
	return claims.UserName
}
