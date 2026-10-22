package bootstrap

import (
	"akatengu/internal/application/command"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services"
	"akatengu/internal/shared/utils/test"
	"context"
	"encoding/json"
	"testing"

	"go.uber.org/zap"
)

// TestBootstrapEventSeeds_AccountsTwice: the account seed is written through
// BFS (BFS R2 D11), and running the seeds again does not write it twice.
func TestBootstrapEventSeeds_AccountsTwice(t *testing.T) {
	db := test.NewTestDB(t)
	// Every new connection to ":memory:" is an empty database; keep a single one.
	db.SetMaxOpenConns(1)
	es := services.NewEventStoreService(event_store.NewUnitOfWork(db), query.NewQueryRepository(db))
	w := command.NewWriter(db, command.NewEngine(db))

	var seeded int
	if err := db.Get(&seeded, `SELECT COUNT(*) FROM accounts WHERE merchant_id = ?`, seedMerchantID); err != nil {
		t.Fatalf("count accounts: %v", err)
	}
	if seeded == 0 {
		t.Fatal("seeds/accounts.sql wrote no account")
	}

	for i := 0; i < 2; i++ {
		if err := BootstrapEventSeeds(context.Background(), db, es, w, zap.NewNop()); err != nil {
			t.Fatalf("run %d: %v", i+1, err)
		}
	}

	var events []struct {
		AggregateType string          `db:"aggregate_type"`
		AggregateID   string          `db:"aggregate_id"`
		Version       int64           `db:"aggregate_version"`
		Source        *string         `db:"source"`
		Payload       json.RawMessage `db:"payload"`
	}
	if err := db.Select(&events, `SELECT aggregate_type, aggregate_id, aggregate_version,
		json_extract(metadata, '$.execution.source') AS source, payload
		FROM event_store WHERE event_type = 'account.bulk_imported' AND merchant_id = ?`, seedMerchantID); err != nil {
		t.Fatalf("read events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("account.bulk_imported events = %d, want 1", len(events))
	}
	e := events[0]
	if e.AggregateType != "ACCOUNT" || e.AggregateID != "1" || e.Version != 1 {
		t.Errorf("event at (%s, %s) version %d, want (ACCOUNT, 1) version 1", e.AggregateType, e.AggregateID, e.Version)
	}
	if e.Source == nil || *e.Source != "seed" {
		t.Errorf("execution source = %v, want seed (written through BFS)", e.Source)
	}
	var p struct {
		Accounts []json.RawMessage `json:"accounts"`
	}
	if err := json.Unmarshal(e.Payload, &p); err != nil || len(p.Accounts) != seeded {
		t.Errorf("payload has %d accounts (err %v), want the %d seeded", len(p.Accounts), err, seeded)
	}

	var after int
	if err := db.Get(&after, `SELECT COUNT(*) FROM accounts WHERE merchant_id = ?`, seedMerchantID); err != nil {
		t.Fatalf("count accounts: %v", err)
	}
	if after != seeded {
		t.Errorf("accounts = %d after the seeds, want %d", after, seeded)
	}
}
