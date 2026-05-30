package projection_test

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services"
	"akatengu/internal/testutil"
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// ─────────────────────────────────────────
// DB query helper
// ─────────────────────────────────────────

type assetScanRow struct {
	PaymentType        string  `db:"payment_type"`
	Status             string  `db:"status"`
	Cost               float64 `db:"cost"`
	DepreciatedPeriods int64   `db:"depreciated_periods"`
	TotalDepreciated   float64 `db:"total_depreciated"`
	DisposalDate       string  `db:"disposal_date"`
	AssetUUID          string  `db:"asset_uuid"`
}

type assetData struct {
	paymentType        string
	status             string
	cost               decimal.Decimal
	depreciatedPeriods int64
	totalDepreciated   decimal.Decimal
	disposalDate       string
	assetUUID          string
}

type projTestDB struct {
	db  *sqlx.DB
	ctx context.Context
}

func (p projTestDB) queryLastAsset() assetData {
	var row assetScanRow
	err := p.db.GetContext(p.ctx, &row,
		`SELECT payment_type, status, cost, depreciated_periods, total_depreciated,
		        COALESCE(disposal_date, '') as disposal_date, asset_uuid
		 FROM fixed_assets WHERE merchant_id=? ORDER BY id DESC LIMIT 1`,
		projTestMID)
	Expect(err).NotTo(HaveOccurred(), "queryLastAsset")
	return assetData{
		paymentType:        row.PaymentType,
		status:             row.Status,
		cost:               decimal.NewFromFloat(row.Cost),
		depreciatedPeriods: row.DepreciatedPeriods,
		totalDepreciated:   decimal.NewFromFloat(row.TotalDepreciated),
		disposalDate:       row.DisposalDate,
		assetUUID:          row.AssetUUID,
	}
}

// ─────────────────────────────────────────
// Service / context helpers
// ─────────────────────────────────────────

const projTestMID int64 = 1

func projTestCtx() context.Context {
	return context.WithValue(context.Background(), ctxkey.MerchantID, projTestMID)
}

func newProjSvc(db *sqlx.DB) *services.EventStoreService {
	uow := event_store.NewUnitOfWork(db)
	q := query.NewQueryRepository(db)
	return services.NewEventStoreService(uow, q)
}

// ─────────────────────────────────────────
// Event appender
// ─────────────────────────────────────────

type projAppender struct {
	svc     *services.EventStoreService
	ctx     context.Context
	aggID   string
	version int64
}

func newProjAppender(svc *services.EventStoreService, ctx context.Context, aggID string) *projAppender {
	return &projAppender{svc: svc, ctx: ctx, aggID: aggID}
}

func (a *projAppender) do(et event_types.EventType, p any) error {
	raw, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	_, err = a.svc.Append(a.ctx, cmd.AppendCmd{
		AggregateType:   enums.AggregateTransaction.Enum(),
		AggregateID:     a.aggID,
		ExpectedVersion: a.version,
		EventType:       et,
		Payload:         raw,
	})
	if err == nil {
		a.version++
	}
	return err
}

// ─────────────────────────────────────────
// DB setup helpers
// ─────────────────────────────────────────

func projInsertLedger(db *sqlx.DB, ctx context.Context, id int64, uuid, typ, acctID string) {
	_, err := db.ExecContext(ctx,
		`INSERT INTO ledger_accounts(ledger_id, merchant_id, account_id, institution, name, type, currency, is_active, version, ledger_uuid)
		 VALUES(?, ?, ?, 'TEST', 'test ledger', ?, 'TWD', 1, 1, ?)`,
		id, projTestMID, acctID, typ, uuid,
	)
	Expect(err).NotTo(HaveOccurred(), "projInsertLedger")
}

func projInsertOpenPeriod(db *sqlx.DB, ctx context.Context, id int64, start string) {
	_, err := db.ExecContext(ctx,
		`INSERT INTO period_closings(closing_id, merchant_id, period_type, period_start, period_end, status, closed_at)
		 VALUES(?, ?, 'MONTHLY', ?, date(?, '+1 month', '-1 day'), 'OPEN', NULL)`,
		id, projTestMID, start, start,
	)
	Expect(err).NotTo(HaveOccurred(), "projInsertOpenPeriod")
}

// ─────────────────────────────────────────
// Spec runner
// ─────────────────────────────────────────

var _ = Describe("FixedAssetProjectionService — applyPurchased", func() {
	for _, s := range buildPurchasedScenarios() {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			db, closeDB := testutil.NewTestDBGinkgo()
			DeferCleanup(closeDB)
			ctx := projTestCtx()
			projInsertOpenPeriod(db, ctx, 1, "2026-05-01")

			svc := newProjSvc(db)
			a := newProjAppender(svc, ctx, "asset-proj-"+s.paymentType)

			var p payload.AssetPurchasedPayload
			if s.paymentType == "CASH" {
				uuid := "test-ledger-uuid-1"
				projInsertLedger(db, ctx, 1, uuid, "BANK_ACCOUNT", "1101-02")
				p = cashPurchasedPayload(uuid)
			} else {
				p = leasePurchasedPayload()
			}

			err := a.do(event_types.EventAssetPurchased.Enum(), p)
			Expect(err).NotTo(HaveOccurred())

			if s.checkDB != nil {
				s.checkDB(projTestDB{db: db, ctx: ctx})
			}
		})
	}
})

var _ = Describe("FixedAssetProjectionService — applyDepreciated", func() {
	for _, s := range buildDepreciatedScenarios() {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			db, closeDB := testutil.NewTestDBGinkgo()
			DeferCleanup(closeDB)
			ctx := projTestCtx()
			projInsertOpenPeriod(db, ctx, 1, "2026-05-01")
			projInsertOpenPeriod(db, ctx, 2, "2026-06-01")

			svc := newProjSvc(db)
			a := newProjAppender(svc, ctx, "asset-depr-"+s.given)

			p := leasePurchasedPayload()
			if s.wantErrContain == "" {
				// normal case: 60-period asset
				err := a.do(event_types.EventAssetPurchased.Enum(), p)
				Expect(err).NotTo(HaveOccurred())
			} else {
				// fully-depreciated case: 1-period asset, depreciated once already
				p.UsefulLifeMonths = 1
				err := a.do(event_types.EventAssetPurchased.Enum(), p)
				Expect(err).NotTo(HaveOccurred())

				assetUUID := projTestDB{db: db, ctx: ctx}.queryLastAsset().assetUUID
				err = a.do(event_types.EventAssetDepreciated.Enum(), depreciatedPayload(assetUUID))
				Expect(err).NotTo(HaveOccurred())
			}

			assetUUID := projTestDB{db: db, ctx: ctx}.queryLastAsset().assetUUID
			err := a.do(event_types.EventAssetDepreciated.Enum(), depreciatedPayload(assetUUID))

			if s.wantErrContain != "" {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(s.wantErrContain))
			} else {
				Expect(err).NotTo(HaveOccurred())
				if s.checkDB != nil {
					s.checkDB(projTestDB{db: db, ctx: ctx})
				}
			}
		})
	}
})

var _ = Describe("FixedAssetProjectionService — applyDisposed", func() {
	for _, s := range buildDisposedScenarios() {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			db, closeDB := testutil.NewTestDBGinkgo()
			DeferCleanup(closeDB)
			ctx := projTestCtx()
			projInsertOpenPeriod(db, ctx, 1, "2026-05-01")

			svc := newProjSvc(db)
			a := newProjAppender(svc, ctx, "asset-disp-"+s.given)

			var assetUUID string
			if s.wantErrContain == "" {
				// 先建立資產
				err := a.do(event_types.EventAssetPurchased.Enum(), leasePurchasedPayload())
				Expect(err).NotTo(HaveOccurred())
				assetUUID = projTestDB{db: db, ctx: ctx}.queryLastAsset().assetUUID
			} else {
				// 使用不存在的 UUID
				assetUUID = "non-existent-uuid"
			}

			err := a.do(event_types.EventAssetDisposed.Enum(), disposedPayload(assetUUID))

			if s.wantErrContain != "" {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(s.wantErrContain))
			} else {
				Expect(err).NotTo(HaveOccurred())
				if s.checkDB != nil {
					s.checkDB(projTestDB{db: db, ctx: ctx})
				}
			}
		})
	}
})
