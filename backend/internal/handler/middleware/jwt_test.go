package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"akatengu/internal/model/db"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/pkg/jwt"
)

// ─────────────────────────────────────────
// Stubs
// ─────────────────────────────────────────

type stubJwtService struct {
	claims *jwt.Claims
	err    error
}

func (s *stubJwtService) GenerateToken(_ *db.User) (string, error) { return "", nil }
func (s *stubJwtService) GenerateTokenWithMerchant(_ *db.User, _ int64, _ string) (string, error) {
	return "", nil
}
func (s *stubJwtService) VerifyToken(_ string) (*jwt.Claims, error) {
	return s.claims, s.err
}

// ─────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────

func jwtRequest(authHeader string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if authHeader != "" {
		r.Header.Set("Authorization", authHeader)
	}
	return r
}

func claimsCapture() (http.HandlerFunc, func() *jwt.Claims) {
	var got *jwt.Claims
	h := func(w http.ResponseWriter, r *http.Request) {
		got, _ = r.Context().Value(ctxkey.UserClaims).(*jwt.Claims)
		w.WriteHeader(http.StatusOK)
	}
	return h, func() *jwt.Claims { return got }
}

// ─────────────────────────────────────────
// Spec runner
// ─────────────────────────────────────────

var _ = Describe("Jwt middleware", func() {
	for _, s := range jwtScenarios {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			svc := &stubJwtService{claims: s.stubClaims, err: s.stubErr}
			next, getClaims := claimsCapture()
			rec := httptest.NewRecorder()

			Jwt(svc)(next).ServeHTTP(rec, jwtRequest(s.authHeader))

			Expect(rec.Code).To(Equal(s.wantCode))
			if s.wantBody != "" {
				Expect(rec.Body.String()).To(ContainSubstring(s.wantBody))
			}
			if s.stubClaims != nil {
				got := getClaims()
				Expect(got).NotTo(BeNil())
				Expect(got.UserID).To(Equal(s.stubClaims.UserID))
				Expect(got.UserName).To(Equal(s.stubClaims.UserName))
			}
		})
	}

	Context("整合：使用真實 JwtService", func() {
		It("GIVEN 以真實 JwtService 產生有效 token\n  WHEN 請求帶上此 token\n  THEN 放行並將正確 UserID 注入 context，回傳 200", func() {
			jwtSvc := jwt.NewJWT("test-secret")
			user := &db.User{UserId: 99, Username: "realuser"}
			token, err := jwtSvc.GenerateToken(user)
			Expect(err).NotTo(HaveOccurred())

			next, getClaims := claimsCapture()
			rec := httptest.NewRecorder()
			Jwt(jwtSvc)(next).ServeHTTP(rec, jwtRequest("Bearer "+token))

			Expect(rec.Code).To(Equal(http.StatusOK))
			got := getClaims()
			Expect(got).NotTo(BeNil())
			Expect(got.UserID).To(Equal(user.UserId))
		})
	})
})
