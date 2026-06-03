package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"akatengu/internal/enums"
	"akatengu/internal/model/db"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/pkg/jwt"
	"akatengu/internal/repos/query"
)

// ─────────────────────────────────────────
// Stubs
// ─────────────────────────────────────────

type stubMerchantRepo struct {
	roleErr error
}

func (s *stubMerchantRepo) Create(_ context.Context, _, _, _ string) (int64, error) { return 0, nil }
func (s *stubMerchantRepo) Get(_ context.Context, _ int64) (*db.Merchant, error)    { return nil, nil }
func (s *stubMerchantRepo) GetUserMerchants(_ context.Context, _ int64) ([]db.Merchant, error) {
	return nil, nil
}
func (s *stubMerchantRepo) GetUserRole(_ context.Context, _, _ int64) (enums.MerchantRoleType, error) {
	return enums.MerchantRoleType{}, s.roleErr
}
func (s *stubMerchantRepo) AddUser(_ context.Context, _, _ int64, _ enums.MerchantRoleType) error {
	return nil
}
func (s *stubMerchantRepo) Update(_ context.Context, _ int64, _, _, _ string) error { return nil }
func (s *stubMerchantRepo) Deactivate(_ context.Context, _ int64) error              { return nil }

var _ query.MerchantRepo = (*stubMerchantRepo)(nil)

// ─────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────

func captureDownstreamMID() (http.HandlerFunc, func() int64) {
	var captured int64
	h := func(w http.ResponseWriter, r *http.Request) {
		captured, _ = r.Context().Value(ctxkey.MerchantID).(int64)
		w.WriteHeader(http.StatusOK)
	}
	return h, func() int64 { return captured }
}

func requestWithContext(ctx context.Context) *http.Request {
	return httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
}

func claimsCtx(merchantID int64) context.Context {
	return context.WithValue(context.Background(), ctxkey.UserClaims,
		&jwt.Claims{UserID: 10, MerchantID: merchantID})
}

// ─────────────────────────────────────────
// Spec runner
// ─────────────────────────────────────────

var _ = Describe("MerchantAuth middleware", func() {
	for _, s := range merchantAuthScenarios {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			stub := &stubMerchantRepo{roleErr: s.stubErr}
			rec := httptest.NewRecorder()
			next, getMID := captureDownstreamMID()

			MerchantAuth(stub)(next).ServeHTTP(rec, requestWithContext(s.ctx))

			Expect(rec.Code).To(Equal(s.wantCode))
			if s.wantBody != "" {
				Expect(rec.Body.String()).To(ContainSubstring(s.wantBody))
			}
			if s.checkMID {
				Expect(getMID()).To(Equal(s.wantMID))
			}
		})
	}
})
