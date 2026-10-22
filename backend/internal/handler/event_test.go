package handler

import (
	"akatengu/internal/bootstrap"
	"akatengu/internal/handler/response"
	dbmodel "akatengu/internal/model/db"
	pkgjwt "akatengu/internal/pkg/jwt"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

// ─────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────

// eventAppendMuxRequest builds a request for the production route with a token
// that has a merchant selected.
func eventAppendMuxRequest(body string) *http.Request {
	token, err := pkgjwt.NewJWT(cfTestJWTSecret).GenerateTokenWithMerchant(
		&dbmodel.User{UserId: cfTestUserID, Username: cfTestUserName}, cfTestMerchantID, "MEMBER")
	Expect(err).NotTo(HaveOccurred())

	req := httptest.NewRequest(http.MethodPost, "/api/event/append", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

// expectEventAppendResult checks the response, the stored events and the
// entry_cf_categories row against the scenario.
func expectEventAppendResult(db *sqlx.DB, rec *httptest.ResponseRecorder, s eventAppendScenario) {
	Expect(rec.Code).To(Equal(s.wantCode), rec.Body.String())
	if s.wantDetail != "" {
		Expect(rec.Header().Get("Content-Type")).To(Equal("application/problem+json"))
		var p response.Problem
		Expect(json.Unmarshal(rec.Body.Bytes(), &p)).To(Succeed())
		Expect(p.Detail).To(ContainSubstring(s.wantDetail))
	}

	var n int
	Expect(db.Get(&n, `SELECT COUNT(*) FROM event_store WHERE event_type = ?`, s.wantEventType)).To(Succeed())
	Expect(n).To(Equal(s.wantEvents))
	if s.wantBFS {
		Expect(db.Get(&n, `SELECT COUNT(*) FROM event_store WHERE event_type = ?
			AND json_extract(metadata, '$.execution.source') = 'api'`, s.wantEventType)).To(Succeed())
		Expect(n).To(Equal(s.wantEvents), "events written without BFS execution metadata")
	}

	Expect(db.Get(&n, `SELECT COUNT(*) FROM entry_cf_categories WHERE entry_uuid = ? AND merchant_id = ?`,
		cfTestDebitEntryUUID, cfTestMerchantID)).To(Succeed())
	Expect(n > 0).To(Equal(s.wantCategory))
}

// ─────────────────────────────────────────
// Spec runner — NewMux POST /api/event/append
// ─────────────────────────────────────────

var _ = Describe("NewMux POST /api/event/append", func() {
	for _, s := range eventAppendScenarios {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			db := newCFTestDB(0, cfTestTxnUUID)
			insertCFAuthFixture(db)

			mux := NewMux(db, bootstrap.Config{JWT: cfTestJWTSecret}, zap.NewNop())

			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, eventAppendMuxRequest(s.body))

			expectEventAppendResult(db, rec, s)
		})
	}
})
