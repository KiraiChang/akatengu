package middleware

import "net/http"

// ─────────────────────────────────────────
// buildPattern scenarios
// ─────────────────────────────────────────

type buildPatternScenario struct {
	given   string
	when    string
	then    string
	prefix  string
	pattern string
	want    string
}

var buildPatternScenarios = []buildPatternScenario{
	{
		given:   `prefix="/api"，pattern 含 method "GET /journals"`,
		when:    "呼叫 buildPattern",
		then:    `回傳 "GET /api/journals"`,
		prefix:  "/api", pattern: "GET /journals", want: "GET /api/journals",
	},
	{
		given:   `prefix="/api"，pattern 含 method "POST /journals"`,
		when:    "呼叫 buildPattern",
		then:    `回傳 "POST /api/journals"`,
		prefix:  "/api", pattern: "POST /journals", want: "POST /api/journals",
	},
	{
		given:   `prefix="/api"，pattern 含路徑參數 "DELETE /journals/{id}"`,
		when:    "呼叫 buildPattern",
		then:    `回傳 "DELETE /api/journals/{id}"`,
		prefix:  "/api", pattern: "DELETE /journals/{id}", want: "DELETE /api/journals/{id}",
	},
	{
		given:   `prefix="/api"，pattern 無 method "/journals"`,
		when:    "呼叫 buildPattern",
		then:    `回傳 "/api/journals"`,
		prefix:  "/api", pattern: "/journals", want: "/api/journals",
	},
	{
		given:   `prefix="" 空字串，pattern 含 method "GET /journals"`,
		when:    "呼叫 buildPattern",
		then:    "原封不動回傳 pattern",
		prefix:  "", pattern: "GET /journals", want: "GET /journals",
	},
	{
		given:   `prefix="" 空字串，pattern 無 method "/journals"`,
		when:    "呼叫 buildPattern",
		then:    "原封不動回傳 pattern",
		prefix:  "", pattern: "/journals", want: "/journals",
	},
	{
		given:   `巢狀 prefix="/v2/admin"（SubGroup 情境），pattern "GET /tenants"`,
		when:    "呼叫 buildPattern",
		then:    `回傳 "GET /v2/admin/tenants"`,
		prefix:  "/v2/admin", pattern: "GET /tenants", want: "GET /v2/admin/tenants",
	},
}

// ─────────────────────────────────────────
// Method routing scenarios
// ─────────────────────────────────────────

type methodRoutingScenario struct {
	given    string
	when     string
	then     string
	method   string
	path     string
	wantCode int
	wantBody string
}

var methodRoutingScenarios = []methodRoutingScenario{
	{
		given:    "已註冊 GET /api/journals",
		when:     "送出 GET /api/journals",
		then:     "回傳 200 且 body 為 list",
		method:   "GET",
		path:     "/api/journals",
		wantCode: http.StatusOK,
		wantBody: "list",
	},
	{
		given:    "已註冊 POST /api/journals",
		when:     "送出 POST /api/journals",
		then:     "回傳 200 且 body 為 create",
		method:   "POST",
		path:     "/api/journals",
		wantCode: http.StatusOK,
		wantBody: "create",
	},
	{
		given:    "未註冊 DELETE /api/journals",
		when:     "送出 DELETE /api/journals",
		then:     "回傳 405 Method Not Allowed",
		method:   "DELETE",
		path:     "/api/journals",
		wantCode: http.StatusMethodNotAllowed,
		wantBody: "",
	},
}
