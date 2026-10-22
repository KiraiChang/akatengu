package services_test

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/services"
	"akatengu/internal/shared/utils/test"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
)

// Fixture identifiers: one rent payment with a debit and a credit entry. The
// accounts come from seeds/accounts.sql; 1101-92 is not in the seeds.
const (
	migratedTxnUUID      = "0192f3f2-0000-7000-8000-000000000001"
	migratedDebitEntry   = "0192f3f2-0000-7000-8000-000000000011"
	migratedCreditEntry  = "0192f3f2-0000-7000-8000-000000000012"
	migratedNewAccountID = "1101-92" // written by account.created
)

func newMigratedTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db := test.NewTestDB(t)
	// Every new connection to ":memory:" is an empty database; keep a single one.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`INSERT INTO transactions(txn_id, txn_uuid, merchant_id, txn_date, description, total_amount, currency, status, version)
		VALUES(1, ?, ?, '2026-10-01', 'f7 migrated rent', 1000, 'TWD', 'ACTIVE', 1)`, migratedTxnUUID, testMID); err != nil {
		t.Fatalf("insert transaction: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO journal_entries(entry_uuid, txn_uuid, merchant_id, txn_id, account_id, debit, credit)
		VALUES(?, ?, ?, 1, '5101-01', 1000, 0), (?, ?, ?, 1, '1101-02', 0, 1000)`,
		migratedDebitEntry, migratedTxnUUID, testMID, migratedCreditEntry, migratedTxnUUID, testMID); err != nil {
		t.Fatalf("insert journal entries: %v", err)
	}
	return db
}

// appendState counts the rows a write through the legacy path touches.
type appendState struct {
	Events     int `db:"events"`
	Versions   int `db:"versions"`
	CFCategory int `db:"cf_category"`
}

func appendStateOf(t *testing.T, db *sqlx.DB) appendState {
	t.Helper()
	var s appendState
	if err := db.Get(&s, `SELECT
		(SELECT COUNT(*) FROM event_store) AS events,
		(SELECT COUNT(*) FROM aggregate_versions) AS versions,
		(SELECT COUNT(*) FROM entry_cf_categories) AS cf_category`); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	return s
}

func mustMarshal(t *testing.T, v any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return raw
}

func TestAppend_MigratedEventType(t *testing.T) {
	test.InitEnums()
	cfType := event_types.EventTransactionCFCategoryUpdated.Enum()
	// Accepted by the legacy pipeline: without the check it is written.
	cfPayload := mustMarshal(t, payload.TransactionCFCategoryUpdatedPayload{
		TxnUUID: migratedTxnUUID,
		Entries: []payload.TransactionCFEntryItem{
			{EntryUUID: migratedDebitEntry, CFCategory: enums.CashFlowCategoryInvesting.Enum()},
		},
	})
	parentID := "1101"
	accountPayload := mustMarshal(t, payload.AccountCreatePayload{
		AccountId:     migratedNewAccountID,
		ParentId:      &parentID,
		Name:          "F7 對照科目",
		Type:          enums.AccountAsset.Enum(),
		NormalBalance: enums.BalanceDebit.Enum(),
		Currency:      "TWD",
		IsActive:      true,
	})

	cases := []struct {
		name    string
		cmd     cmd.AppendCmd
		wantErr error // nil = the event is written
	}{
		{
			name: "已遷移的事件以交易 UUID 為 aggregate",
			cmd: cmd.AppendCmd{AggregateType: enums.AggregateTransaction.Enum(), AggregateID: migratedTxnUUID,
				EventType: cfType, Payload: cfPayload},
			wantErr: services.ErrEventTypeMigrated,
		},
		{
			name: "已遷移的事件寫到交易 UUID 以外的 aggregate",
			cmd: cmd.AppendCmd{AggregateType: enums.AggregateTransaction.Enum(), AggregateID: "",
				EventType: cfType, Payload: cfPayload},
			wantErr: services.ErrEventTypeMigrated,
		},
		{
			name: "對照：未遷移的事件照常寫入",
			cmd: cmd.AppendCmd{AggregateType: enums.AggregateAccount.Enum(), AggregateID: migratedNewAccountID,
				EventType: event_types.EventAccountCreated.Enum(), Payload: accountPayload},
		},
		{
			name: "已遷移的事件搭配舊 pipeline 會拒絕的 payload",
			cmd: cmd.AppendCmd{AggregateType: enums.AggregateTransaction.Enum(), AggregateID: migratedTxnUUID,
				EventType: cfType, Payload: json.RawMessage(`{}`)},
			wantErr: services.ErrEventTypeMigrated,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := newMigratedTestDB(t)
			before := appendStateOf(t, db)

			_, err := newSvc(db).Append(testCtx(), tc.cmd)

			after := appendStateOf(t, db)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				if after != before {
					t.Fatalf("rows changed: before %+v, after %+v", before, after)
				}
				return
			}
			if err != nil {
				t.Fatalf("Append: %v", err)
			}
			if after.Events != before.Events+1 {
				t.Fatalf("event_store rows = %d, want %d", after.Events, before.Events+1)
			}
		})
	}
}
