package services_test

import (
	"akatengu/internal/application/command"
	domainevent "akatengu/internal/domain/event"
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	kerrors "akatengu/internal/kernel/errors"
	kevent "akatengu/internal/kernel/event"
	dbmodel "akatengu/internal/model/db"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store/projection_repo"
	"akatengu/internal/services"
	"akatengu/internal/shared/utils/test"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

// Fixture identifiers. 1101, 1101-02 and 1203-02 come from seeds/accounts.sql;
// 1101-90, 1101-91 and 9999-99 are not in the seeds.
const (
	versionCreatedAccountID    = "1101-90" // written by account.created
	versionImportedAccountID   = "1101-91" // written by account.bulk_imported
	versionLedgerAccountID     = "1101-02"
	versionInvestmentAccountID = "1203-02"
	versionMissingAccountID    = "9999-99"
)

const versionMissingLedgerID int64 = 999

// versionConflictText matches a conflict by the F6 message: BFS returns a
// kerrors.EventError, which does not unwrap, and keeps the message (BFS R2 D9).
var versionConflictText = projection_repo.ErrVersionConflict.Error()

// ─────────────────────────────────────────
// Fixture
// ─────────────────────────────────────────

// versionFixture is a migrated DB with two accounts, one ledger account and one
// investment, written the way POST /api/event/append writes them: through BFS
// for the event types migrated to it, otherwise EventStoreService.Append.
type versionFixture struct {
	t               *testing.T
	db              *sqlx.DB
	svc             *services.EventStoreService
	writer          *command.Writer
	versions        query.AggerateRepo
	ledgerID        int64
	ledgerEventUUID string // the ledger account's creation event
	investmentUUID  string
}

func newVersionFixture(t *testing.T) *versionFixture {
	t.Helper()
	db := test.NewTestDB(t)
	// Every new connection to ":memory:" is an empty database; keep a single one.
	db.SetMaxOpenConns(1)
	f := &versionFixture{t: t, db: db, svc: newSvc(db), writer: command.NewWriter(db, command.NewEngine(db)),
		versions: query.NewAggerateRepo(db)}

	f.mustAppend(enums.AggregateAccount.Enum(), versionCreatedAccountID, event_types.EventAccountCreated.Enum(),
		versionAccountCreate(versionCreatedAccountID, "版本測試科目"))
	// The generic API does not accept the bulk import; write it as the bootstrap
	// seed does (bootstrap/event_seed.go).
	f.mustSeed(event_types.EventAccountBulkImported.Enum(),
		payload.AccountBulkImportedPayload{Accounts: []payload.AccountCreatePayload{
			versionAccountCreate(versionImportedAccountID, "版本測試匯入科目"),
		}})

	evt := f.mustAppend(enums.AggregateAccount.Enum(), versionLedgerAccountID, event_types.EventLedgerAccountCreated.Enum(),
		payload.LedgerAccountCreatePayload{
			AccountId:   versionLedgerAccountID,
			Institution: "版本測試銀行",
			Name:        "版本測試帳戶",
			Currency:    "TWD",
			IsActive:    true,
			Type:        enums.LedgerAccountTypeBankAccount.Enum(),
		})
	f.ledgerEventUUID = evt.EventUuid
	err := db.Get(&f.ledgerID, `SELECT ledger_id FROM ledger_accounts WHERE ledger_uuid = ? AND merchant_id = ?`,
		evt.EventUuid, testMID)
	if err != nil {
		t.Fatalf("ledger id: %v", err)
	}

	// The investment's uuid is its creation event's UUID.
	evt = f.mustAppend(enums.AggregateTransaction.Enum(), "", event_types.EventInvestmentCreated.Enum(),
		versionInvestmentCreate("版本測試投資"))
	f.investmentUUID = evt.EventUuid
	return f
}

// append reads the aggregate's current version right before appending, as the
// frontend does with getAggregateVersion, and dispatches the request as
// POST /api/event/append does (handler/event.go).
func (f *versionFixture) append(aggType enums.AggregateType, aggID string, et event_types.EventType, p any) (*dbmodel.EventStore, error) {
	f.t.Helper()
	raw, err := json.Marshal(p)
	if err != nil {
		f.t.Fatalf("marshal payload: %v", err)
	}
	req := cmd.AppendCmd{
		AggregateType:   aggType,
		AggregateID:     aggID,
		ExpectedVersion: f.aggVersion(aggType, aggID),
		EventType:       et,
		Payload:         raw,
	}
	result, handled, err := f.writer.DispatchLegacy(testCtx(), req, kevent.EventParams{Source: "api"})
	if !handled {
		return f.svc.Append(testCtx(), req)
	}
	return result, err
}

// mustSeed writes p through BFS at version 0, as bootstrap/event_seed.go does.
func (f *versionFixture) mustSeed(et event_types.EventType, p any) {
	f.t.Helper()
	raw, err := json.Marshal(p)
	if err != nil {
		f.t.Fatalf("marshal payload: %v", err)
	}
	decoded, err := domainevent.Decode(et, raw)
	if err != nil {
		f.t.Fatalf("decode %s: %v", et, err)
	}
	pl, ok := decoded.(kevent.Payload)
	if !ok {
		f.t.Fatalf("decoded %T is not an event payload", decoded)
	}
	evt, err := command.Build(testCtx(), pl, 0, kevent.EventParams{Source: "seed"})
	if err != nil {
		f.t.Fatalf("Build %s: %v", et, err)
	}
	if _, err := f.writer.Append(testCtx(), evt); err != nil {
		f.t.Fatalf("Append %s: %v", et, err)
	}
}

func (f *versionFixture) mustAppend(aggType enums.AggregateType, aggID string, et event_types.EventType, p any) *dbmodel.EventStore {
	f.t.Helper()
	evt, err := f.append(aggType, aggID, et, p)
	if err != nil {
		f.t.Fatalf("Append %s: %v", et, err)
	}
	return evt
}

func (f *versionFixture) aggVersion(aggType enums.AggregateType, aggID string) int64 {
	f.t.Helper()
	v, err := f.versions.GetVersion(testCtx(), aggType, aggID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0
	}
	if err != nil {
		f.t.Fatalf("aggregate version: %v", err)
	}
	return v
}

// ─────────────────────────────────────────
// Snapshots
// ─────────────────────────────────────────

// versionSnapshot is what a rejected write must leave unchanged.
type versionSnapshot struct {
	exists     bool
	name       string
	version    int64
	events     int   // rows in event_store
	aggVersion int64 // version of the aggregate the write goes to (0 = none)
}

// snapshot reads name and version of the row whose keyCol equals key in table.
func (f *versionFixture) snapshot(table, keyCol string, key any, aggType enums.AggregateType, aggID string) versionSnapshot {
	f.t.Helper()
	s := versionSnapshot{aggVersion: f.aggVersion(aggType, aggID)}

	var row struct {
		Name    string `db:"name"`
		Version int64  `db:"version"`
	}
	err := f.db.Get(&row, `SELECT name, version FROM `+table+` WHERE `+keyCol+` = ? AND merchant_id = ?`, key, testMID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		f.t.Fatalf("read %s: %v", table, err)
	default:
		s.exists, s.name, s.version = true, row.Name, row.Version
	}

	if err := f.db.Get(&s.events, `SELECT COUNT(*) FROM event_store`); err != nil {
		f.t.Fatalf("count events: %v", err)
	}
	return s
}

func (f *versionFixture) accountSnapshot(accountID string) versionSnapshot {
	return f.snapshot("accounts", "account_id", accountID, enums.AggregateAccount.Enum(), accountID)
}

func (f *versionFixture) ledgerSnapshot(ledgerID int64) versionSnapshot {
	return f.snapshot("ledger_accounts", "ledger_id", ledgerID, enums.AggregateAccount.Enum(), versionLedgerAccountID)
}

func (f *versionFixture) investmentSnapshot() versionSnapshot {
	return f.snapshot("investments", "uuid", f.investmentUUID, enums.AggregateTransaction.Enum(), "")
}

// expectStep checks one write. A rejected write must leave the row, event_store
// and the aggregate version unchanged; an accepted one adds exactly one event.
func expectStep(t *testing.T, err, wantErr error, wantErrText string, before, after versionSnapshot) {
	t.Helper()
	switch {
	case wantErr != nil:
		if !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	case wantErrText != "":
		if err == nil || !strings.Contains(err.Error(), wantErrText) {
			t.Fatalf("err = %v, want an error containing %q", err, wantErrText)
		}
	case err != nil:
		t.Fatalf("err = %v, want nil", err)
	}

	if err != nil {
		if after != before {
			t.Fatalf("rejected write changed state: before %+v, after %+v", before, after)
		}
		return
	}
	if after.events != before.events+1 || after.aggVersion != before.aggVersion+1 {
		t.Fatalf("accepted write: before %+v, after %+v, want one more event and aggregate version", before, after)
	}
}

func expectRow(t *testing.T, got versionSnapshot, wantName string, wantVersion int64) {
	t.Helper()
	if wantName == "" {
		if got.exists {
			t.Fatalf("row exists: %+v, want none", got)
		}
		return
	}
	if !got.exists || got.name != wantName || got.version != wantVersion {
		t.Fatalf("row = %+v, want name %q version %d", got, wantName, wantVersion)
	}
}

// ─────────────────────────────────────────
// Payloads
// ─────────────────────────────────────────

func versionAccountCreate(accountID, name string) payload.AccountCreatePayload {
	parentID := "1101"
	return payload.AccountCreatePayload{
		AccountId:     accountID,
		ParentId:      &parentID,
		Name:          name,
		Type:          enums.AccountAsset.Enum(),
		NormalBalance: enums.BalanceDebit.Enum(),
		Currency:      "TWD",
		IsActive:      true,
	}
}

func versionAccountUpdate(accountID, name string, version int64) payload.AccountUpdatedPayload {
	c := versionAccountCreate(accountID, name)
	return payload.AccountUpdatedPayload{
		AccountId:     c.AccountId,
		ParentId:      c.ParentId,
		Name:          c.Name,
		Type:          c.Type,
		NormalBalance: c.NormalBalance,
		Currency:      c.Currency,
		IsActive:      c.IsActive,
		Version:       version,
	}
}

func versionLedgerUpdate(ledgerID int64, name string, version int64) payload.LedgerAccountUpdatedPayload {
	return payload.LedgerAccountUpdatedPayload{
		LedgerId:    ledgerID,
		AccountId:   versionLedgerAccountID,
		Institution: "版本測試銀行",
		Name:        name,
		Currency:    "TWD",
		IsActive:    true,
		Version:     version,
		Type:        enums.LedgerAccountTypeBankAccount.Enum(),
	}
}

func versionInvestmentCreate(name string) payload.InvestmentCreatedPayload {
	return payload.InvestmentCreatedPayload{
		AccountId:    versionInvestmentAccountID,
		AssetType:    enums.AssetTypeStock.Enum(),
		Currency:     "TWD",
		Symbol:       "VERSION-TEST",
		Name:         name,
		CostMethod:   enums.CostMethodFIFO.Enum(),
		IFRSCategory: enums.IFRSCategoryFVOCI.Enum(),
		IsActive:     true,
	}
}

// ─────────────────────────────────────────
// Tests
// ─────────────────────────────────────────

func TestAccountUpdatedVersion(t *testing.T) {
	type step struct {
		create      bool  // account.created instead of account.updated
		version     int64 // version sent with account.updated
		name        string
		wantErr     error
		wantErrText string
	}
	cases := []struct {
		name        string
		accountID   string
		steps       []step // applied in order; earlier steps set up the state
		wantName    string // "" = the account must not exist
		wantVersion int64
	}{
		{
			name:        "account.created 之後以 version 1 修改",
			accountID:   versionCreatedAccountID,
			steps:       []step{{version: 1, name: "修改後名稱"}},
			wantName:    "修改後名稱",
			wantVersion: 2,
		},
		{
			name:        "account.bulk_imported 之後以 version 1 修改",
			accountID:   versionImportedAccountID,
			steps:       []step{{version: 1, name: "修改後名稱"}},
			wantName:    "修改後名稱",
			wantVersion: 2,
		},
		{
			name:      "已在 version 2 時以過期的 version 1 修改",
			accountID: versionCreatedAccountID,
			steps: []step{
				{version: 1, name: "修改後名稱"},
				{version: 1, name: "過期畫面的名稱", wantErrText: versionConflictText},
			},
			wantName:    "修改後名稱",
			wantVersion: 2,
		},
		{
			// account.created only creates its aggregate (BFS R2 section 3), so the
			// existing aggregate is a conflict before the unique key is reached.
			name:      "已在 version 2 時以同一個編號重複新增",
			accountID: versionCreatedAccountID,
			steps: []step{
				{version: 1, name: "修改後名稱"},
				{create: true, name: "重複新增的名稱", wantErrText: string(kerrors.ErrVersionConflict)},
			},
			wantName:    "修改後名稱",
			wantVersion: 2,
		},
		{
			name:      "重複新增失敗後，再以過期的 version 1 修改",
			accountID: versionCreatedAccountID,
			steps: []step{
				{version: 1, name: "修改後名稱"},
				{create: true, name: "重複新增的名稱", wantErrText: string(kerrors.ErrVersionConflict)},
				{version: 1, name: "過期畫面的名稱", wantErrText: versionConflictText},
			},
			wantName:    "修改後名稱",
			wantVersion: 2,
		},
		{
			name:      "修改不存在的科目",
			accountID: versionMissingAccountID,
			steps: []step{
				{version: 1, name: "不存在的科目", wantErrText: versionConflictText},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newVersionFixture(t)
			for _, s := range tc.steps {
				et, p := event_types.EventAccountUpdated.Enum(), any(versionAccountUpdate(tc.accountID, s.name, s.version))
				if s.create {
					et, p = event_types.EventAccountCreated.Enum(), versionAccountCreate(tc.accountID, s.name)
				}
				before := f.accountSnapshot(tc.accountID)
				_, err := f.append(enums.AggregateAccount.Enum(), tc.accountID, et, p)
				expectStep(t, err, s.wantErr, s.wantErrText, before, f.accountSnapshot(tc.accountID))
			}
			expectRow(t, f.accountSnapshot(tc.accountID), tc.wantName, tc.wantVersion)
		})
	}
}

func TestLedgerAccountUpdatedVersion(t *testing.T) {
	type step struct {
		version     int64
		name        string
		wantErrText string
	}
	cases := []struct {
		name        string
		missing     bool   // update versionMissingLedgerID instead of the fixture ledger
		steps       []step // applied in order; earlier steps set up the state
		wantName    string // fixture ledger afterwards
		wantVersion int64
	}{
		{
			name:        "以 version 1 修改",
			steps:       []step{{version: 1, name: "修改後帳戶"}},
			wantName:    "修改後帳戶",
			wantVersion: 2,
		},
		{
			name: "已在 version 2 時以過期的 version 1 修改",
			steps: []step{
				{version: 1, name: "修改後帳戶"},
				{version: 1, name: "過期畫面的帳戶", wantErrText: versionConflictText},
			},
			wantName:    "修改後帳戶",
			wantVersion: 2,
		},
		{
			name:        "修改不存在的 ledger_id",
			missing:     true,
			steps:       []step{{version: 1, name: "不存在的帳戶", wantErrText: versionConflictText}},
			wantName:    "版本測試帳戶",
			wantVersion: 1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newVersionFixture(t)
			ledgerID := f.ledgerID
			if tc.missing {
				ledgerID = versionMissingLedgerID
			}
			for _, s := range tc.steps {
				before := f.ledgerSnapshot(ledgerID)
				_, err := f.append(enums.AggregateAccount.Enum(), versionLedgerAccountID,
					event_types.EventLedgerAccountUpdated.Enum(), versionLedgerUpdate(ledgerID, s.name, s.version))
				expectStep(t, err, nil, s.wantErrText, before, f.ledgerSnapshot(ledgerID))
			}
			expectRow(t, f.ledgerSnapshot(f.ledgerID), tc.wantName, tc.wantVersion)
		})
	}
}

// TestLedgerAccountEvents_AccountAggregate: a ledger account's events are in
// its account's aggregate (BFS R2 D7), and its ledger_uuid is the UUID of its
// creation event, as the legacy projection derives it (D12).
func TestLedgerAccountEvents_AccountAggregate(t *testing.T) {
	f := newVersionFixture(t)
	if _, err := f.append(enums.AggregateAccount.Enum(), versionLedgerAccountID,
		event_types.EventLedgerAccountUpdated.Enum(), versionLedgerUpdate(f.ledgerID, "修改後帳戶", 1)); err != nil {
		t.Fatalf("update ledger account: %v", err)
	}

	type eventRow struct {
		EventUuid     string `db:"event_uuid"`
		EventType     string `db:"event_type"`
		AggregateType string `db:"aggregate_type"`
		AggregateID   string `db:"aggregate_id"`
		Version       int64  `db:"aggregate_version"`
	}
	var got []eventRow
	if err := f.db.Select(&got, `SELECT event_uuid, event_type, aggregate_type, aggregate_id, aggregate_version
		FROM event_store WHERE event_type LIKE 'ledger_account.%' AND merchant_id = ? ORDER BY event_id`, testMID); err != nil {
		t.Fatalf("read events: %v", err)
	}
	want := []eventRow{
		{EventUuid: f.ledgerEventUUID, EventType: "ledger_account.created", AggregateType: "ACCOUNT", AggregateID: versionLedgerAccountID, Version: 1},
		{EventType: "ledger_account.updated", AggregateType: "ACCOUNT", AggregateID: versionLedgerAccountID, Version: 2},
	}
	if len(got) != len(want) {
		t.Fatalf("events = %+v, want %d", got, len(want))
	}
	for i := range want {
		if i > 0 {
			want[i].EventUuid = got[i].EventUuid
		}
		if got[i] != want[i] {
			t.Errorf("event %d = %+v, want %+v", i, got[i], want[i])
		}
	}

	var ledgerUUID string
	if err := f.db.Get(&ledgerUUID, `SELECT ledger_uuid FROM ledger_accounts WHERE ledger_id = ? AND merchant_id = ?`,
		f.ledgerID, testMID); err != nil {
		t.Fatalf("read ledger account: %v", err)
	}
	if ledgerUUID != f.ledgerEventUUID {
		t.Errorf("ledger_uuid = %s, want the creation event %s", ledgerUUID, f.ledgerEventUUID)
	}
}

func TestInvestmentUpdatedVersion(t *testing.T) {
	type step struct {
		version int64
		name    string
		wantErr error
	}
	cases := []struct {
		name        string
		steps       []step // applied in order; earlier steps set up the state
		wantName    string
		wantVersion int64
	}{
		{
			name:        "以 version 1 修改",
			steps:       []step{{version: 1, name: "修改後投資"}},
			wantName:    "修改後投資",
			wantVersion: 2,
		},
		{
			name: "已在 version 2 時以過期的 version 1 修改",
			steps: []step{
				{version: 1, name: "修改後投資"},
				{version: 1, name: "過期畫面的投資", wantErr: projection_repo.ErrVersionConflict},
			},
			wantName:    "修改後投資",
			wantVersion: 2,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newVersionFixture(t)
			for _, s := range tc.steps {
				before := f.investmentSnapshot()
				_, err := f.append(enums.AggregateTransaction.Enum(), "", event_types.EventInvestmentUpdated.Enum(),
					payload.InvestmentUpdatedPayload{
						InvestmentCreatedPayload: versionInvestmentCreate(s.name),
						InvestmentUUID:           f.investmentUUID,
						Version:                  s.version,
					})
				expectStep(t, err, s.wantErr, "", before, f.investmentSnapshot())
			}
			expectRow(t, f.investmentSnapshot(), tc.wantName, tc.wantVersion)
		})
	}
}
