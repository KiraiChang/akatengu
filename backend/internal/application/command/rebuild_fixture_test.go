package command_test

import (
	"akatengu/internal/application/command"
	domainevent "akatengu/internal/domain/event"
	"akatengu/internal/enums/event_types"
	kevent "akatengu/internal/kernel/event"
	dbmodel "akatengu/internal/model/db"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

// The BFS R3 fixture writes every event type of the registry for one merchant,
// through the same Build and Append as the write APIs. The payloads are
// written by hand: the accounts are the ones seeds/accounts.sql gives merchant
// 1, the settings are rows of seeds/sys_accounts.sql and
// seeds/account_type_configs.sql.

// fixtureAccountID is not in seeds/accounts.sql.
const fixtureAccountID = "1101-81"

// fixtureTxn is a merchant's prerequisite transaction (BFS R3 D9): the cash
// flow handler needs an ACTIVE transaction and its entries, which only the
// legacy path creates before R4, so they are written with SQL.
type fixtureTxn struct{ txnUUID, debitEntry, creditEntry string }

// Merchant 1's transaction is the one newCommandTestDB writes.
var fixtureTxns = map[int64]fixtureTxn{
	1: {cmdTxnUUID, cmdDebitEntry, cmdCreditEntry},
	2: {"0192f3f5-0000-7000-8000-000000000001", "0192f3f5-0000-7000-8000-000000000011", "0192f3f5-0000-7000-8000-000000000012"},
}

// seedAccount is an account of the account.bulk_imported payload.
type seedAccount struct {
	AccountID        string  `db:"account_id" json:"account_id"`
	ParentID         *string `db:"parent_id" json:"parent_id"`
	Name             string  `db:"name" json:"name"`
	Type             string  `db:"type" json:"type"`
	NormalBalance    string  `db:"normal_balance" json:"normal_balance"`
	Currency         string  `db:"currency" json:"currency"`
	IsSummary        bool    `db:"is_summary" json:"is_summary"`
	IsActive         bool    `db:"is_active" json:"is_active"`
	Note             *string `db:"note" json:"note"`
	CashFlowCategory *string `db:"cash_flow_category" json:"cash_flow_category,omitempty"`
}

// newRebuildTestDB is newCommandTestDB plus merchant 2's prerequisite
// transaction, and the accounts seeds/accounts.sql wrote, read as
// BootstrapEventSeeds reads them: in account_id order, parents first. No
// entry_cf_categories row is written; that table belongs to the cash flow
// projector.
func newRebuildTestDB(t *testing.T) (*sqlx.DB, []seedAccount) {
	t.Helper()
	db := newCommandTestDB(t)

	var accounts []seedAccount
	if err := db.Select(&accounts, `SELECT account_id, parent_id, name, type, normal_balance, currency,
		is_summary, is_active, note, cash_flow_category FROM accounts WHERE merchant_id = 1 ORDER BY account_id`); err != nil {
		t.Fatalf("read seeded accounts: %v", err)
	}

	m2 := fixtureTxns[2]
	if _, err := db.Exec(`INSERT INTO transactions(txn_id, txn_uuid, merchant_id, txn_date, description, total_amount, currency, status, version)
		VALUES(2, ?, 2, '2026-10-01', 'r3 merchant 2 rent', 1000, 'TWD', 'ACTIVE', 1)`, m2.txnUUID); err != nil {
		t.Fatalf("insert transaction: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO journal_entries(entry_uuid, txn_uuid, merchant_id, txn_id, account_id, debit, credit)
		VALUES(?, ?, 2, 2, '5101-01', 1000, 0), (?, ?, 2, 2, '1101-02', 0, 1000)`,
		m2.debitEntry, m2.txnUUID, m2.creditEntry, m2.txnUUID); err != nil {
		t.Fatalf("insert journal entries: %v", err)
	}
	return db, accounts
}

// fixtureWriter writes the events of one merchant and records their types.
type fixtureWriter struct {
	t          *testing.T
	db         *sqlx.DB
	merchantID int64
	w          *command.Writer
	written    []event_types.EventType
}

// version is the aggregate version the client reads before sending.
func (f *fixtureWriter) version(aggType, aggID string) int64 {
	f.t.Helper()
	var v int64
	if err := f.db.Get(&v, `SELECT COALESCE(MAX(current_version), 0) FROM aggregate_versions
		WHERE aggregate_type = ? AND aggregate_id = ? AND merchant_id = ?`, aggType, aggID, f.merchantID); err != nil {
		f.t.Fatalf("read version of (%s, %s): %v", aggType, aggID, err)
	}
	return v
}

// write sends payload, the JSON the client sends, through Build and Append.
func (f *fixtureWriter) write(eventType event_types.EventType, payload any, expected int64) *dbmodel.EventStore {
	f.t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		f.t.Fatalf("marshal %s: %v", eventType, err)
	}
	decoded, err := domainevent.Decode(eventType, raw)
	if err != nil {
		f.t.Fatalf("decode %s: %v", eventType, err)
	}
	p, ok := decoded.(kevent.Payload)
	if !ok {
		f.t.Fatalf("decoded %T is not an event payload", decoded)
	}
	ctx := cmdCtx(f.merchantID)
	evt, err := command.Build(ctx, p, expected, kevent.EventParams{Source: "api"})
	if err != nil {
		f.t.Fatalf("build %s: %v", eventType, err)
	}
	got, err := f.w.Append(ctx, evt)
	if err != nil {
		f.t.Fatalf("append %s of merchant %d: %v", eventType, f.merchantID, err)
	}
	f.written = append(f.written, eventType)
	return got
}

// writeFixtureEvents writes the events of every registered event type for
// merchantID and returns the event types written, in order.
func writeFixtureEvents(t *testing.T, db *sqlx.DB, merchantID int64, accounts []seedAccount) []event_types.EventType {
	t.Helper()
	f := &fixtureWriter{t: t, db: db, merchantID: merchantID, w: command.NewWriter(db, command.NewEngine(db))}

	f.write(event_types.EventAccountBulkImported.Enum(), map[string]any{"accounts": accounts}, f.version("ACCOUNT", "1"))

	account := map[string]any{
		"account_id": fixtureAccountID, "parent_id": "1101", "name": "R3 測試存款", "type": "ASSET",
		"normal_balance": "DEBIT", "currency": "TWD", "is_active": true, "cash_flow_category": "CASH",
	}
	f.write(event_types.EventAccountCreated.Enum(), account, 0)
	account["name"], account["version"] = "R3 測試存款（改名）", 1
	f.write(event_types.EventAccountUpdated.Enum(), account, f.version("ACCOUNT", fixtureAccountID))

	ledger := map[string]any{
		"account_id": "1101-02", "institution": "R3 測試銀行", "name": "R3 活期帳戶",
		"type": "BANK_ACCOUNT", "currency": "TWD", "is_active": true,
	}
	created := f.write(event_types.EventLedgerAccountCreated.Enum(), ledger, f.version("ACCOUNT", "1101-02"))
	var recorded struct {
		LedgerID int64 `json:"ledger_id"`
	}
	if err := json.Unmarshal(created.Payload, &recorded); err != nil {
		t.Fatalf("decode ledger_account.created: %v", err)
	}
	ledger["ledger_id"], ledger["name"], ledger["version"] = recorded.LedgerID, "R3 活期帳戶（改名）", 1
	f.write(event_types.EventLedgerAccountUpdated.Enum(), ledger, f.version("ACCOUNT", "1101-02"))

	f.write(event_types.EventSysAccountUpdated.Enum(), map[string]any{
		"sys_code": "SYS:ASSET:PREPAID_INTEREST", "description": "預付利息（資產）", "account_id": "1104-04",
	}, f.version("SYS_CONFIG", "1"))
	f.write(event_types.EventLedgerAccountTypeConfigUpdated.Enum(), map[string]any{
		"type": "BANK_ACCOUNT", "account_id": "1101",
	}, f.version("ACCOUNT_CONFIG", "1"))
	f.write(event_types.EventAssetTypeAccountConfigUpdated.Enum(), map[string]any{
		"asset_type": "STOCK", "realized_gain_account_id": "4203-01", "realized_loss_account_id": "5402-11",
		"unrealized_gain_account_id": "4204-01", "unrealized_loss_account_id": "5402-21", "oci_account_id": "3102-02",
		"fee_account_id": "5402-01", "tax_account_id": "5402-05", "account_id": "1203-02",
	}, f.version("ACCOUNT_CONFIG", "1"))

	prepaid := f.write(event_types.EventPrepaidCategoryCreated.Enum(), map[string]any{
		"name": "預付利息", "account_id": "1104-04", "expense_account_id": "5401-01",
	}, 0)
	f.write(event_types.EventPrepaidCategoryUpdated.Enum(), map[string]any{
		"category_uuid": prepaid.EventUuid, "name": "預付貸款利息", "account_id": "1104-04", "expense_account_id": "5401-01",
	}, f.version("PREPAID_CATEGORY", prepaid.EventUuid))
	f.write(event_types.EventPrepaidCategoryDeleted.Enum(), map[string]any{"category_uuid": prepaid.EventUuid},
		f.version("PREPAID_CATEGORY", prepaid.EventUuid))

	fixed := f.write(event_types.EventFixedAssetCategoryCreated.Enum(), map[string]any{
		"name": "電腦設備", "asset_account_id": "1201-04", "accum_depreciation_account_id": "1201-99",
		"depreciation_expense_account_id": "5501-03",
	}, 0)
	f.write(event_types.EventFixedAssetCategoryUpdated.Enum(), map[string]any{
		"category_uuid": fixed.EventUuid, "name": "電腦及週邊設備", "asset_account_id": "1201-04",
		"accum_depreciation_account_id": "1201-99", "depreciation_expense_account_id": "5501-03",
	}, f.version("FIXED_ASSET_CATEGORY", fixed.EventUuid))
	f.write(event_types.EventFixedAssetCategoryDeleted.Enum(), map[string]any{"category_uuid": fixed.EventUuid},
		f.version("FIXED_ASSET_CATEGORY", fixed.EventUuid))

	f.write(event_types.EventPeriodMonthStarted.Enum(), map[string]any{"period_start": "2026-10"}, 0)
	f.write(event_types.EventPeriodAnnualStarted.Enum(), map[string]any{"year": 2026}, 0)

	txn := fixtureTxns[merchantID]
	f.write(event_types.EventTransactionCFCategoryUpdated.Enum(), map[string]any{
		"txn_uuid": txn.txnUUID, "entries": []map[string]any{{"entry_uuid": txn.debitEntry, "cf_category": "INVESTING"}},
	}, f.version("TRANSACTION", txn.txnUUID))

	return f.written
}

// expectAllRegistered fails when the fixture misses an event type of the
// registry (BFS R3 D9).
func expectAllRegistered(t *testing.T, written []event_types.EventType) {
	t.Helper()
	seen := make(map[event_types.EventType]bool, len(written))
	for _, e := range written {
		seen[e] = true
	}
	for _, s := range domainevent.Specs() {
		if !seen[s.EventType] {
			t.Errorf("the fixture does not write %s", s.EventType)
		}
	}
}

// merchantTables lists the tables with a merchant_id column.
func merchantTables(t *testing.T, db *sqlx.DB) []string {
	t.Helper()
	var tables []string
	if err := db.Select(&tables, `SELECT m.name FROM sqlite_master m
		WHERE m.type = 'table' AND m.name NOT LIKE 'sqlite_%'
		  AND EXISTS (SELECT 1 FROM pragma_table_info(m.name) WHERE name = 'merchant_id')
		ORDER BY m.name`); err != nil {
		t.Fatalf("list tables: %v", err)
	}
	return tables
}

// merchantRows dumps every column of merchantID's rows of each table, in rowid
// order. A table without rows of the merchant has no entry.
func merchantRows(t *testing.T, db *sqlx.DB, tables []string, merchantID int64) map[string][]string {
	t.Helper()
	out := make(map[string][]string)
	for _, table := range tables {
		rows, err := db.Queryx(fmt.Sprintf(`SELECT * FROM %q WHERE merchant_id = ? ORDER BY rowid`, table), merchantID)
		if err != nil {
			t.Fatalf("read %s: %v", table, err)
		}
		for rows.Next() {
			row := make(map[string]any)
			if err := rows.MapScan(row); err != nil {
				rows.Close()
				t.Fatalf("scan %s: %v", table, err)
			}
			out[table] = append(out[table], formatRow(row))
		}
		if err := rows.Close(); err != nil {
			t.Fatalf("close %s: %v", table, err)
		}
	}
	return out
}

func formatRow(row map[string]any) string {
	keys := make([]string, 0, len(row))
	for k := range row {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fields := make([]string, len(keys))
	for i, k := range keys {
		v := row[k]
		if b, ok := v.([]byte); ok {
			v = string(b)
		}
		fields[i] = fmt.Sprintf("%s=%v", k, v)
	}
	return strings.Join(fields, ", ")
}
