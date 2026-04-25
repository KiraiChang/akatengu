package middleware

import (
	"akatengu/internal/model/db"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/services"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// stubJwtService 實作 JwtService 供測試使用
type stubJwtService struct {
	claims *services.Claims
	err    error
}

func (s *stubJwtService) GenerateToken(_ *db.User) (string, error) { return "", nil }
func (s *stubJwtService) VerifyToken(_ string) (*services.Claims, error) {
	return s.claims, s.err
}

func jwtRequest(method, target, authHeader string) *http.Request {
	r := httptest.NewRequest(method, target, nil)
	if authHeader != "" {
		r.Header.Set("Authorization", authHeader)
	}
	return r
}

// next handler：把 context 裡的 claims 寫到回應 header 方便斷言
func claimsCapture(t *testing.T) (http.HandlerFunc, func() *services.Claims) {
	t.Helper()
	var got *services.Claims
	h := func(w http.ResponseWriter, r *http.Request) {
		got, _ = r.Context().Value(ctxkey.UserClaims).(*services.Claims)
		w.WriteHeader(http.StatusOK)
	}
	return h, func() *services.Claims { return got }
}

// ── Authorization header 缺失 ─────────────────────────────────────

func TestJwtMiddleware_MissingAuthHeader(t *testing.T) {
	svc := &stubJwtService{}
	mw := Jwt(svc)(okHandler("ok"))

	w := httptest.NewRecorder()
	mw.ServeHTTP(w, jwtRequest("GET", "/", ""))

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
	assertBodyContains(t, w, "missing token")
}

// ── scheme 不是 Bearer ────────────────────────────────────────────

func TestJwtMiddleware_NonBearerScheme(t *testing.T) {
	svc := &stubJwtService{}
	mw := Jwt(svc)(okHandler("ok"))

	w := httptest.NewRecorder()
	mw.ServeHTTP(w, jwtRequest("GET", "/", "Token sometoken"))

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
	assertBodyContains(t, w, "invalid token")
}

// ── 只有 "Bearer"，缺少 token ─────────────────────────────────────

func TestJwtMiddleware_BearerOnly_NoToken(t *testing.T) {
	svc := &stubJwtService{}
	mw := Jwt(svc)(okHandler("ok"))

	w := httptest.NewRecorder()
	mw.ServeHTTP(w, jwtRequest("GET", "/", "Bearer"))

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
	assertBodyContains(t, w, "invalid token")
}

// ── Bearer + 多餘欄位 ─────────────────────────────────────────────

func TestJwtMiddleware_TooManyParts(t *testing.T) {
	svc := &stubJwtService{}
	mw := Jwt(svc)(okHandler("ok"))

	w := httptest.NewRecorder()
	mw.ServeHTTP(w, jwtRequest("GET", "/", "Bearer token extra"))

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
	assertBodyContains(t, w, "invalid token")
}

// ── VerifyToken 回傳錯誤 ──────────────────────────────────────────

func TestJwtMiddleware_InvalidToken(t *testing.T) {
	svc := &stubJwtService{err: errors.New("unauthorized")}
	mw := Jwt(svc)(okHandler("ok"))

	w := httptest.NewRecorder()
	mw.ServeHTTP(w, jwtRequest("GET", "/", "Bearer badtoken"))

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
	assertBodyContains(t, w, "unauthorized")
}

// ── 有效 token：通過並注入 claims ─────────────────────────────────

func TestJwtMiddleware_ValidToken_PassThrough(t *testing.T) {
	want := &services.Claims{UserID: 7, UserName: "alice"}
	svc := &stubJwtService{claims: want}

	next, getClaims := claimsCapture(t)
	mw := Jwt(svc)(next)

	w := httptest.NewRecorder()
	mw.ServeHTTP(w, jwtRequest("GET", "/", "Bearer validtoken"))

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	got := getClaims()
	if got == nil {
		t.Fatal("claims not set in context")
	}
	if got.UserID != want.UserID {
		t.Errorf("UserID: want %d, got %d", want.UserID, got.UserID)
	}
	if got.UserName != want.UserName {
		t.Errorf("UserName: want %q, got %q", want.UserName, got.UserName)
	}
}

// ── 整合：使用真實 JwtService 產生 token 後通過 middleware ────────

func TestJwtMiddleware_RealToken_PassThrough(t *testing.T) {
	jwtSvc := services.NewJWT("test-secret")
	user := &db.User{UserId: 99, Username: "realuser"}
	token, err := jwtSvc.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	next, getClaims := claimsCapture(t)
	mw := Jwt(jwtSvc)(next)

	w := httptest.NewRecorder()
	mw.ServeHTTP(w, jwtRequest("GET", "/", "Bearer "+token))

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	got := getClaims()
	if got == nil {
		t.Fatal("claims not set in context")
	}
	if got.UserID != user.UserId {
		t.Errorf("UserID: want %d, got %d", user.UserId, got.UserID)
	}
}

// ── helpers ───────────────────────────────────────────────────────

func assertBodyContains(t *testing.T, w *httptest.ResponseRecorder, substr string) {
	t.Helper()
	body := w.Body.String()
	if len(body) == 0 || !contains(body, substr) {
		t.Errorf("body = %q, want it to contain %q", body, substr)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && stringContains(s, sub))
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
