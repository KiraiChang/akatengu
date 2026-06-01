package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// ─────────────────────────────────────────
// Helpers（共用於 jwt_test.go）
// ─────────────────────────────────────────

func markMiddleware(key, val string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(key, val)
			next.ServeHTTP(w, r)
		})
	}
}

func okHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}
}

// ─────────────────────────────────────────
// Spec runner
// ─────────────────────────────────────────

var _ = Describe("Group.buildPattern", func() {
	for _, s := range buildPatternScenarios {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			g := &Group{prefix: s.prefix}
			Expect(g.buildPattern(s.pattern)).To(Equal(s.want))
		})
	}
})

var _ = Describe("NewGroup", func() {
	Context("prefix 正確，middleware 已設定", func() {
		It("GIVEN prefix=/api 與一個標記 middleware\n  WHEN 請求打到 /api/journals\n  THEN 回傳 200、body 正確、middleware header 已設定", func() {
			mux := http.NewServeMux()
			g := NewGroup(mux, "/api", markMiddleware("X-Group", "api"))
			g.HandleFunc("GET /journals", okHandler("journals"))

			r := httptest.NewRequest("GET", "/api/journals", nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Header().Get("X-Group")).To(Equal("api"))
			Expect(w.Body.String()).To(Equal("journals"))
		})
	})

	Context("prefix 錯誤（請求不帶 prefix）", func() {
		It("GIVEN prefix=/api，路由已註冊\n  WHEN 請求不帶 /api prefix\n  THEN 回傳 404", func() {
			mux := http.NewServeMux()
			g := NewGroup(mux, "/api")
			g.HandleFunc("GET /journals", okHandler("journals"))

			r := httptest.NewRequest("GET", "/journals", nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)

			Expect(w.Code).To(Equal(http.StatusNotFound))
		})
	})
})

var _ = Describe("NewGroupNoPrefix", func() {
	Context("pattern 不加 prefix", func() {
		It("GIVEN 無 prefix 的 group，middleware 已設定\n  WHEN 請求打到 /journals\n  THEN 回傳 200 且 middleware header 已設定", func() {
			mux := http.NewServeMux()
			g := NewGroupNoPrefix(mux, markMiddleware("X-Group", "noprefix"))
			g.HandleFunc("GET /journals", okHandler("journals"))

			r := httptest.NewRequest("GET", "/journals", nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Header().Get("X-Group")).To(Equal("noprefix"))
		})
	})

	Context("middleware 不洩漏到未加入 group 的路由", func() {
		It("GIVEN group 帶有 X-Auth middleware，公開路由直接掛 mux\n  WHEN 請求打到公開路由 /health\n  THEN 回傳 200 且無 X-Auth header", func() {
			mux := http.NewServeMux()
			g := NewGroupNoPrefix(mux, markMiddleware("X-Auth", "jwt"))
			g.HandleFunc("GET /journals", okHandler("journals"))
			mux.HandleFunc("GET /health", okHandler("ok"))

			r := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Header().Get("X-Auth")).To(BeEmpty())
		})
	})
})

var _ = Describe("SubGroup", func() {
	Context("繼承父層 prefix 與所有 middleware", func() {
		It("GIVEN 父層 prefix=/v2 含兩個 middleware，子群組 /admin 含第三個 middleware\n  WHEN 請求打到 /v2/admin/tenants\n  THEN 三個 middleware header 均存在", func() {
			mux := http.NewServeMux()
			parent := NewGroup(mux, "/v2",
				markMiddleware("X-MW-1", "logger"),
				markMiddleware("X-MW-2", "jwt"),
			)
			child := parent.SubGroup("/admin", markMiddleware("X-MW-3", "role"))
			child.HandleFunc("GET /tenants", okHandler("tenants"))

			r := httptest.NewRequest("GET", "/v2/admin/tenants", nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Header().Get("X-MW-1")).To(Equal("logger"))
			Expect(w.Header().Get("X-MW-2")).To(Equal("jwt"))
			Expect(w.Header().Get("X-MW-3")).To(Equal("role"))
		})
	})

	Context("不影響父層路由", func() {
		It("GIVEN 父層路由 /v2/journals，子群組新增 X-MW-2 middleware\n  WHEN 請求打到父層路由 /v2/journals\n  THEN 回傳 200 且無 X-MW-2 header", func() {
			mux := http.NewServeMux()
			parent := NewGroup(mux, "/v2", markMiddleware("X-MW-1", "jwt"))
			parent.HandleFunc("GET /journals", okHandler("journals"))

			child := parent.SubGroup("/admin", markMiddleware("X-MW-2", "role"))
			child.HandleFunc("GET /tenants", okHandler("tenants"))

			r := httptest.NewRequest("GET", "/v2/journals", nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Header().Get("X-MW-2")).To(BeEmpty())
		})
	})
})

var _ = Describe("Chain 執行順序", func() {
	It("GIVEN middleware A、B、C 依序加入\n  WHEN 請求通過 chain\n  THEN 執行順序為 A:before B:before C:before handler C:after B:after A:after", func() {
		var order []string

		mw := func(label string) Middleware {
			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					order = append(order, label+":before")
					next.ServeHTTP(w, r)
					order = append(order, label+":after")
				})
			}
		}

		mux := http.NewServeMux()
		g := NewGroupNoPrefix(mux, mw("A"), mw("B"), mw("C"))
		g.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "handler")
		})

		r := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		Expect(order).To(Equal([]string{
			"A:before", "B:before", "C:before",
			"handler",
			"C:after", "B:after", "A:after",
		}))
	})
})

var _ = Describe("Group method routing", func() {
	var mux *http.ServeMux

	BeforeEach(func() {
		mux = http.NewServeMux()
		g := NewGroup(mux, "/api")
		g.HandleFunc("GET /journals", okHandler("list"))
		g.HandleFunc("POST /journals", okHandler("create"))
	})

	for _, s := range methodRoutingScenarios {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			r := httptest.NewRequest(s.method, s.path, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)

			Expect(w.Code).To(Equal(s.wantCode))
			if s.wantBody != "" {
				Expect(w.Body.String()).To(Equal(s.wantBody))
			}
		})
	}
})
