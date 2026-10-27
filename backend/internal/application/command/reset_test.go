package command_test

import (
	"akatengu/internal/application/command"
	"akatengu/internal/persistence/repos"
	"akatengu/internal/pkg/ctxkey"
	"context"
	"reflect"
	"testing"

	"github.com/jmoiron/sqlx"
)

// eventCoreTables are written with the events, not by the projectors; ResetAll
// keeps them, except the projectors' checkpoints (BFS R3 D6, D9).
var eventCoreTables = map[string]bool{
	"event_store": true, "aggregate_versions": true, "snapshots": true, "projection_checkpoints": true,
}

// wantResetScope is the clearing scope of BFS R3 plan 1.1. The test finds the
// scope from what BFS writes; this list only documents it, so a projector that
// writes a new table shows up here.
var wantResetScope = []string{
	"account_closure", "accounts", "asset_type_account_config", "entry_cf_categories", "fixed_asset_categories",
	"ledger_account_type_config", "ledger_accounts", "period_closings", "prepaid_categories", "sys_accounts",
}

// resetAll runs ResetAll of NewEngine's projectors for merchantID in one
// transaction.
func resetAll(db *sqlx.DB, merchantID int64) error {
	ctx := context.WithValue(context.Background(), ctxkey.MerchantID, merchantID)
	return repos.NewUnitOfWork(db).Do(ctx, func(tx *repos.DbTransaction) error {
		return command.NewProjectors(db).ResetAll(ctx, tx)
	})
}

// expectSameRows reports each table whose rows differ.
func expectSameRows(t *testing.T, label string, got, want map[string][]string, tables []string) {
	t.Helper()
	for _, table := range tables {
		if !reflect.DeepEqual(got[table], want[table]) {
			t.Errorf("%s: %s has %d rows, want %d:\n got  %q\n want %q",
				label, table, len(got[table]), len(want[table]), got[table], want[table])
		}
	}
}

// TestResetAll covers BFS R3 段 2 (D1, D9): the clearing scope is the tables
// that differ before and after BFS writes every registered event type for
// merchant 2, and ResetAll returns every other table to that baseline.
func TestResetAll(t *testing.T) {
	db, accounts := newRebuildTestDB(t)
	// A legacy projection's checkpoint of merchant 2, which ResetAll keeps.
	if _, err := db.Exec(`INSERT INTO projection_checkpoints(projection_name, merchant_id, last_event_id)
		VALUES ('TRANSACTION', 2, 0)`); err != nil {
		t.Fatalf("insert legacy checkpoint: %v", err)
	}
	writeFixtureEvents(t, db, 1, accounts)

	tables := merchantTables(t, db)
	baseline := merchantRows(t, db, tables, 2)
	expectAllRegistered(t, writeFixtureEvents(t, db, 2, accounts))
	written := merchantRows(t, db, tables, 2)
	merchant1 := merchantRows(t, db, tables, 1)

	var scope []string
	for _, table := range tables {
		if !eventCoreTables[table] && !reflect.DeepEqual(baseline[table], written[table]) {
			scope = append(scope, table)
		}
	}
	if !reflect.DeepEqual(scope, wantResetScope) {
		t.Errorf("BFS wrote %q, want %q", scope, wantResetScope)
	}
	for _, table := range scope {
		if len(baseline[table]) != 0 {
			t.Errorf("the prerequisite data wrote %s, a table of the projectors", table)
		}
	}

	if err := resetAll(db, 2); err != nil {
		t.Fatalf("ResetAll: %v", err)
	}

	t.Run("情境 1：BFS 寫入前後有差異的資料表，商戶 2 的資料全部清空", func(t *testing.T) {
		got := merchantRows(t, db, tables, 2)
		for _, table := range scope {
			if len(got[table]) != 0 {
				t.Errorf("%s still has %d rows of merchant 2", table, len(got[table]))
			}
		}
	})

	t.Run("情境 2：前置的交易與分錄、其他資料表回到基準，商戶 1 不變", func(t *testing.T) {
		got := merchantRows(t, db, tables, 2)
		want := make(map[string][]string, len(tables))
		for _, table := range tables {
			want[table] = baseline[table]
			if eventCoreTables[table] && table != "projection_checkpoints" {
				want[table] = written[table]
			}
		}
		expectSameRows(t, "merchant 2", got, want, tables)
		expectSameRows(t, "merchant 1", merchantRows(t, db, tables, 1), merchant1, tables)
	})

	t.Run("情境 3：連續執行 ResetAll 兩次", func(t *testing.T) {
		before := merchantRows(t, db, tables, 2)
		if err := resetAll(db, 2); err != nil {
			t.Fatalf("second ResetAll: %v", err)
		}
		expectSameRows(t, "merchant 2", merchantRows(t, db, tables, 2), before, tables)
	})
}
