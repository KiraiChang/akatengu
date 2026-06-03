package handler

import (
	"akatengu/internal/model/db"
	"akatengu/internal/model/request"
	"akatengu/internal/pkg/ctxkey"
	pkgjwt "akatengu/internal/pkg/jwt"
	"akatengu/internal/services"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

// ─────────────────────────────────────────
// Sentinel errors used by scenarios
// ─────────────────────────────────────────

var (
	errMerchantForbiddenStub = services.ErrMerchantForbidden
	errInternalStub          = errors.New("db timeout")
)

// ─────────────────────────────────────────
// Stub service
// ─────────────────────────────────────────

type stubMerchantService struct {
	updateErr    error
	deactivateErr error
}

func (s *stubMerchantService) Create(_ context.Context, _ int64, _ request.CreateMerchant) (int64, error) {
	return 1, nil
}
func (s *stubMerchantService) Select(_ context.Context, _ int64, _ string, _ request.SelectMerchant) (string, error) {
	return "token", nil
}
func (s *stubMerchantService) ListByUser(_ context.Context, _ int64) ([]db.Merchant, error) {
	return nil, nil
}
func (s *stubMerchantService) Update(_ context.Context, _, _ int64, _ request.UpdateMerchant) error {
	return s.updateErr
}
func (s *stubMerchantService) Deactivate(_ context.Context, _, _ int64) error {
	return s.deactivateErr
}

var _ services.MerchantService = (*stubMerchantService)(nil)

// ─────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────

func newTestMerchantHandler(svc services.MerchantService) *merchantHandler {
	return &merchantHandler{svc: svc, logger: zap.NewNop()}
}

func merchantRequest(method, path, body string) *http.Request {
	var bodyReader *strings.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	} else {
		bodyReader = strings.NewReader("")
	}
	req := httptest.NewRequest(method, path, bodyReader)
	ctx := context.WithValue(req.Context(), ctxkey.UserClaims, &pkgjwt.Claims{UserID: 10})
	return req.WithContext(ctx)
}

// ─────────────────────────────────────────
// Spec runner — Update
// ─────────────────────────────────────────

var _ = Describe("merchantHandler.Update", func() {
	for _, s := range updateMerchantScenarios {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			stub := &stubMerchantService{updateErr: s.svcErr}
			h := newTestMerchantHandler(stub)

			// 使用標準 mux 以支援 PathValue
			mux := http.NewServeMux()
			mux.HandleFunc("PUT /merchant/{merchant_id}", h.Update)

			rec := httptest.NewRecorder()
			req := merchantRequest(s.method, s.path, s.body)
			mux.ServeHTTP(rec, req)

			Expect(rec.Code).To(Equal(s.wantCode))
			if s.wantBody != "" {
				Expect(rec.Body.String()).To(ContainSubstring(s.wantBody))
			}
		})
	}
})

// ─────────────────────────────────────────
// Spec runner — Deactivate
// ─────────────────────────────────────────

var _ = Describe("merchantHandler.Deactivate", func() {
	for _, s := range deactivateMerchantScenarios {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			stub := &stubMerchantService{deactivateErr: s.svcErr}
			h := newTestMerchantHandler(stub)

			mux := http.NewServeMux()
			mux.HandleFunc("DELETE /merchant/{merchant_id}", h.Deactivate)

			rec := httptest.NewRecorder()
			req := merchantRequest(s.method, s.path, s.body)
			mux.ServeHTTP(rec, req)

			Expect(rec.Code).To(Equal(s.wantCode))
			if s.wantBody != "" {
				Expect(rec.Body.String()).To(ContainSubstring(s.wantBody))
			}
		})
	}
})
