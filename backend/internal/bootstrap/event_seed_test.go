package bootstrap

import (
	"akatengu/internal/application/command"
	"akatengu/internal/database"
	"akatengu/internal/shared/utils/test"
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"go.uber.org/zap"
)

// TestBootstrapEventSeeds_AccountsTwice: the account seed is written through
// BFS (BFS R2 D11), and running the seeds again does not write it twice.
func TestBootstrapEventSeeds_AccountsTwice(t *testing.T) {
	db := test.NewTestDB(t)
	// Every new connection to ":memory:" is an empty database; keep a single one.
	db.SetMaxOpenConns(1)
	w := command.NewWriter(db, command.NewEngine(db))

	var seeded int
	if err := db.Get(&seeded, `SELECT COUNT(*) FROM accounts WHERE merchant_id = ?`, seedMerchantID); err != nil {
		t.Fatalf("count accounts: %v", err)
	}
	if seeded == 0 {
		t.Fatal("seeds/accounts.sql wrote no account")
	}

	for i := 0; i < 2; i++ {
		if err := BootstrapEventSeeds(context.Background(), db, w, zap.NewNop()); err != nil {
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

// TestBootstrapEventSeeds_ConfigsTwice: the system account and account config
// seeds are written through BFS at the versions the seed counts (BFS R2 D11),
// and running the seeds again does not write them twice.
func TestBootstrapEventSeeds_ConfigsTwice(t *testing.T) {
	ctx := context.Background()
	db := test.NewTestDB(t)
	// Every new connection to ":memory:" is an empty database; keep a single one.
	db.SetMaxOpenConns(1)
	for _, s := range []database.Seeder{
		database.NewSQLFileSeeder("sys_accounts", "sys_accounts.sql"),
		database.NewSQLFileSeeder("account_type_configs", "account_type_configs.sql"),
	} {
		if err := s.Seed(ctx, db.DB); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	countOf := func(table string) int {
		t.Helper()
		var n int
		if err := db.Get(&n, `SELECT COUNT(*) FROM `+table+` WHERE merchant_id = ?`, seedMerchantID); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if n == 0 {
			t.Fatalf("the seeds wrote no %s row", table)
		}
		return n
	}
	sysRows := countOf("sys_accounts")
	ledgerRows := countOf("ledger_account_type_config")
	assetRows := countOf("asset_type_account_config")

	w := command.NewWriter(db, command.NewEngine(db))
	for i := 0; i < 2; i++ {
		if err := BootstrapEventSeeds(ctx, db, w, zap.NewNop()); err != nil {
			t.Fatalf("run %d: %v", i+1, err)
		}
	}

	type eventRow struct {
		AggregateType string `db:"aggregate_type"`
		AggregateID   string `db:"aggregate_id"`
		Version       int64  `db:"aggregate_version"`
		EventType     string `db:"event_type"`
		Source        string `db:"source"`
	}
	var got []eventRow
	if err := db.Select(&got, `SELECT aggregate_type, aggregate_id, aggregate_version, event_type,
		COALESCE(json_extract(metadata, '$.execution.source'), '') AS source
		FROM event_store WHERE aggregate_type IN ('SYS_CONFIG', 'ACCOUNT_CONFIG') AND merchant_id = ?
		ORDER BY aggregate_type DESC, event_id`, seedMerchantID); err != nil {
		t.Fatalf("read events: %v", err)
	}

	var want []eventRow
	for i := 0; i < sysRows; i++ {
		want = append(want, eventRow{"SYS_CONFIG", "1", int64(i + 1), "sys_config.sys_account_updated", "seed"})
	}
	// The ledger type configs come first, then the asset type configs, in one
	// version sequence.
	for i := 0; i < ledgerRows+assetRows; i++ {
		et := "account_config.ledger_type_updated"
		if i >= ledgerRows {
			et = "account_config.asset_type_updated"
		}
		want = append(want, eventRow{"ACCOUNT_CONFIG", "1", int64(i + 1), et, "seed"})
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("events =\n%+v\nwant\n%+v", got, want)
	}
}
