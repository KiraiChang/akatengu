package database_test

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"akatengu/internal/database"
)

// The last migration before event_store and aggregate_versions accept the
// PERIOD aggregate type (20261022004, BFS R2 D8).
const beforePeriodAggregateType int64 = 20261022003

// Fixture: events and version rows of three aggregate types, with every
// nullable column set in some row. The counter (5) is ahead of the largest
// event_id (3), as after the latest two events were deleted.
var periodAggregateFixture = []string{
	`INSERT INTO event_store
		(event_id, event_uuid, occurred_at, merchant_id, aggregate_type, aggregate_id, aggregate_version,
		 event_type, payload, metadata, updated_by)
	VALUES
		(1, '0192f3f2-0000-7000-8000-000000000a01', '2026-10-01 09:00:00', 1, 'ACCOUNT', '1', 1,
		 'account.bulk_imported', '{"accounts":[]}', NULL, NULL),
		(2, '0192f3f2-0000-7000-8000-000000000a02', '2026-10-01 09:01:00', 1, 'ACCOUNT', '', 1,
		 'period.month_started', '{"period_start":"2026-10"}', '{"execution":{"source":"api"}}', 'r2-user'),
		(3, '0192f3f2-0000-7000-8000-000000000a03', '2026-10-02 09:00:00', 2, 'PREPAID_CATEGORY',
		 '0192f3f2-0000-7000-8000-000000000a03', 1, 'prepaid_category.created',
		 '{"name":"保險費","account_id":"1104-01","expense_account_id":"5201-01"}', NULL, '')`,
	`UPDATE sqlite_sequence SET seq = 5 WHERE name = 'event_store'`,
	`INSERT INTO aggregate_versions (aggregate_type, aggregate_id, merchant_id, current_version)
	VALUES ('ACCOUNT', '1', 1, 1), ('ACCOUNT', '', 1, 1),
	       ('PREPAID_CATEGORY', '0192f3f2-0000-7000-8000-000000000a03', 2, 1)`,
}

func newPrePeriodAggregateDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// Every new connection to ":memory:" is an empty database; keep a single one.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })

	if err := database.MigrateUpTo(context.Background(), db, beforePeriodAggregateType); err != nil {
		t.Fatalf("migrate up to %d: %v", beforePeriodAggregateType, err)
	}
	for _, s := range periodAggregateFixture {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("fixture: %v", err)
		}
	}
	return db
}

// uniqueOf lists the UNIQUE constraints declared on table, one sorted
// "col,col,..." string per constraint.
func uniqueOf(t *testing.T, db *sqlx.DB, table string) []string {
	t.Helper()
	got := stringsOf(t, db, `SELECT group_concat(name, ',') FROM (
			SELECT il.name AS idx, ii.name FROM pragma_index_list(?) il
			JOIN pragma_index_info(il.name) ii
			WHERE il."unique" = 1 AND il.origin = 'u' ORDER BY il.name, ii.seqno)
		GROUP BY idx`, table)
	sort.Strings(got)
	return got
}

func TestPeriodAggregateType_Migration(t *testing.T) {
	db := newPrePeriodAggregateDB(t)
	// Rows and columns the rebuild must keep as they were.
	unchanged := map[string]string{
		"event_store":                `SELECT * FROM event_store ORDER BY event_id`,
		"event_store columns":        `SELECT cid, name, type, "notnull", dflt_value, pk FROM pragma_table_info('event_store') ORDER BY cid`,
		"aggregate_versions":         `SELECT * FROM aggregate_versions ORDER BY aggregate_type, aggregate_id, merchant_id`,
		"aggregate_versions columns": `SELECT cid, name, type, "notnull", dflt_value, pk FROM pragma_table_info('aggregate_versions') ORDER BY cid`,
	}
	before := map[string][]map[string]any{}
	for name, q := range unchanged {
		before[name] = rowsOf(t, db, q)
	}
	const indexesQ = `SELECT name FROM sqlite_master
		WHERE type = 'index' AND tbl_name = ? AND sql IS NOT NULL ORDER BY name`
	beforeIndexes := map[string][]string{}
	beforeUnique := map[string][]string{}
	for _, table := range []string{"event_store", "aggregate_versions"} {
		beforeIndexes[table] = stringsOf(t, db, indexesQ, table)
		beforeUnique[table] = uniqueOf(t, db, table)
	}

	if err := database.RunMigrations(context.Background(), db, zap.NewNop()); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	for name, q := range unchanged {
		if after := rowsOf(t, db, q); !reflect.DeepEqual(after, before[name]) {
			t.Errorf("%s changed:\nbefore %v\nafter  %v", name, before[name], after)
		}
	}
	wantIndexes := map[string][]string{
		"event_store":        {"idx_es_aggregate", "idx_es_occurred", "idx_es_type"},
		"aggregate_versions": nil,
	}
	wantUnique := map[string][]string{
		"event_store":        {"aggregate_type,aggregate_id,aggregate_version,merchant_id", "event_uuid"},
		"aggregate_versions": nil,
	}
	for _, table := range []string{"event_store", "aggregate_versions"} {
		if got := stringsOf(t, db, indexesQ, table); !reflect.DeepEqual(got, wantIndexes[table]) || !reflect.DeepEqual(got, beforeIndexes[table]) {
			t.Errorf("%s indexes = %v, want %v (before %v)", table, got, wantIndexes[table], beforeIndexes[table])
		}
		if got := uniqueOf(t, db, table); !reflect.DeepEqual(got, wantUnique[table]) || !reflect.DeepEqual(got, beforeUnique[table]) {
			t.Errorf("%s unique constraints = %v, want %v (before %v)", table, got, wantUnique[table], beforeUnique[table])
		}
	}

	// PERIOD is accepted, other unknown types are still rejected, and a new
	// event does not reuse a deleted event_id.
	const insertVersion = `INSERT INTO aggregate_versions (aggregate_type, aggregate_id, merchant_id, current_version)
		VALUES (?, '1', 1, 1)`
	const insertEvent = `INSERT INTO event_store
		(merchant_id, aggregate_type, aggregate_id, aggregate_version, event_type, payload)
		VALUES (1, ?, '1', 1, 'period.month_started', '{"period_start":"2026-11"}')`
	if _, err := db.Exec(insertVersion, "PERIOD"); err != nil {
		t.Errorf("insert PERIOD version row: %v", err)
	}
	res, err := db.Exec(insertEvent, "PERIOD")
	if err != nil {
		t.Fatalf("insert PERIOD event: %v", err)
	}
	if id, err := res.LastInsertId(); err != nil || id != 6 {
		t.Errorf("next event_id = %d (err %v), want 6 (the fixture counter + 1)", id, err)
	}
	for _, q := range []string{insertVersion, insertEvent} {
		if _, err := db.Exec(q, "UNKNOWN_TYPE"); err == nil || !strings.Contains(err.Error(), "CHECK constraint failed") {
			t.Errorf("insert UNKNOWN_TYPE: err = %v, want a CHECK constraint failure", err)
		}
	}
}
