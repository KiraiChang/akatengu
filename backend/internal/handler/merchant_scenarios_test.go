package handler

import (
	"net/http"
)

type merchantHandlerScenario struct {
	given    string
	when     string
	then     string
	method   string
	path     string
	body     string
	svcErr   error
	wantCode int
	wantBody string
}

var updateMerchantScenarios = []merchantHandlerScenario{
	{
		given:    "有效的 JWT 與請求 body",
		when:     "PUT /merchant/1",
		then:     "回傳 200 且 body 為空物件",
		method:   http.MethodPut,
		path:     "/merchant/1",
		body:     `{"name":"新商戶","display_name":"新顯示名稱","currency":"USD"}`,
		wantCode: http.StatusOK,
		wantBody: "{}",
	},
	{
		given:    "merchant_id 非數字",
		when:     "PUT /merchant/abc",
		then:     "回傳 400",
		method:   http.MethodPut,
		path:     "/merchant/abc",
		body:     `{"name":"A","display_name":"B","currency":"TWD"}`,
		wantCode: http.StatusBadRequest,
	},
	{
		given:    "request body 為無效 JSON",
		when:     "PUT /merchant/1",
		then:     "回傳 400",
		method:   http.MethodPut,
		path:     "/merchant/1",
		body:     `not-json`,
		wantCode: http.StatusBadRequest,
	},
	{
		given:    "name 欄位空白",
		when:     "PUT /merchant/1",
		then:     "回傳 422",
		method:   http.MethodPut,
		path:     "/merchant/1",
		body:     `{"name":"","display_name":"B","currency":"TWD"}`,
		wantCode: http.StatusUnprocessableEntity,
	},
	{
		given:    "使用者不屬於此商戶（service 回傳 ErrMerchantForbidden）",
		when:     "PUT /merchant/1",
		then:     "回傳 403",
		method:   http.MethodPut,
		path:     "/merchant/1",
		body:     `{"name":"A","display_name":"B","currency":"TWD"}`,
		svcErr:   errMerchantForbiddenStub,
		wantCode: http.StatusForbidden,
	},
	{
		given:    "DB 發生非預期錯誤",
		when:     "PUT /merchant/1",
		then:     "回傳 500",
		method:   http.MethodPut,
		path:     "/merchant/1",
		body:     `{"name":"A","display_name":"B","currency":"TWD"}`,
		svcErr:   errInternalStub,
		wantCode: http.StatusInternalServerError,
	},
}

var deactivateMerchantScenarios = []merchantHandlerScenario{
	{
		given:    "有效的 JWT 且使用者屬於此商戶",
		when:     "DELETE /merchant/1",
		then:     "回傳 200 且 body 為空物件",
		method:   http.MethodDelete,
		path:     "/merchant/1",
		wantCode: http.StatusOK,
		wantBody: "{}",
	},
	{
		given:    "merchant_id 非數字",
		when:     "DELETE /merchant/abc",
		then:     "回傳 400",
		method:   http.MethodDelete,
		path:     "/merchant/abc",
		wantCode: http.StatusBadRequest,
	},
	{
		given:    "使用者不屬於此商戶（service 回傳 ErrMerchantForbidden）",
		when:     "DELETE /merchant/1",
		then:     "回傳 403",
		method:   http.MethodDelete,
		path:     "/merchant/1",
		svcErr:   errMerchantForbiddenStub,
		wantCode: http.StatusForbidden,
	},
	{
		given:    "DB 發生非預期錯誤",
		when:     "DELETE /merchant/1",
		then:     "回傳 500",
		method:   http.MethodDelete,
		path:     "/merchant/1",
		svcErr:   errInternalStub,
		wantCode: http.StatusInternalServerError,
	},
}
