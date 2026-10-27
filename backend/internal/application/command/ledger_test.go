package command_test

import (
	"akatengu/internal/application/command"
	domainevent "akatengu/internal/domain/event"
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	kevent "akatengu/internal/kernel/event"
	dbmodel "akatengu/internal/model/db"
	"akatengu/internal/model/request/cmd"
	"encoding/json"
	"testing"

	"github.com/jmoiron/sqlx"
)

// cmdLedgerAccountID is a leaf account in seeds/accounts.sql.
const cmdLedgerAccountID = "1101-02"

// ledgerCreatedRequest is the generic API request for a new ledger account, as
// frontside/src/api/ledger.ts sends it: version is the ('ACCOUNT', account_id)
// version the frontend reads first, and extra is merged into the payload.
func ledgerCreatedRequest(t *testing.T, version int64, name string, extra map[string]any) cmd.AppendCmd {
	t.Helper()
	p := map[string]any{
		"account_id": cmdLedgerAccountID, "institution": "R3 段 1 銀行", "name": name,
		"type": "BANK_ACCOUNT", "currency": "TWD", "is_active": true,
	}
	for k, v := range extra {
		p[k] = v
	}
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return cmd.AppendCmd{AggregateType: enums.AggregateAccount.Enum(), AggregateID: cmdLedgerAccountID,
		ExpectedVersion: version, EventType: event_types.EventLedgerAccountCreated.Enum(), Payload: raw}
}

// createLedger sends ledgerCreatedRequest for merchantID through the dispatch
// of the generic API. The handler does not check the account, so a merchant
// without its own accounts can still hold a ledger_id.
func createLedger(t *testing.T, db *sqlx.DB, merchantID int64, name string, extra map[string]any) *dbmodel.EventStore {
	t.Helper()
	var version int64
	if err := db.Get(&version, `SELECT COALESCE(MAX(current_version), 0) FROM aggregate_versions
		WHERE aggregate_type = 'ACCOUNT' AND aggregate_id = ? AND merchant_id = ?`, cmdLedgerAccountID, merchantID); err != nil {
		t.Fatalf("read version: %v", err)
	}
	got, handled, err := command.NewWriter(db, command.NewEngine(db)).DispatchLegacy(cmdCtx(merchantID),
		ledgerCreatedRequest(t, version, name, extra), kevent.EventParams{Source: "api"})
	if !handled || err != nil {
		t.Fatalf("create ledger %q of merchant %d: handled = %v, err = %v", name, merchantID, handled, err)
	}
	return got
}

type ledgerRow struct {
	LedgerID   int64  `db:"ledger_id"`
	MerchantID int64  `db:"merchant_id"`
	Name       string `db:"name"`
}

// TestDispatchLegacy_LedgerCreatedRecordsID covers BFS R3 D3: the handler
// allocates the ledger_id in the Run's transaction and the event records it,
// so the read model and a rebuild use the same ID (ISSUE-031).
func TestDispatchLegacy_LedgerCreatedRecordsID(t *testing.T) {
	type ledger struct {
		merchantID int64
		name       string
		extra      map[string]any
	}
	tests := []struct {
		name   string
		before []ledger // created first, in order
		create ledger
		wantID int64
	}{
		{
			name:   "商戶 1 建立第一個明細帳戶",
			create: ledger{1, "R3 第一個帳戶", nil},
			wantID: 1,
		},
		{
			name:   "商戶 1 已有 ledger_id 1、商戶 2 已有 2，商戶 1 再建立一個",
			before: []ledger{{1, "R3 商戶 1 帳戶", nil}, {2, "R3 商戶 2 帳戶", nil}},
			create: ledger{1, "R3 商戶 1 第二個帳戶", nil},
			wantID: 3,
		},
		{
			name:   "送出的 payload 帶 ledger_id 99",
			create: ledger{1, "R3 帶 ID 的帳戶", map[string]any{"ledger_id": 99}},
			wantID: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newCommandTestDB(t)
			for _, l := range tt.before {
				createLedger(t, db, l.merchantID, l.name, l.extra)
			}

			got := createLedger(t, db, tt.create.merchantID, tt.create.name, tt.create.extra)

			var p domainevent.LedgerAccountCreatePayload
			if err := json.Unmarshal(got.Payload, &p); err != nil {
				t.Fatalf("decode payload %s: %v", got.Payload, err)
			}
			if p.LedgerId != tt.wantID {
				t.Errorf("event ledger_id = %d, want %d", p.LedgerId, tt.wantID)
			}
			var row ledgerRow
			if err := db.Get(&row, `SELECT ledger_id, merchant_id, name FROM ledger_accounts WHERE ledger_uuid = ?`,
				got.EventUuid); err != nil {
				t.Fatalf("read the ledger account named after the event: %v", err)
			}
			if want := (ledgerRow{tt.wantID, tt.create.merchantID, tt.create.name}); row != want {
				t.Errorf("ledger account = %+v, want %+v", row, want)
			}
			var rows int
			if err := db.Get(&rows, `SELECT COUNT(*) FROM ledger_accounts`); err != nil {
				t.Fatalf("count: %v", err)
			}
			if rows != len(tt.before)+1 {
				t.Errorf("ledger accounts = %d, want %d", rows, len(tt.before)+1)
			}
		})
	}
}
