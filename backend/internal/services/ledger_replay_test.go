package services_test

import (
	"akatengu/internal/application/command"
	domainevent "akatengu/internal/domain/event"
	"akatengu/internal/enums/event_types"
	kevent "akatengu/internal/kernel/event"
	dbmodel "akatengu/internal/model/db"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/shared/utils/test"
	"context"
	"encoding/json"
	"testing"

	"github.com/jmoiron/sqlx"
)

// Merchant 2 has no seeded accounts, so its events are all the replay sees.
const (
	replayMerchantID int64 = 2
	replayAccountID        = "1101-93"
)

// writeReplayEvent writes an event of merchant 2 through BFS, from the payload
// JSON the frontend sends.
func writeReplayEvent(t *testing.T, db *sqlx.DB, eventType event_types.EventType, p map[string]any, expected int64) *dbmodel.EventStore {
	t.Helper()
	ctx := context.WithValue(context.Background(), ctxkey.MerchantID, replayMerchantID)
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	decoded, err := domainevent.Decode(eventType, raw)
	if err != nil {
		t.Fatalf("decode %s: %v", eventType, err)
	}
	evt, err := command.Build(ctx, decoded.(kevent.Payload), expected, kevent.EventParams{Source: "api"})
	if err != nil {
		t.Fatalf("build %s: %v", eventType, err)
	}
	got, err := command.NewWriter(db, command.NewEngine(db)).Append(ctx, evt)
	if err != nil {
		t.Fatalf("append %s: %v", eventType, err)
	}
	return got
}

// TestReplay_TwoLedgerAccounts covers BFS R3 段 1 scenario 8. The BFS payload
// of ledger_account.created records ledger_id, but the legacy projection keeps
// its own CreateLedgerAccount query, so the legacy Replay still lets SQLite
// assign the IDs and writes both ledger accounts back (BFS R3 D3).
func TestReplay_TwoLedgerAccounts(t *testing.T) {
	db := test.NewTestDB(t)
	// Every new connection to ":memory:" is an empty database; keep a single one.
	db.SetMaxOpenConns(1)

	writeReplayEvent(t, db, event_types.EventAccountCreated.Enum(), map[string]any{
		"account_id": replayAccountID, "name": "R3 重播科目", "type": "ASSET", "normal_balance": "DEBIT",
		"currency": "TWD", "is_active": true,
	}, 0)
	ledger := func(name string) map[string]any {
		return map[string]any{
			"account_id": replayAccountID, "institution": "R3 重播銀行", "name": name,
			"type": "BANK_ACCOUNT", "currency": "TWD", "is_active": true,
		}
	}
	first := writeReplayEvent(t, db, event_types.EventLedgerAccountCreated.Enum(), ledger("R3 重播帳戶一"), 1)
	second := writeReplayEvent(t, db, event_types.EventLedgerAccountCreated.Enum(), ledger("R3 重播帳戶二"), 2)

	ctx := context.WithValue(context.Background(), ctxkey.MerchantID, replayMerchantID)
	if _, err := newSvc(db).Replay(ctx, 0, nil); err != nil {
		t.Fatalf("legacy Replay: %v", err)
	}

	type ledgerRow struct {
		LedgerUUID string `db:"ledger_uuid"`
		Name       string `db:"name"`
	}
	var rows []ledgerRow
	if err := db.Select(&rows, `SELECT ledger_uuid, name FROM ledger_accounts WHERE merchant_id = ? ORDER BY ledger_id`,
		replayMerchantID); err != nil {
		t.Fatalf("read ledger accounts: %v", err)
	}
	want := []ledgerRow{{first.EventUuid, "R3 重播帳戶一"}, {second.EventUuid, "R3 重播帳戶二"}}
	if len(rows) != len(want) || rows[0] != want[0] || rows[1] != want[1] {
		t.Errorf("ledger accounts after Replay = %+v, want %+v", rows, want)
	}
}
