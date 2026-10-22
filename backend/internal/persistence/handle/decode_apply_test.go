package handle_test

import (
	domainevent "akatengu/internal/domain/event"
	"akatengu/internal/enums"
	"akatengu/internal/kernel/event"
	"akatengu/internal/persistence/handle"
	"akatengu/internal/persistence/repos"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/process"
	"akatengu/internal/repos/query"
	"akatengu/internal/runtime/engine"
	"akatengu/internal/runtime/mediator"
	"akatengu/internal/runtime/safety"
	"akatengu/internal/shared/utils/test"
	"context"
	"reflect"
	"testing"

	"github.com/jmoiron/sqlx"
)

// Fixture identifiers: one rent payment with a debit and a credit entry; the
// accounts come from seeds/accounts.sql.
const (
	decodeTestMerchantID  int64 = 1
	decodeTestUserName          = "r1-registry-tester"
	decodeTestTxnUUID           = "0192f3f1-0000-7000-8000-000000000001"
	decodeTestDebitEntry        = "0192f3f1-0000-7000-8000-000000000011"
	decodeTestCreditEntry       = "0192f3f1-0000-7000-8000-000000000012"
)

type cfCategoryRow struct {
	EntryUUID   string `db:"entry_uuid"`
	CFCategory  string `db:"cf_category"`
	IsConfirmed bool   `db:"is_confirmed"`
	UpdatedBy   string `db:"updated_by"`
}

func cfCategoryRows(t *testing.T, db *sqlx.DB) []cfCategoryRow {
	t.Helper()
	var rows []cfCategoryRow
	err := db.Select(&rows, `SELECT entry_uuid, cf_category, is_confirmed, COALESCE(updated_by, '') AS updated_by
		FROM entry_cf_categories WHERE merchant_id = ? ORDER BY entry_uuid`, decodeTestMerchantID)
	if err != nil {
		t.Fatalf("select entry_cf_categories: %v", err)
	}
	return rows
}

// TestDecodeThenApply checks the R1 acceptance condition: an event read back
// from event_store, decoded through the event registry, can be applied by the
// projector and restores the read model.
func TestDecodeThenApply(t *testing.T) {
	db := test.NewTestDB(t)
	// Every new connection to ":memory:" is an empty database; keep a single one.
	db.SetMaxOpenConns(1)
	ctx := context.WithValue(context.Background(), ctxkey.MerchantID, decodeTestMerchantID)

	for _, stmt := range []struct {
		query string
		args  []any
	}{
		{`INSERT INTO transactions(txn_id, txn_uuid, merchant_id, txn_date, description, total_amount, currency, status, version)
			VALUES(1, ?, ?, '2026-10-01', 'r1 registry rent', 1000, 'TWD', 'ACTIVE', 1)`,
			[]any{decodeTestTxnUUID, decodeTestMerchantID}},
		{`INSERT INTO journal_entries(entry_uuid, txn_uuid, merchant_id, txn_id, account_id, debit, credit)
			VALUES(?, ?, ?, 1, '5101-01', 1000, 0), (?, ?, ?, 1, '1101-02', 0, 1000)`,
			[]any{decodeTestDebitEntry, decodeTestTxnUUID, decodeTestMerchantID, decodeTestCreditEntry, decodeTestTxnUUID, decodeTestMerchantID}},
	} {
		if _, err := db.Exec(stmt.query, stmt.args...); err != nil {
			t.Fatalf("fixture: %v", err)
		}
	}

	// Write the event through BFS, wired the same way as NewMux.
	m := mediator.NewMediator()
	process.Registry(m)
	projectors := handle.NewRegistry()
	projectors.Register(handle.NewCashFlowCategoryProjector(db))
	uow := repos.NewUnitOfWork(db)
	bfs := engine.NewBFSEngine(engine.NewExecutor(m)).
		WithMiddleware(safety.BackpressureScheduler(3), safety.DepthGuard(10), safety.CycleDetector()).
		WithProjector(projectors).
		WithUnitOfWork(uow)

	root := event.NewFullEvent(event.ParseAggregateID(decodeTestTxnUUID), domainevent.TransactionCFCategoryUpdatedPayload{
		TxnUUID: decodeTestTxnUUID,
		Entries: []domainevent.TransactionCFEntryItem{
			{EntryUUID: decodeTestDebitEntry, CFCategory: enums.CashFlowCategoryInvesting.Enum()},
		},
	}, event.EventParams{Source: "api"})
	root.AggregateType = enums.AggregateTransaction.Enum()
	root.MerchantID = decodeTestMerchantID
	root.UpdatedBy = decodeTestUserName
	if _, err := bfs.Run(ctx, root); err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := cfCategoryRows(t, db)
	if len(want) != 1 || want[0].CFCategory != "INVESTING" || !want[0].IsConfirmed || want[0].UpdatedBy != decodeTestUserName {
		t.Fatalf("entry_cf_categories after Run = %+v", want)
	}
	if _, err := db.Exec(`DELETE FROM entry_cf_categories`); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Rebuild only from event_store.
	stored, err := query.NewQueryRepository(db).Event.GetByAggregate(ctx, enums.AggregateTransaction.Enum(), decodeTestTxnUUID)
	if err != nil || len(stored) != 1 {
		t.Fatalf("stored events = %d, err = %v, want 1", len(stored), err)
	}
	row := stored[0]
	payload, err := domainevent.Decode(row.EventType, row.Payload)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	var updatedBy string
	if row.UpdatedBy != nil {
		updatedBy = *row.UpdatedBy
	}
	evt := event.Event{
		Id:            row.EventId,
		Uuid:          row.EventUuid,
		AggregateType: row.AggregateType,
		AggregateUuid: row.AggregateId,
		EventType:     row.EventType,
		Payload:       payload,
		MerchantID:    row.MerchantID,
		UpdatedBy:     updatedBy,
	}
	if err := uow.Do(ctx, func(tx *repos.DbTransaction) error {
		return projectors.ApplyAll(ctx, tx, evt)
	}); err != nil {
		t.Fatalf("ApplyAll: %v", err)
	}

	if got := cfCategoryRows(t, db); !reflect.DeepEqual(got, want) {
		t.Fatalf("entry_cf_categories after apply = %+v, want %+v", got, want)
	}
}
