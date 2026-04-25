package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// ── 測試用 Middleware ────────────────────────────────────────────

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

// ── buildPattern ─────────────────────────────────────────────────

func TestBuildPattern(t *testing.T) {
	cases := []struct {
		prefix, pattern, want string
	}{
		// 有 prefix + 有 method
		{"/api", "GET /journals", "GET /api/journals"},
		{"/api", "POST /journals", "POST /api/journals"},
		{"/api", "DELETE /journals/{id}", "DELETE /api/journals/{id}"},

		// 有 prefix + 無 method
		{"/api", "/journals", "/api/journals"},

		// 空 prefix — 原封不動
		{"", "GET /journals", "GET /journals"},
		{"", "/journals", "/journals"},

		// 巢狀 prefix（SubGroup 情境）
		{"/v2/admin", "GET /tenants", "GET /v2/admin/tenants"},
	}

	for _, c := range cases {
		g := &Group{prefix: c.prefix}
		got := g.buildPattern(c.pattern)
		if got != c.want {
			t.Errorf("buildPattern(%q, %q)\n got  %q\n want %q",
				c.prefix, c.pattern, got, c.want)
		}
	}
}

// ── NewGroup ──────────────────────────────────────────────────────

func TestNewGroup_PrefixApplied(t *testing.T) {
	mux := http.NewServeMux()
	g := NewGroup(mux, "/api", markMiddleware("X-Group", "api"))
	g.HandleFunc("GET /journals", okHandler("journals"))

	r := httptest.NewRequest("GET", "/api/journals", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("X-Group"); got != "api" {
		t.Errorf("middleware not applied: X-Group = %q", got)
	}
	if w.Body.String() != "journals" {
		t.Errorf("unexpected body: %q", w.Body.String())
	}
}

func TestNewGroup_WrongPrefix_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	g := NewGroup(mux, "/api")
	g.HandleFunc("GET /journals", okHandler("journals"))

	r := httptest.NewRequest("GET", "/journals", nil) // 缺少 /api prefix
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ── NewGroupNoPrefix ──────────────────────────────────────────────

func TestNewGroupNoPrefix_PatternUnchanged(t *testing.T) {
	mux := http.NewServeMux()
	g := NewGroupNoPrefix(mux, markMiddleware("X-Group", "noprefix"))
	g.HandleFunc("GET /journals", okHandler("journals"))

	r := httptest.NewRequest("GET", "/journals", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("X-Group"); got != "noprefix" {
		t.Errorf("middleware not applied: X-Group = %q", got)
	}
}

func TestNewGroupNoPrefix_MiddlewareNotLeakToOtherRoutes(t *testing.T) {
	mux := http.NewServeMux()

	g := NewGroupNoPrefix(mux, markMiddleware("X-Auth", "jwt"))
	g.HandleFunc("GET /journals", okHandler("journals"))

	// 公開路由直接掛 mux，不經過 Group
	mux.HandleFunc("GET /health", okHandler("ok"))

	r := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("X-Auth"); got != "" {
		t.Errorf("middleware leaked to public route: X-Auth = %q", got)
	}
}

// ── SubGroup ──────────────────────────────────────────────────────

func TestSubGroup_InheritsParentPrefixAndMiddleware(t *testing.T) {
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

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	for key, want := range map[string]string{
		"X-MW-1": "logger",
		"X-MW-2": "jwt",
		"X-MW-3": "role",
	} {
		if got := w.Header().Get(key); got != want {
			t.Errorf("middleware %q: got %q, want %q", key, got, want)
		}
	}
}

func TestSubGroup_DoesNotAffectParentRoutes(t *testing.T) {
	mux := http.NewServeMux()

	parent := NewGroup(mux, "/v2", markMiddleware("X-MW-1", "jwt"))
	parent.HandleFunc("GET /journals", okHandler("journals"))

	child := parent.SubGroup("/admin", markMiddleware("X-MW-2", "role"))
	child.HandleFunc("GET /tenants", okHandler("tenants"))

	// 父層路由不應該有 X-MW-2
	r := httptest.NewRequest("GET", "/v2/journals", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("X-MW-2"); got != "" {
		t.Errorf("child middleware leaked to parent route: X-MW-2 = %q", got)
	}
}

// ── Chain 執行順序 ─────────────────────────────────────────────────

func TestChain_ExecutionOrder(t *testing.T) {
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

	want := []string{
		"A:before", "B:before", "C:before",
		"handler",
		"C:after", "B:after", "A:after",
	}
	if len(order) != len(want) {
		t.Fatalf("order length mismatch:\n got  %v\n want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("order[%d] = %q, want %q", i, order[i], want[i])
		}
	}
}

// ── Method 路由 ────────────────────────────────────────────────────

func TestGroup_MethodRouting(t *testing.T) {
	mux := http.NewServeMux()
	g := NewGroup(mux, "/api")
	g.HandleFunc("GET /journals", okHandler("list"))
	g.HandleFunc("POST /journals", okHandler("create"))

	cases := []struct {
		method, path string
		wantCode     int
		wantBody     string
	}{
		{"GET", "/api/journals", http.StatusOK, "list"},
		{"POST", "/api/journals", http.StatusOK, "create"},
		{"DELETE", "/api/journals", http.StatusMethodNotAllowed, ""},
	}

	for _, c := range cases {
		r := httptest.NewRequest(c.method, c.path, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != c.wantCode {
			t.Errorf("%s %s: got %d, want %d", c.method, c.path, w.Code, c.wantCode)
		}
		if c.wantBody != "" && w.Body.String() != c.wantBody {
			t.Errorf("%s %s: body = %q, want %q", c.method, c.path, w.Body.String(), c.wantBody)
		}
	}
}
