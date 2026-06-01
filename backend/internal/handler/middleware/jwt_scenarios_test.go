package middleware

import (
	"errors"
	"net/http"

	"akatengu/internal/pkg/jwt"
)

type jwtScenario struct {
	given      string
	when       string
	then       string
	authHeader string
	stubErr    error
	stubClaims *jwt.Claims
	wantCode   int
	wantBody   string
}

var jwtScenarios = []jwtScenario{
	{
		given:      "Authorization header 缺失",
		when:       "收到 GET 請求",
		then:       "回傳 401 且訊息包含 missing token",
		authHeader: "",
		wantCode:   http.StatusUnauthorized,
		wantBody:   "missing token",
	},
	{
		given:      "scheme 不是 Bearer（使用 Token）",
		when:       "收到 GET 請求",
		then:       "回傳 401 且訊息包含 invalid token",
		authHeader: "Token sometoken",
		wantCode:   http.StatusUnauthorized,
		wantBody:   "invalid token",
	},
	{
		given:      "只有 Bearer 關鍵字，缺少 token 部分",
		when:       "收到 GET 請求",
		then:       "回傳 401 且訊息包含 invalid token",
		authHeader: "Bearer",
		wantCode:   http.StatusUnauthorized,
		wantBody:   "invalid token",
	},
	{
		given:      "Bearer 後帶有多餘欄位",
		when:       "收到 GET 請求",
		then:       "回傳 401 且訊息包含 invalid token",
		authHeader: "Bearer token extra",
		wantCode:   http.StatusUnauthorized,
		wantBody:   "invalid token",
	},
	{
		given:      "VerifyToken 回傳 unauthorized 錯誤",
		when:       "收到帶有 Bearer token 的請求",
		then:       "回傳 401 且訊息包含 unauthorized",
		authHeader: "Bearer badtoken",
		stubErr:    errors.New("unauthorized"),
		wantCode:   http.StatusUnauthorized,
		wantBody:   "unauthorized",
	},
	{
		given:      "VerifyToken 驗證成功，回傳有效 Claims",
		when:       "收到帶有 Bearer token 的請求",
		then:       "放行請求並將 Claims（UserID=7、UserName=alice）注入到下游 context",
		authHeader: "Bearer validtoken",
		stubClaims: &jwt.Claims{UserID: 7, UserName: "alice"},
		wantCode:   http.StatusOK,
	},
}
