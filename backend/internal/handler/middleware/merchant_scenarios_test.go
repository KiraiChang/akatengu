package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
)

type merchantAuthScenario struct {
	given    string
	when     string
	then     string
	ctx      context.Context
	stubErr  error
	wantCode int
	wantBody string
	checkMID bool
	wantMID  int64
}

var merchantAuthScenarios = []merchantAuthScenario{
	{
		given:    "context 中沒有 UserClaims",
		when:     "收到任意 HTTP 請求",
		then:     "回傳 403 且訊息包含 merchant not selected",
		ctx:      context.Background(),
		wantCode: http.StatusForbidden,
		wantBody: "merchant not selected",
	},
	{
		given:    "Claims 存在但 MerchantID 為 0",
		when:     "收到任意 HTTP 請求",
		then:     "回傳 403 且訊息包含 merchant not selected",
		ctx:      claimsCtx(0),
		wantCode: http.StatusForbidden,
		wantBody: "merchant not selected",
	},
	{
		given:    "有效 Claims 且 GetUserRole 回傳 sql.ErrNoRows（使用者不屬於此 Merchant）",
		when:     "收到任意 HTTP 請求",
		then:     "回傳 403 且訊息包含 forbidden",
		ctx:      claimsCtx(5),
		stubErr:  sql.ErrNoRows,
		wantCode: http.StatusForbidden,
		wantBody: "forbidden",
	},
	{
		given:    "有效 Claims 且 GetUserRole 回傳非預期錯誤",
		when:     "收到任意 HTTP 請求",
		then:     "回傳 500 且訊息包含 internal server error",
		ctx:      claimsCtx(5),
		stubErr:  errors.New("db connection timeout"),
		wantCode: http.StatusInternalServerError,
		wantBody: "internal server error",
	},
	{
		given:    "有效 Claims 且 GetUserRole 成功",
		when:     "收到任意 HTTP 請求",
		then:     "放行請求並回傳 200",
		ctx:      claimsCtx(5),
		wantCode: http.StatusOK,
	},
	{
		given:    "有效 Claims（MerchantID=5）且 GetUserRole 成功",
		when:     "收到任意 HTTP 請求",
		then:     "將 MerchantID=5 注入到下游 context",
		ctx:      claimsCtx(5),
		wantCode: http.StatusOK,
		checkMID: true,
		wantMID:  5,
	},
}
