package handler

import (
	"akatengu/internal/application/command"
	"akatengu/internal/bootstrap"
	dbmodel "akatengu/internal/model/db"
	"akatengu/internal/pkg/ctxkey"
	pkgjwt "akatengu/internal/pkg/jwt"
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

// ─────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────

// newSettingTestDB opens a migrated in-memory DB. It has the account seeds but
// no config or sys_accounts rows, so the first write inserts them.
func newSettingTestDB(existingVersion int64) *sqlx.DB {
	db, closeDB := test.NewTestDBGinkgo()
	DeferCleanup(closeDB)
	// Every new connection to ":memory:" is an empty database; keep a single one.
	db.SetMaxOpenConns(1)

	if existingVersion > 0 {
		_, err := db.Exec(`INSERT INTO aggregate_versions(aggregate_type, aggregate_id, merchant_id, current_version)
			VALUES('ACCOUNT_CONFIG', '1', ?, ?)`, cfTestMerchantID, existingVersion)
		Expect(err).NotTo(HaveOccurred())
	}
	return db
}

// settingHandlerRequest builds a request for the handler with the merchant and
// user already in the context, as the middlewares would leave them.
func settingHandlerRequest(r settingRequest) *http.Request {
	req := httptest.NewRequest(r.method, r.path, strings.NewReader(r.body))
	ctx := context.WithValue(req.Context(), ctxkey.MerchantID, cfTestMerchantID)
	ctx = context.WithValue(ctx, ctxkey.UserClaims, &pkgjwt.Claims{UserName: cfTestUserName})
	return req.WithContext(ctx)
}

// settingMuxRequest builds a request for the production route with a token that
// has a merchant selected.
func settingMuxRequest(r settingRequest) *http.Request {
	token, err := pkgjwt.NewJWT(cfTestJWTSecret).GenerateTokenWithMerchant(
		&dbmodel.User{UserId: cfTestUserID, Username: cfTestUserName}, cfTestMerchantID, "MEMBER")
	Expect(err).NotTo(HaveOccurred())

	req := httptest.NewRequest(r.method, "/api"+r.path, strings.NewReader(r.body))
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

// expectSettingResult checks every stored event and the config read models
// against the scenario.
func expectSettingResult(db *sqlx.DB, s settingScenario) {
	var events []struct {
		AggregateType    string `db:"aggregate_type"`
		AggregateID      string `db:"aggregate_id"`
		AggregateVersion int64  `db:"aggregate_version"`
		EventType        string `db:"event_type"`
		Payload          string `db:"payload"`
		MerchantID       int64  `db:"merchant_id"`
		UpdatedBy        string `db:"updated_by"`
	}
	err := db.Select(&events, `SELECT aggregate_type, aggregate_id, aggregate_version, event_type, payload,
		merchant_id, COALESCE(updated_by, '') AS updated_by FROM event_store ORDER BY event_id`)
	Expect(err).NotTo(HaveOccurred())
	Expect(events).To(HaveLen(len(s.wantEvents)))
	for i, want := range s.wantEvents {
		got := events[i]
		Expect(got.AggregateType).To(Equal(want.aggregateType))
		Expect(got.AggregateID).To(Equal("1"))
		Expect(got.AggregateVersion).To(Equal(want.version))
		Expect(got.EventType).To(Equal(want.eventType))
		Expect(got.Payload).To(MatchJSON(want.payload))
		Expect(got.MerchantID).To(Equal(cfTestMerchantID))
		Expect(got.UpdatedBy).To(Equal(cfTestUserName))
	}

	for typ, want := range s.wantLedgerTypes {
		var accountID string
		err := db.Get(&accountID, `SELECT account_id FROM ledger_account_type_config WHERE merchant_id = ? AND type = ?`,
			cfTestMerchantID, typ)
		if want == "" {
			Expect(err).To(MatchError(sql.ErrNoRows))
		} else {
			Expect(err).NotTo(HaveOccurred())
			Expect(accountID).To(Equal(want))
		}
	}

	for assetType, want := range s.wantAssetTypes {
		var row assetTypeConfigRow
		err := db.Get(&row, `SELECT realized_gain_account_id, realized_loss_account_id, unrealized_gain_account_id,
			unrealized_loss_account_id, oci_account_id, fee_account_id, tax_account_id, account_id
			FROM asset_type_account_config WHERE merchant_id = ? AND asset_type = ?`, cfTestMerchantID, assetType)
		if want == nil {
			Expect(err).To(MatchError(sql.ErrNoRows))
		} else {
			Expect(err).NotTo(HaveOccurred())
			Expect(row).To(Equal(*want))
		}
	}

	for sysCode, want := range s.wantSysAccounts {
		var row sysAccountRow
		err := db.Get(&row, `SELECT description, account_id FROM sys_accounts WHERE merchant_id = ? AND sys_code = ?`,
			cfTestMerchantID, sysCode)
		if want == nil {
			Expect(err).To(MatchError(sql.ErrNoRows))
		} else {
			Expect(err).NotTo(HaveOccurred())
			Expect(row).To(Equal(*want))
		}
	}
}

// ─────────────────────────────────────────
// Spec runner — config update handlers
// ─────────────────────────────────────────

var _ = Describe("settingHandler and sysHandler config updates", func() {
	for _, s := range settingScenarios {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			db := newSettingTestDB(s.existingVersion)
			w := command.NewWriter(db, command.NewEngine(db))

			setting := newSettingHandler(db, w, zap.NewNop())
			sys := newSysHandler(db, w, zap.NewNop())
			mux := http.NewServeMux()
			mux.HandleFunc("PUT /setting/ledger-account-type/{type}", setting.UpdateLedgerAccountTypeConfig)
			mux.HandleFunc("PUT /setting/asset-type/{type}", setting.UpdateAssetTypeAccountConfig)
			mux.HandleFunc("POST /sys/account", sys.UpdateSysAccount)

			for _, r := range s.requests {
				rec := httptest.NewRecorder()
				mux.ServeHTTP(rec, settingHandlerRequest(r))
				Expect(rec.Code).To(Equal(s.wantCode), rec.Body.String())
			}

			expectSettingResult(db, s)
		})
	}
})

// ─────────────────────────────────────────
// Spec runner — NewMux config update routes
// ─────────────────────────────────────────

var _ = Describe("NewMux config update routes", func() {
	for _, s := range newMuxSettingScenarios {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			db := newSettingTestDB(s.existingVersion)
			insertCFAuthFixture(db)

			mux := NewMux(db, bootstrap.Config{JWT: cfTestJWTSecret}, zap.NewNop())

			for _, r := range s.requests {
				rec := httptest.NewRecorder()
				mux.ServeHTTP(rec, settingMuxRequest(r))
				Expect(rec.Code).To(Equal(s.wantCode), rec.Body.String())
			}

			expectSettingResult(db, s)
		})
	}
})
