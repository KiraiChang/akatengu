package services_test

import (
	"akatengu/internal/database"
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/shared/utils/test"
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// The last migration before category read-model versions are aligned with
// their aggregates (20261022002, backend ISSUE-029).
const beforeCategoryVersionSync int64 = 20261022001

// syncedCategoryUUID is the category written by the migration fixture.
const syncedCategoryUUID = "0192f3f2-0000-7000-8000-000000000801"

// categoryKind runs a scenario for one category type. The two types share the
// same handler, pipeline and projection code; the accounts come from
// seeds/accounts.sql, the same as category_test.go.
type categoryKind struct {
	name          string
	table         string
	aggregate     enums.AggregateType
	createdType   event_types.EventType
	updatedType   event_types.EventType
	deletedType   event_types.EventType
	createPayload any
	updatePayload func(categoryUUID, name string) any
	deletePayload func(categoryUUID string) any
	// fixtureSQL inserts a merchant 1 category at version 1; its arguments are
	// the category UUID and is_active.
	fixtureSQL string
}

// categoryKinds builds the kinds after the enums are registered.
func categoryKinds() []categoryKind {
	test.InitEnums()
	return []categoryKind{
		{
			name:          "預付類別",
			table:         "prepaid_categories",
			aggregate:     enums.AggregatePrepaidCategory.Enum(),
			createdType:   event_types.EventPrepaidCategoryCreated.Enum(),
			updatedType:   event_types.EventPrepaidCategoryUpdated.Enum(),
			deletedType:   event_types.EventPrepaidCategoryDeleted.Enum(),
			createPayload: payload.PrepaidCategoryCreatedPayload{Name: "保險費", AccountID: "1104-01", ExpenseAccountID: "5201-01"},
			updatePayload: func(categoryUUID, name string) any {
				return payload.PrepaidCategoryUpdatedPayload{CategoryUUID: categoryUUID, Name: name, AccountID: "1104-01", ExpenseAccountID: "5201-01"}
			},
			deletePayload: func(categoryUUID string) any {
				return payload.PrepaidCategoryDeletedPayload{CategoryUUID: categoryUUID}
			},
			fixtureSQL: `INSERT INTO prepaid_categories
				(category_uuid, merchant_id, name, account_id, expense_account_id, is_active, updated_by, updated_at, version)
				VALUES (?, 1, '保險費', '1104-01', '5201-01', ?, 'f8-user', '2026-06-01 09:00:00', 1)`,
		},
		{
			name:        "固定資產類別",
			table:       "fixed_asset_categories",
			aggregate:   enums.AggregateFixedAssetCategory.Enum(),
			createdType: event_types.EventFixedAssetCategoryCreated.Enum(),
			updatedType: event_types.EventFixedAssetCategoryUpdated.Enum(),
			deletedType: event_types.EventFixedAssetCategoryDeleted.Enum(),
			createPayload: payload.FixedAssetCategoryCreatedPayload{Name: "電腦設備",
				AssetAccountID: "1601-01", AccumDepreciationAccountID: "1601-02", DepreciationExpenseAccountID: "5601-01"},
			updatePayload: func(categoryUUID, name string) any {
				return payload.FixedAssetCategoryUpdatedPayload{CategoryUUID: categoryUUID, Name: name,
					AssetAccountID: "1601-01", AccumDepreciationAccountID: "1601-02", DepreciationExpenseAccountID: "5601-01"}
			},
			deletePayload: func(categoryUUID string) any {
				return payload.FixedAssetCategoryDeletedPayload{CategoryUUID: categoryUUID}
			},
			fixtureSQL: `INSERT INTO fixed_asset_categories
				(category_uuid, merchant_id, name, asset_account_id, accum_depreciation_account_id,
				 depreciation_expense_account_id, is_active, updated_by, updated_at, version)
				VALUES (?, 1, '電腦設備', '1601-01', '1601-02', '5601-01', ?, 'f8-user', '2026-06-01 09:00:00', 1)`,
		},
	}
}

// categoryState is what a category write touches: the read model, the version
// of the category_uuid aggregate (0 = no row) and the number of events.
type categoryState struct {
	Name      string `db:"name"`
	IsActive  bool   `db:"is_active"`
	Version   int64  `db:"version"`
	Aggregate int64  `db:"aggregate"`
	Events    int    `db:"events"`
}

func categoryStateOf(t *testing.T, db *sqlx.DB, k categoryKind, categoryUUID string) categoryState {
	t.Helper()
	var s categoryState
	if err := db.Get(&s, `SELECT c.name, c.is_active, c.version,
		COALESCE((SELECT current_version FROM aggregate_versions
		          WHERE aggregate_type = ? AND aggregate_id = c.category_uuid AND merchant_id = c.merchant_id), 0) AS aggregate,
		(SELECT COUNT(*) FROM event_store) AS events
		FROM `+k.table+` c WHERE c.category_uuid = ?`, k.aggregate.String(), categoryUUID); err != nil {
		t.Fatalf("read %s state: %v", k.table, err)
	}
	return s
}

// categoryRowOf reads every column of a category, to compare whole rows.
func categoryRowOf(t *testing.T, db *sqlx.DB, k categoryKind, categoryUUID string) map[string]any {
	t.Helper()
	row := map[string]any{}
	if err := db.QueryRowx(`SELECT * FROM `+k.table+` WHERE category_uuid = ?`, categoryUUID).MapScan(row); err != nil {
		t.Fatalf("read %s row: %v", k.table, err)
	}
	return row
}

// appendCategory writes an update or delete the way the category handlers do:
// on the category_uuid aggregate at the version the screen sends.
func appendCategory(db *sqlx.DB, k categoryKind, categoryUUID string, expected int64, eventType event_types.EventType, p any) error {
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = newSvc(db).Append(testCtx(), cmd.AppendCmd{
		AggregateType:   k.aggregate,
		AggregateID:     categoryUUID,
		ExpectedVersion: expected,
		EventType:       eventType,
		Payload:         raw,
	})
	return err
}

func TestCategoryVersion_Write(t *testing.T) {
	type categoryStep struct {
		delete   bool
		name     string // the new name of an update
		expected int64  // 0 = the read-model version, as the screen sends
		wantErr  string // "" = written
	}
	cases := []struct {
		name  string
		steps []categoryStep
		want  categoryState // Events is not compared
	}{
		{
			name:  "建立後連續修改兩次",
			steps: []categoryStep{{name: "第一次修改"}, {name: "第二次修改"}},
			want:  categoryState{Name: "第二次修改", IsActive: true, Version: 3, Aggregate: 3},
		},
		{
			name:  "修改一次後停用",
			steps: []categoryStep{{name: "第一次修改"}, {delete: true}},
			want:  categoryState{Name: "第一次修改", IsActive: false, Version: 3, Aggregate: 3},
		},
		{
			name: "修改一次後以修改前的版本 1 再修改",
			steps: []categoryStep{
				{name: "第一次修改"},
				{name: "過期的修改", expected: 1, wantErr: "UNIQUE constraint failed: aggregate_versions."},
			},
			want: categoryState{Name: "第一次修改", IsActive: true, Version: 2, Aggregate: 2},
		},
	}
	for _, k := range categoryKinds() {
		for _, tc := range cases {
			t.Run(k.name+"/"+tc.name, func(t *testing.T) {
				db := test.NewTestDB(t)
				// Every new connection to ":memory:" is an empty database; keep a single one.
				db.SetMaxOpenConns(1)

				// Created the way the category handlers do: a new aggregate at version 0.
				aggregateID, err := uuid.NewV7()
				if err != nil {
					t.Fatalf("new aggregate id: %v", err)
				}
				created, err := newSvc(db).Append(testCtx(), cmd.AppendCmd{
					AggregateType: k.aggregate,
					AggregateID:   aggregateID.String(),
					EventType:     k.createdType,
					Payload:       mustMarshal(t, k.createPayload),
				})
				if err != nil {
					t.Fatalf("create: %v", err)
				}
				categoryUUID := created.EventUuid

				for i, step := range tc.steps {
					before := categoryStateOf(t, db, k, categoryUUID)
					beforeRow := categoryRowOf(t, db, k, categoryUUID)
					expected := step.expected
					if expected == 0 {
						expected = before.Version
					}
					eventType, p := k.updatedType, k.updatePayload(categoryUUID, step.name)
					if step.delete {
						eventType, p = k.deletedType, k.deletePayload(categoryUUID)
					}

					err := appendCategory(db, k, categoryUUID, expected, eventType, p)

					after := categoryStateOf(t, db, k, categoryUUID)
					if step.wantErr != "" {
						if err == nil || !strings.Contains(err.Error(), step.wantErr) {
							t.Fatalf("step %d: err = %v, want %q", i, err, step.wantErr)
						}
						if after != before {
							t.Fatalf("step %d: state changed: before %+v, after %+v", i, before, after)
						}
						if afterRow := categoryRowOf(t, db, k, categoryUUID); !reflect.DeepEqual(afterRow, beforeRow) {
							t.Fatalf("step %d: row changed:\nbefore %v\nafter  %v", i, beforeRow, afterRow)
						}
						continue
					}
					if err != nil {
						t.Fatalf("step %d: Append: %v", i, err)
					}
					if after.Version != before.Version+1 || after.Aggregate != after.Version {
						t.Fatalf("step %d: version = %d, aggregate = %d, want both %d", i, after.Version, after.Aggregate, before.Version+1)
					}
				}

				got := categoryStateOf(t, db, k, categoryUUID)
				got.Events = 0
				if got != tc.want {
					t.Fatalf("state = %+v, want %+v", got, tc.want)
				}
			})
		}
	}
}

func newPreCategorySyncDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// Every new connection to ":memory:" is an empty database; keep a single one.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })

	if err := database.MigrateUpTo(context.Background(), db, beforeCategoryVersionSync); err != nil {
		t.Fatalf("migrate up to %d: %v", beforeCategoryVersionSync, err)
	}
	return db
}

func TestCategoryVersion_Migration(t *testing.T) {
	type versionRow struct {
		merchantID int64
		version    int64
	}
	cases := []struct {
		name        string
		isActive    bool
		versionRow  *versionRow // the category_uuid aggregate row; nil = none
		wantVersion int64       // the read-model version after the migration
		wantEdited  int64       // read-model and aggregate version after an edit; 0 = not edited
	}{
		{
			name:        "舊版修改過一次：讀模型 1、aggregate 2",
			isActive:    true,
			versionRow:  &versionRow{merchantID: 1, version: 2},
			wantVersion: 2,
			wantEdited:  3,
		},
		{
			name:        "舊版停用過一次：讀模型 1、停用、aggregate 2",
			versionRow:  &versionRow{merchantID: 1, version: 2},
			wantVersion: 2,
		},
		{
			name:        "從未修改：讀模型 1、沒有版本紀錄",
			isActive:    true,
			wantVersion: 1,
			wantEdited:  2,
		},
		{
			// Not reachable through the API: the pipeline looks the category up by merchant.
			name:        "同一 category_uuid 只在商戶 2 下有版本紀錄",
			isActive:    true,
			versionRow:  &versionRow{merchantID: 2, version: 5},
			wantVersion: 1,
		},
	}
	for _, k := range categoryKinds() {
		for _, tc := range cases {
			t.Run(k.name+"/"+tc.name, func(t *testing.T) {
				ctx := context.Background()
				db := newPreCategorySyncDB(t)
				if _, err := db.Exec(k.fixtureSQL, syncedCategoryUUID, tc.isActive); err != nil {
					t.Fatalf("insert category: %v", err)
				}
				if tc.versionRow != nil {
					if _, err := db.Exec(`INSERT INTO aggregate_versions (aggregate_type, aggregate_id, merchant_id, current_version)
						VALUES (?, ?, ?, ?)`, k.aggregate.String(), syncedCategoryUUID, tc.versionRow.merchantID, tc.versionRow.version); err != nil {
						t.Fatalf("insert version row: %v", err)
					}
				}
				before := categoryRowOf(t, db, k, syncedCategoryUUID)

				if err := database.RunMigrations(ctx, db, zap.NewNop()); err != nil {
					t.Fatalf("run migrations: %v", err)
				}
				// Seeded after the migrations, the same as test.NewTestDB.
				if err := database.NewSQLFileSeeder("accounts", "accounts.sql").Seed(ctx, db.DB); err != nil {
					t.Fatalf("seed accounts: %v", err)
				}

				after := categoryRowOf(t, db, k, syncedCategoryUUID)
				if got := after["version"]; got != tc.wantVersion {
					t.Fatalf("version after migration = %v, want %d", got, tc.wantVersion)
				}
				// Only version may change.
				delete(before, "version")
				delete(after, "version")
				if !reflect.DeepEqual(after, before) {
					t.Fatalf("row changed:\nbefore %v\nafter  %v", before, after)
				}
				if tc.wantEdited == 0 {
					return
				}

				if err := appendCategory(db, k, syncedCategoryUUID, tc.wantVersion, k.updatedType,
					k.updatePayload(syncedCategoryUUID, "migration 後修改")); err != nil {
					t.Fatalf("edit after migration: %v", err)
				}
				got := categoryStateOf(t, db, k, syncedCategoryUUID)
				if got.Version != tc.wantEdited || got.Aggregate != tc.wantEdited {
					t.Fatalf("after edit: version = %d, aggregate = %d, want both %d", got.Version, got.Aggregate, tc.wantEdited)
				}
			})
		}
	}
}
