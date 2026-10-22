package database_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"akatengu/internal/database"

	_ "modernc.org/sqlite"
)

// The last migration before period_closings is rebuilt with the merchant in its
// unique key (20261022001, backend ISSUE-030).
const beforePeriodClosingRebuild int64 = 20261021001

// Fixture: periods of two merchants that the old (period_type, period_start)
// key allows, one per status, and one row per snapshot table referencing the
// closed period. The counter (5) is ahead of the largest closing_id (3), as
// after the latest two periods were deleted.
var periodClosingFixture = []string{
	`INSERT INTO period_closings
		(closing_id, merchant_id, closing_uuid, period_type, period_start, period_end, status,
		 snapshot, closed_at, note, reopen_at, reopen_reason, updated_by, updated_at)
	VALUES
		(1, 1, '0192f3f2-0000-7000-8000-000000000901', 'MONTHLY', '2026-09-01', '2026-09-30', 'CLOSED',
		 '{"accounts":[]}', '2026-10-01 09:00:00', 'f9 closed', NULL, NULL, 'f9-user', '2026-10-01 09:00:00'),
		(2, 1, '0192f3f2-0000-7000-8000-000000000902', 'MONTHLY', '2026-10-01', '2026-10-31', 'REOPENED',
		 NULL, '2026-11-01 09:00:00', NULL, '2026-11-02 09:00:00', 'f9 reopen', 'f9-user', '2026-11-02 09:00:00'),
		(3, 2, '0192f3f2-0000-7000-8000-000000000903', 'ANNUAL', '2026-01-01', '2026-12-31', 'OPEN',
		 NULL, NULL, NULL, NULL, NULL, 'f9-user', '2026-01-01 09:00:00')`,
	`UPDATE sqlite_sequence SET seq = 5 WHERE name = 'period_closings'`,
	// account_id and ledger_id are not looked up: the rebuild only touches closing_id references.
	`INSERT INTO account_balance_snapshots (closing_id, merchant_id, account_id, debit_total, credit_total)
	VALUES (1, 1, '1101-01', 1000, 0)`,
	`INSERT INTO ledger_account_balance_snapshots (closing_id, merchant_id, ledger_id, debit_total, credit_total)
	VALUES (1, 1, 1, 1000, 0)`,
}

func newPreRebuildDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// Every new connection to ":memory:" is an empty database, and
	// PRAGMA foreign_keys applies per connection; keep a single one.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })

	if err := database.MigrateUpTo(context.Background(), db, beforePeriodClosingRebuild); err != nil {
		t.Fatalf("migrate up to %d: %v", beforePeriodClosingRebuild, err)
	}
	// Written with foreign keys off: account_balance_snapshots cannot be
	// written with them on (backend ISSUE-033).
	for _, s := range periodClosingFixture {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("fixture: %v", err)
		}
	}
	return db
}

func rowsOf(t *testing.T, db *sqlx.DB, q string) []map[string]any {
	t.Helper()
	rows, err := db.Queryx(q)
	if err != nil {
		t.Fatalf("query %q: %v", q, err)
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		row := map[string]any{}
		if err := rows.MapScan(row); err != nil {
			t.Fatalf("scan %q: %v", q, err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows %q: %v", q, err)
	}
	return out
}

func stringsOf(t *testing.T, db *sqlx.DB, q string, args ...any) []string {
	t.Helper()
	var out []string
	if err := db.Select(&out, q, args...); err != nil {
		t.Fatalf("query %q: %v", q, err)
	}
	return out
}

func TestPeriodClosingMerchantUnique_Migration(t *testing.T) {
	// Rows and columns the rebuild must keep as they were.
	unchanged := map[string]string{
		"period_closings":                  `SELECT * FROM period_closings ORDER BY closing_id`,
		"period_closings columns":          `SELECT cid, name, type, "notnull", dflt_value, pk FROM pragma_table_info('period_closings') ORDER BY cid`,
		"account_balance_snapshots":        `SELECT * FROM account_balance_snapshots ORDER BY id`,
		"ledger_account_balance_snapshots": `SELECT * FROM ledger_account_balance_snapshots ORDER BY id`,
	}
	const (
		indexesQ = `SELECT name FROM sqlite_master
			WHERE type = 'index' AND tbl_name = 'period_closings' AND sql IS NOT NULL ORDER BY name`
		// One "col,col,..." string per UNIQUE constraint declared on the table.
		uniqueQ = `SELECT group_concat(name, ',') FROM (
				SELECT il.name AS idx, ii.name FROM pragma_index_list('period_closings') il
				JOIN pragma_index_info(il.name) ii
				WHERE il."unique" = 1 AND il.origin = 'u' ORDER BY il.name, ii.seqno)
			GROUP BY idx`
	)
	snapshotTables := []string{"account_balance_snapshots", "ledger_account_balance_snapshots"}
	wantIndexes := []string{"idx_period_closings_end", "idx_period_closings_merchant", "idx_period_closings_status"}
	wantUnique := []string{"merchant_id,period_type,period_start"}
	const wantNextClosingID int64 = 6 // the fixture counter + 1

	cases := []struct {
		name        string
		foreignKeys bool
	}{
		{name: "外鍵關閉（與目前的執行環境相同）"},
		{name: "外鍵開啟", foreignKeys: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := newPreRebuildDB(t)
			before := map[string][]map[string]any{}
			for name, q := range unchanged {
				before[name] = rowsOf(t, db, q)
			}
			if tc.foreignKeys {
				if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
					t.Fatalf("enable foreign keys: %v", err)
				}
				var on int
				if err := db.Get(&on, `PRAGMA foreign_keys`); err != nil || on != 1 {
					t.Fatalf("foreign_keys = %d (err %v), want 1", on, err)
				}
			}

			if err := database.RunMigrations(context.Background(), db, zap.NewNop()); err != nil {
				t.Fatalf("run migrations: %v", err)
			}

			for name, q := range unchanged {
				if after := rowsOf(t, db, q); !reflect.DeepEqual(after, before[name]) {
					t.Errorf("%s changed:\nbefore %v\nafter  %v", name, before[name], after)
				}
			}
			// Integrity is checked with a join, not PRAGMA foreign_key_check:
			// account_balance_snapshots fails it outright (backend ISSUE-033).
			for _, table := range snapshotTables {
				var orphans int
				if err := db.Get(&orphans, `SELECT COUNT(*) FROM `+table+` s
					LEFT JOIN period_closings p ON p.closing_id = s.closing_id
					WHERE p.closing_id IS NULL`); err != nil {
					t.Fatalf("count %s orphans: %v", table, err)
				}
				if orphans != 0 {
					t.Errorf("%s: %d rows reference a missing period", table, orphans)
				}
				var fks int
				if err := db.Get(&fks, `SELECT COUNT(*) FROM pragma_foreign_key_list(?)
					WHERE "from" = 'closing_id' AND "table" = 'period_closings' AND "to" = 'closing_id'`, table); err != nil {
					t.Fatalf("list %s foreign keys: %v", table, err)
				}
				if fks != 1 {
					t.Errorf("%s: %d foreign keys closing_id -> period_closings(closing_id), want 1", table, fks)
				}
			}
			if got := stringsOf(t, db, indexesQ); !reflect.DeepEqual(got, wantIndexes) {
				t.Errorf("indexes = %v, want %v", got, wantIndexes)
			}
			if got := stringsOf(t, db, uniqueQ); !reflect.DeepEqual(got, wantUnique) {
				t.Errorf("unique constraints = %v, want %v", got, wantUnique)
			}

			// Merchant 2 opens the month merchant 1 closed; the new row must not
			// reuse a deleted closing_id.
			res, err := db.Exec(`INSERT INTO period_closings
				(merchant_id, closing_uuid, period_type, period_start, period_end, status)
				VALUES (2, '0192f3f2-0000-7000-8000-000000000906', 'MONTHLY', '2026-09-01', '2026-09-30', 'OPEN')`)
			if err != nil {
				t.Fatalf("insert period after migration: %v", err)
			}
			if id, err := res.LastInsertId(); err != nil || id != wantNextClosingID {
				t.Errorf("next closing_id = %d (err %v), want %d", id, err, wantNextClosingID)
			}
		})
	}
}
