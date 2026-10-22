package command_test

import (
	"akatengu/internal/application/command"
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	kerrors "akatengu/internal/kernel/errors"
	kevent "akatengu/internal/kernel/event"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/shared/utils/test"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

// cmdNewAccountID is not in seeds/accounts.sql; 1101 is.
const (
	cmdNewAccountID  = "1101-92"
	cmdSeedAccountID = "1101"
)

// accountCreatedRequest is the generic API request for a new account, as
// frontside/src/api/account.ts sends it.
func accountCreatedRequest(t *testing.T, aggType enums.AggregateType, aggID, accountID string) cmd.AppendCmd {
	t.Helper()
	raw, err := json.Marshal(map[string]any{
		"account_id": accountID, "parent_id": "1101", "name": "R2 段 2 科目",
		"type": "ASSET", "normal_balance": "DEBIT", "currency": "TWD", "is_active": true,
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return cmd.AppendCmd{AggregateType: aggType, AggregateID: aggID, ExpectedVersion: 0,
		EventType: event_types.EventAccountCreated.Enum(), Payload: raw}
}

// accountCounts are the rows an account write touches.
type accountCounts struct {
	Events   int `db:"events"`
	Versions int `db:"versions"`
	Accounts int `db:"accounts"`
}

func accountCountsOf(t *testing.T, db *sqlx.DB) accountCounts {
	t.Helper()
	var c accountCounts
	if err := db.Get(&c, `SELECT
		(SELECT COUNT(*) FROM event_store) AS events,
		(SELECT COUNT(*) FROM aggregate_versions) AS versions,
		(SELECT COUNT(*) FROM accounts WHERE merchant_id = ?) AS accounts`, cmdMerchantID); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	return c
}

func TestDispatchLegacy_AccountCreated(t *testing.T) {
	db := newCommandTestDB(t)
	ctx := cmdCtx(cmdMerchantID)
	before := accountCountsOf(t, db)

	got, handled, err := command.NewWriter(db, command.NewEngine(db)).DispatchLegacy(ctx,
		accountCreatedRequest(t, enums.AggregateAccount.Enum(), cmdNewAccountID, cmdNewAccountID),
		kevent.EventParams{Source: "api"})

	if !handled || err != nil {
		t.Fatalf("handled = %v, err = %v; want written through BFS", handled, err)
	}
	if got.EventType != event_types.EventAccountCreated.Enum() || got.AggregateType != enums.AggregateAccount.Enum() ||
		got.AggregateId != cmdNewAccountID || got.AggregateVersion != 1 || got.MerchantID != cmdMerchantID {
		t.Errorf("record = %+v, want account.created at (ACCOUNT, %s) version 1", got, cmdNewAccountID)
	}
	if got.UpdatedBy == nil || *got.UpdatedBy != cmdUserName {
		t.Errorf("UpdatedBy = %v, want %q", got.UpdatedBy, cmdUserName)
	}
	var meta kevent.Metadata
	if err := json.Unmarshal(got.Metadata, &meta); err != nil ||
		meta.Tracing.RequestID != cmdRequestID || meta.Execution.Source != "api" {
		t.Errorf("metadata = %s (err %v), want request %q and source api", got.Metadata, err, cmdRequestID)
	}

	var row struct {
		Name    string `db:"name"`
		Version int64  `db:"version"`
	}
	if err := db.Get(&row, `SELECT name, version FROM accounts WHERE account_id = ? AND merchant_id = ?`,
		cmdNewAccountID, cmdMerchantID); err != nil {
		t.Fatalf("read account: %v", err)
	}
	if row.Name != "R2 段 2 科目" || row.Version != 1 {
		t.Errorf("account = %+v, want the new account at version 1", row)
	}
	after := accountCountsOf(t, db)
	if after != (accountCounts{Events: before.Events + 1, Versions: before.Versions + 1, Accounts: before.Accounts + 1}) {
		t.Errorf("rows: before %+v, after %+v; want one more event, version record and account", before, after)
	}
}

func TestDispatchLegacy_AccountCreatedRejected(t *testing.T) {
	test.InitEnums()
	// Exactly one of wantCode, wantErr and wantErrText is set.
	cases := []struct {
		name        string
		aggType     enums.AggregateType
		aggID       string
		accountID   string
		seedVersion int64 // version record of ('ACCOUNT', accountID) before the request; 0 = none
		wantCode    kerrors.EventErrorCode
		wantErr     error
		wantErrText string
	}{
		{
			name:    "aggregate 已存在：新建的版本策略回傳衝突，不看前端送出的版本",
			aggType: enums.AggregateAccount.Enum(), aggID: cmdNewAccountID, accountID: cmdNewAccountID,
			seedVersion: 2,
			wantCode:    kerrors.ErrVersionConflict,
		},
		{
			name:    "科目編號已在讀模型、沒有 aggregate（種子科目）：唯一鍵擋下並回復",
			aggType: enums.AggregateAccount.Enum(), aggID: cmdSeedAccountID, accountID: cmdSeedAccountID,
			wantErrText: "UNIQUE constraint failed",
		},
		{
			name:    "aggregate ID 與 payload.account_id 不同",
			aggType: enums.AggregateAccount.Enum(), aggID: cmdSeedAccountID, accountID: cmdNewAccountID,
			wantErr: command.ErrLegacyAggregateMismatch,
		},
		{
			name:    "aggregate 類型不是 ACCOUNT",
			aggType: enums.AggregateTransaction.Enum(), aggID: cmdNewAccountID, accountID: cmdNewAccountID,
			wantErr: command.ErrLegacyAggregateMismatch,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := newCommandTestDB(t)
			if tc.seedVersion != 0 {
				if _, err := db.Exec(`INSERT INTO aggregate_versions(aggregate_type, aggregate_id, merchant_id, current_version)
					VALUES('ACCOUNT', ?, ?, ?)`, tc.accountID, cmdMerchantID, tc.seedVersion); err != nil {
					t.Fatalf("seed version: %v", err)
				}
			}
			before := accountCountsOf(t, db)

			got, handled, err := command.NewWriter(db, command.NewEngine(db)).DispatchLegacy(cmdCtx(cmdMerchantID),
				accountCreatedRequest(t, tc.aggType, tc.aggID, tc.accountID), kevent.EventParams{Source: "api"})

			if !handled {
				t.Fatal("handed to the legacy path, want handled by the BFS dispatch")
			}
			var ee kerrors.EventError
			switch {
			case err == nil:
				t.Fatal("dispatch succeeded")
			case tc.wantCode != "" && (!errors.As(err, &ee) || ee.Code != tc.wantCode):
				t.Errorf("err = %v, want code %s", err, tc.wantCode)
			case tc.wantErr != nil && !errors.Is(err, tc.wantErr):
				t.Errorf("err = %v, want %v", err, tc.wantErr)
			case tc.wantErrText != "" && !strings.Contains(err.Error(), tc.wantErrText):
				t.Errorf("err = %v, want an error containing %q", err, tc.wantErrText)
			}
			if got != nil {
				t.Errorf("result = %+v, want nil", got)
			}
			if after := accountCountsOf(t, db); after != before {
				t.Errorf("rows changed: before %+v, after %+v", before, after)
			}
		})
	}
}
