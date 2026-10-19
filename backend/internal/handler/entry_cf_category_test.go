package handler

import (
	"akatengu/internal/bootstrap"
	dbmodel "akatengu/internal/model/db"
	"akatengu/internal/persistence/handle"
	"akatengu/internal/persistence/repos"
	"akatengu/internal/pkg/ctxkey"
	pkgjwt "akatengu/internal/pkg/jwt"
	"akatengu/internal/process"
	"akatengu/internal/runtime/engine"
	"akatengu/internal/runtime/mediator"
	"akatengu/internal/runtime/safety"
	"akatengu/internal/shared/utils/test"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

const (
	cfTestMerchantID int64 = 1
	cfTestUserID     int64 = 1
	cfTestUserName         = "cf-tester"
	cfTestJWTSecret        = "cf-test-secret"
)

// ─────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────

// newCFTestDB opens a migrated in-memory DB with the transaction fixture.
func newCFTestDB(existingVersion int64, txnUUID string) *sqlx.DB {
	db, closeDB := test.NewTestDBGinkgo()
	DeferCleanup(closeDB)
	// Every new connection to ":memory:" is an empty database; keep a single one.
	db.SetMaxOpenConns(1)

	insertCFFixture(db)
	if existingVersion > 0 {
		insertCFAggregateVersion(db, txnUUID, existingVersion)
	}
	return db
}

// newCFTestEngine wires the BFS engine the same way as NewMux.
func newCFTestEngine(db *sqlx.DB) *engine.BFSEngine {
	m := mediator.NewMediator().WithQuery(db)
	process.Registry(m)

	projector := handle.NewRegistry()
	projector.Register(handle.NewCashFlowCategoryProjector(db))

	return engine.NewBFSEngine(engine.NewExecutor(m)).
		WithMiddleware(safety.BackpressureScheduler(3), safety.DepthGuard(10), safety.CycleDetector()).
		WithProjector(projector).
		WithUnitOfWork(repos.NewUnitOfWork(db))
}

// insertCFFixture stores two explicit test transactions: a rent payment with a
// debit and a credit entry, and another transaction whose entry must be rejected.
func insertCFFixture(db *sqlx.DB) {
	exec := func(query string, args ...any) {
		_, err := db.Exec(query, args...)
		Expect(err).NotTo(HaveOccurred())
	}
	insertTxn := `INSERT INTO transactions(txn_id, txn_uuid, merchant_id, txn_date, description, total_amount, currency, status, version)
		VALUES(?, ?, ?, '2026-10-01', ?, 1000, 'TWD', 'ACTIVE', 1)`
	insertEntry := `INSERT INTO journal_entries(entry_uuid, txn_uuid, merchant_id, txn_id, account_id, debit, credit)
		VALUES(?, ?, ?, ?, ?, ?, ?)`

	exec(insertTxn, 1, cfTestTxnUUID, cfTestMerchantID, "cf test rent")
	exec(insertEntry, cfTestDebitEntryUUID, cfTestTxnUUID, cfTestMerchantID, 1, "5101-01", 1000, 0)
	exec(insertEntry, cfTestCreditEntryUUID, cfTestTxnUUID, cfTestMerchantID, 1, "1101-02", 0, 1000)

	exec(insertTxn, 2, cfOtherTxnUUID, cfTestMerchantID, "cf test other")
	exec(insertEntry, cfOtherEntryUUID, cfOtherTxnUUID, cfTestMerchantID, 2, "5101-01", 1000, 0)
}

// insertCFAuthFixture stores the user and merchant membership required by the
// Jwt and MerchantAuth middlewares.
func insertCFAuthFixture(db *sqlx.DB) {
	exec := func(query string, args ...any) {
		_, err := db.Exec(query, args...)
		Expect(err).NotTo(HaveOccurred())
	}
	exec(`INSERT INTO users(user_id, username, password, status) VALUES(?, ?, 'unused', 'ACTIVE')`, cfTestUserID, cfTestUserName)
	exec(`INSERT INTO merchants(merchant_id, name, display_name) VALUES(?, 'cf test', 'cf test')`, cfTestMerchantID)
	exec(`INSERT INTO user_merchants(user_id, merchant_id, role) VALUES(?, ?, 'MEMBER')`, cfTestUserID, cfTestMerchantID)
}

func insertCFAggregateVersion(db *sqlx.DB, txnUUID string, version int64) {
	_, err := db.Exec(`INSERT INTO aggregate_versions(aggregate_type, aggregate_id, merchant_id, current_version)
		VALUES('TRANSACTION', ?, ?, ?)`, txnUUID, cfTestMerchantID, version)
	Expect(err).NotTo(HaveOccurred())
}

func cfCategoryRequest(txnUUID, body string) *http.Request {
	req := httptest.NewRequest(http.MethodPut, "/txn/"+txnUUID+"/cf-category", strings.NewReader(body))
	ctx := context.WithValue(req.Context(), ctxkey.MerchantID, cfTestMerchantID)
	ctx = context.WithValue(ctx, ctxkey.UserClaims, &pkgjwt.Claims{UserName: cfTestUserName})
	return req.WithContext(ctx)
}

// cfMuxRequest builds a request for the production route with a token that has
// a merchant selected.
func cfMuxRequest(txnUUID, body string) *http.Request {
	token, err := pkgjwt.NewJWT(cfTestJWTSecret).GenerateTokenWithMerchant(
		&dbmodel.User{UserId: cfTestUserID, Username: cfTestUserName}, cfTestMerchantID, "MEMBER")
	Expect(err).NotTo(HaveOccurred())

	req := httptest.NewRequest(http.MethodPut, "/api/txn/"+txnUUID+"/cf-category", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func countCFTxnEvents(db *sqlx.DB, txnUUID string) int {
	var n int
	err := db.Get(&n, `SELECT COUNT(*) FROM event_store WHERE aggregate_type = 'TRANSACTION' AND aggregate_id = ?`, txnUUID)
	Expect(err).NotTo(HaveOccurred())
	return n
}

// expectCFCategoryResult checks the response, the stored event and the
// entry_cf_categories row against the scenario.
func expectCFCategoryResult(db *sqlx.DB, rec *httptest.ResponseRecorder, s cfCategoryScenario) {
	Expect(rec.Code).To(Equal(s.wantCode), rec.Body.String())
	Expect(countCFTxnEvents(db, s.txnUUID)).To(Equal(s.wantEventCount))

	if s.wantVersion > 0 {
		var evt struct {
			EventType        string `db:"event_type"`
			MerchantID       int64  `db:"merchant_id"`
			AggregateVersion int64  `db:"aggregate_version"`
			UpdatedBy        string `db:"updated_by"`
		}
		err := db.Get(&evt, `SELECT event_type, merchant_id, aggregate_version, COALESCE(updated_by, '') AS updated_by
			FROM event_store WHERE aggregate_type = 'TRANSACTION' AND aggregate_id = ?`, s.txnUUID)
		Expect(err).NotTo(HaveOccurred())
		Expect(evt.EventType).To(Equal("transaction.cf_category_updated"))
		Expect(evt.MerchantID).To(Equal(cfTestMerchantID))
		Expect(evt.AggregateVersion).To(Equal(s.wantVersion))
		Expect(evt.UpdatedBy).To(Equal(cfTestUserName))
	}

	var cf struct {
		Category    string `db:"cf_category"`
		IsConfirmed bool   `db:"is_confirmed"`
	}
	err := db.Get(&cf, `SELECT cf_category, is_confirmed FROM entry_cf_categories WHERE entry_uuid = ? AND merchant_id = ?`,
		cfTestDebitEntryUUID, cfTestMerchantID)
	if s.wantCategory == "" {
		Expect(err).To(MatchError(sql.ErrNoRows))
	} else {
		Expect(err).NotTo(HaveOccurred())
		Expect(cf.Category).To(Equal(s.wantCategory))
		Expect(cf.IsConfirmed).To(BeTrue())
	}
}

// ─────────────────────────────────────────
// Spec runner — UpdateCFCategory
// ─────────────────────────────────────────

var _ = Describe("entryCFCategoryHandler.UpdateCFCategory", func() {
	for _, s := range updateCFCategoryScenarios {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			db := newCFTestDB(s.existingVersion, s.txnUUID)

			h := newEntryCFCategoryHandler(db, nil, zap.NewNop(), newCFTestEngine(db))
			mux := http.NewServeMux()
			mux.HandleFunc("PUT /txn/{txn_uuid}/cf-category", h.UpdateCFCategory)

			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, cfCategoryRequest(s.txnUUID, s.body))

			expectCFCategoryResult(db, rec, s)
		})
	}
})

// ─────────────────────────────────────────
// Spec runner — NewMux PUT /api/txn/{txn_uuid}/cf-category
// ─────────────────────────────────────────

var _ = Describe("NewMux PUT /api/txn/{txn_uuid}/cf-category", func() {
	for _, s := range newMuxCFCategoryScenarios {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			db := newCFTestDB(s.existingVersion, s.txnUUID)
			insertCFAuthFixture(db)

			mux := NewMux(db, bootstrap.Config{JWT: cfTestJWTSecret}, zap.NewNop())

			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, cfMuxRequest(s.txnUUID, s.body))

			expectCFCategoryResult(db, rec, s)
		})
	}
})
