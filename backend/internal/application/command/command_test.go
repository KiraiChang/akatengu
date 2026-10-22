package command_test

import (
	"akatengu/internal/application/command"
	domainevent "akatengu/internal/domain/event"
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	kerrors "akatengu/internal/kernel/errors"
	kevent "akatengu/internal/kernel/event"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/pkg/jwt"
	"akatengu/internal/shared/utils/test"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

// Fixture identifiers: one rent payment with a debit and a credit entry. The
// accounts come from seeds/accounts.sql.
const (
	cmdTxnUUID     = "0192f3f4-0000-7000-8000-000000000001"
	cmdDebitEntry  = "0192f3f4-0000-7000-8000-000000000011"
	cmdCreditEntry = "0192f3f4-0000-7000-8000-000000000012"
	cmdMerchantID  = int64(1)
	cmdUserName    = "r2-tester"
	cmdRequestID   = "r2-request"
)

func newCommandTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db := test.NewTestDB(t)
	// Every new connection to ":memory:" is an empty database; keep a single one.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`INSERT INTO transactions(txn_id, txn_uuid, merchant_id, txn_date, description, total_amount, currency, status, version)
		VALUES(1, ?, ?, '2026-10-01', 'r2 command rent', 1000, 'TWD', 'ACTIVE', 1)`, cmdTxnUUID, cmdMerchantID); err != nil {
		t.Fatalf("insert transaction: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO journal_entries(entry_uuid, txn_uuid, merchant_id, txn_id, account_id, debit, credit)
		VALUES(?, ?, ?, 1, '5101-01', 1000, 0), (?, ?, ?, 1, '1101-02', 0, 1000)`,
		cmdDebitEntry, cmdTxnUUID, cmdMerchantID, cmdCreditEntry, cmdTxnUUID, cmdMerchantID); err != nil {
		t.Fatalf("insert journal entries: %v", err)
	}
	return db
}

// cmdCtx is a request context after the Jwt, MerchantAuth and RequestID
// middlewares; merchantID 0 leaves the merchant out.
func cmdCtx(merchantID int64) context.Context {
	ctx := context.WithValue(context.Background(), ctxkey.UserClaims, &jwt.Claims{UserName: cmdUserName})
	ctx = context.WithValue(ctx, ctxkey.RequestID, cmdRequestID)
	if merchantID != 0 {
		ctx = context.WithValue(ctx, ctxkey.MerchantID, merchantID)
	}
	return ctx
}

func cfPayload() domainevent.TransactionCFCategoryUpdatedPayload {
	return domainevent.TransactionCFCategoryUpdatedPayload{
		TxnUUID: cmdTxnUUID,
		Entries: []domainevent.TransactionCFEntryItem{
			{EntryUUID: cmdDebitEntry, CFCategory: enums.CashFlowCategoryInvesting.Enum()},
		},
	}
}

// unregisteredPayload is a payload type the registry does not list.
type unregisteredPayload struct{}

func (unregisteredPayload) EventType() event_types.EventType {
	return event_types.EventTransactionCreated.Enum()
}

// writeCounts are the rows a cash flow category write touches.
type writeCounts struct {
	Events     int `db:"events"`
	Categories int `db:"categories"`
}

func writeCountsOf(t *testing.T, db *sqlx.DB) writeCounts {
	t.Helper()
	var c writeCounts
	if err := db.Get(&c, `SELECT
		(SELECT COUNT(*) FROM event_store) AS events,
		(SELECT COUNT(*) FROM entry_cf_categories) AS categories`); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	return c
}

func TestBuild_CFCategoryUpdated(t *testing.T) {
	test.InitEnums()

	evt, err := command.Build(cmdCtx(cmdMerchantID), cfPayload(), 4, kevent.EventParams{Source: "api"})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if evt.Uuid == "" {
		t.Error("Uuid is empty")
	}
	if evt.EventType != event_types.EventTransactionCFCategoryUpdated.Enum() {
		t.Errorf("EventType = %v", evt.EventType)
	}
	if evt.AggregateType != enums.AggregateTransaction.Enum() || evt.AggregateUuid != cmdTxnUUID {
		t.Errorf("aggregate = (%v, %s), want (TRANSACTION, %s)", evt.AggregateType, evt.AggregateUuid, cmdTxnUUID)
	}
	if evt.MerchantID != cmdMerchantID || evt.UpdatedBy != cmdUserName {
		t.Errorf("MerchantID = %d, UpdatedBy = %q; want %d, %q", evt.MerchantID, evt.UpdatedBy, cmdMerchantID, cmdUserName)
	}
	// The cash flow category policy validates the client's version.
	if evt.RootVersion != kevent.RootVersionExpected || evt.ExpectedVersion != 4 {
		t.Errorf("RootVersion = %v, ExpectedVersion = %d; want RootVersionExpected, 4", evt.RootVersion, evt.ExpectedVersion)
	}
	if evt.Metadata.Tracing.RequestID != cmdRequestID || evt.Metadata.Tracing.CorrelationID == "" {
		t.Errorf("tracing = %+v, want request %q and a correlation ID", evt.Metadata.Tracing, cmdRequestID)
	}
	if evt.Metadata.Execution.PartitionKey != cmdTxnUUID || evt.Metadata.Execution.Source != "api" {
		t.Errorf("execution = %+v, want partition %q and source api", evt.Metadata.Execution, cmdTxnUUID)
	}
	if !reflect.DeepEqual(evt.Payload, cfPayload()) {
		t.Errorf("Payload = %#v, want %#v", evt.Payload, cfPayload())
	}
}

func TestBuild_Rejected(t *testing.T) {
	test.InitEnums()
	cases := []struct {
		name     string
		ctx      context.Context
		payload  kevent.Payload
		wantCode kerrors.EventErrorCode // "" = any error
	}{
		{
			name:     "未登錄的事件類型",
			ctx:      cmdCtx(cmdMerchantID),
			payload:  unregisteredPayload{},
			wantCode: kerrors.ErrUnknownEventType,
		},
		{
			name:    "ctx 沒有商戶",
			ctx:     cmdCtx(0),
			payload: cfPayload(),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := command.Build(tc.ctx, tc.payload, 0, kevent.EventParams{Source: "api"})
			if err == nil {
				t.Fatal("Build succeeded")
			}
			if tc.wantCode != "" {
				var ee kerrors.EventError
				if !errors.As(err, &ee) || ee.Code != tc.wantCode {
					t.Fatalf("err = %v, want code %s", err, tc.wantCode)
				}
			}
		})
	}
}

func TestWriter_Append(t *testing.T) {
	test.InitEnums()
	cases := []struct {
		name         string
		seedVersion  int64 // aggregate_versions of the transaction before the write; 0 = none
		expected     int64
		wantCode     kerrors.EventErrorCode // "" = written
		wantVersion  int64
		wantCategory int
	}{
		{
			name:         "版本相符：寫入並回讀 event_store 的紀錄",
			wantVersion:  1,
			wantCategory: 1,
		},
		{
			name:        "過期的版本：回傳 Run 的錯誤，不寫入",
			seedVersion: 2,
			expected:    1,
			wantCode:    kerrors.ErrVersionConflict,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := newCommandTestDB(t)
			if tc.seedVersion != 0 {
				if _, err := db.Exec(`INSERT INTO aggregate_versions(aggregate_type, aggregate_id, merchant_id, current_version)
					VALUES('TRANSACTION', ?, ?, ?)`, cmdTxnUUID, cmdMerchantID, tc.seedVersion); err != nil {
					t.Fatalf("seed version: %v", err)
				}
			}
			ctx := cmdCtx(cmdMerchantID)
			evt, err := command.Build(ctx, cfPayload(), tc.expected, kevent.EventParams{Source: "api"})
			if err != nil {
				t.Fatalf("Build: %v", err)
			}

			got, err := command.NewWriter(db, command.NewEngine(db)).Append(ctx, evt)

			counts := writeCountsOf(t, db)
			if tc.wantCode != "" {
				var ee kerrors.EventError
				if !errors.As(err, &ee) || ee.Code != tc.wantCode {
					t.Fatalf("err = %v, want code %s", err, tc.wantCode)
				}
				if got != nil || counts != (writeCounts{}) {
					t.Fatalf("result = %+v, rows = %+v; want nothing written", got, counts)
				}
				return
			}
			if err != nil {
				t.Fatalf("Append: %v", err)
			}
			if got.EventUuid != evt.Uuid || got.EventType != evt.EventType ||
				got.AggregateType != enums.AggregateTransaction.Enum() || got.AggregateId != cmdTxnUUID ||
				got.AggregateVersion != tc.wantVersion || got.MerchantID != cmdMerchantID {
				t.Errorf("record = %+v, want event %s at (TRANSACTION, %s) version %d", got, evt.Uuid, cmdTxnUUID, tc.wantVersion)
			}
			if got.UpdatedBy == nil || *got.UpdatedBy != cmdUserName {
				t.Errorf("UpdatedBy = %v, want %q", got.UpdatedBy, cmdUserName)
			}
			var meta kevent.Metadata
			if err := json.Unmarshal(got.Metadata, &meta); err != nil || meta.Tracing.RequestID != cmdRequestID {
				t.Errorf("metadata = %s (err %v), want request %q", got.Metadata, err, cmdRequestID)
			}
			if counts.Events != 1 || counts.Categories != tc.wantCategory {
				t.Errorf("rows = %+v, want 1 event and %d categories", counts, tc.wantCategory)
			}
		})
	}
}

func TestWriter_DispatchLegacy(t *testing.T) {
	test.InitEnums()
	cfJSON, err := json.Marshal(cfPayload())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	cases := []struct {
		name        string
		req         cmd.AppendCmd
		wantHandled bool
		wantErr     error // errors.Is target; nil = no error
	}{
		{
			name: "未登錄的事件類型交回舊路徑",
			req: cmd.AppendCmd{AggregateType: enums.AggregateTransaction.Enum(), AggregateID: "",
				EventType: event_types.EventInvestmentCreated.Enum(), Payload: json.RawMessage(`{"account_id":"1203-02"}`)},
		},
		{
			name: "已登錄、沒有舊映射的 cf 被拒絕",
			req: cmd.AppendCmd{AggregateType: enums.AggregateTransaction.Enum(), AggregateID: cmdTxnUUID,
				EventType: event_types.EventTransactionCFCategoryUpdated.Enum(), Payload: cfJSON},
			wantHandled: true,
			wantErr:     command.ErrLegacyNotAccepted,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := newCommandTestDB(t)

			got, handled, err := command.NewWriter(db, command.NewEngine(db)).
				DispatchLegacy(cmdCtx(cmdMerchantID), tc.req, kevent.EventParams{Source: "api"})

			if handled != tc.wantHandled {
				t.Errorf("handled = %v, want %v", handled, tc.wantHandled)
			}
			if tc.wantErr == nil && err != nil || tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Errorf("err = %v, want %v", err, tc.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), tc.req.EventType.String()) {
				t.Errorf("err = %v, want the event type in it", err)
			}
			if got != nil {
				t.Errorf("result = %+v, want nil", got)
			}
			if counts := writeCountsOf(t, db); counts != (writeCounts{}) {
				t.Errorf("rows = %+v, want nothing written", counts)
			}
		})
	}
}
