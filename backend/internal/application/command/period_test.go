package command_test

import (
	"akatengu/internal/application/command"
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	kevent "akatengu/internal/kernel/event"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/shared/utils/test"
	"encoding/json"
	"errors"
	"testing"
)

const cmdOctoberPayload = `{"period_start":"2026-10"}`

// TestDispatchLegacy_PeriodStarted: the opening is written to the merchant's
// PERIOD aggregate at its current version + 1, whatever version the frontend
// sends with ("ACCOUNT", "") (BFS R2 section 3, in_tx).
func TestDispatchLegacy_PeriodStarted(t *testing.T) {
	db := newCommandTestDB(t)
	// The PERIOD aggregate is already at version 3; the frontend reads
	// ('ACCOUNT', ''), which has no version.
	if _, err := db.Exec(`INSERT INTO aggregate_versions(aggregate_type, aggregate_id, merchant_id, current_version)
		VALUES('PERIOD', '1', ?, 3)`, cmdMerchantID); err != nil {
		t.Fatalf("seed version: %v", err)
	}

	got, handled, err := command.NewWriter(db, command.NewEngine(db)).DispatchLegacy(cmdCtx(cmdMerchantID),
		cmd.AppendCmd{AggregateType: enums.AggregateAccount.Enum(), AggregateID: "", ExpectedVersion: 0,
			EventType: event_types.EventPeriodMonthStarted.Enum(), Payload: json.RawMessage(cmdOctoberPayload)},
		kevent.EventParams{Source: "api"})

	if !handled || err != nil {
		t.Fatalf("handled = %v, err = %v; want written through BFS", handled, err)
	}
	if got.AggregateType != enums.AggregatePeriod.Enum() || got.AggregateId != "1" || got.AggregateVersion != 4 ||
		got.MerchantID != cmdMerchantID {
		t.Errorf("record = %+v, want period.month_started at (PERIOD, 1) version 4", got)
	}
	if got.UpdatedBy == nil || *got.UpdatedBy != cmdUserName {
		t.Errorf("UpdatedBy = %v, want %q", got.UpdatedBy, cmdUserName)
	}
	var meta kevent.Metadata
	if err := json.Unmarshal(got.Metadata, &meta); err != nil || meta.Execution.Source != "api" {
		t.Errorf("metadata = %s (err %v), want source api", got.Metadata, err)
	}
	var period struct {
		ClosingUUID string `db:"closing_uuid"`
		Status      string `db:"status"`
	}
	if err := db.Get(&period, `SELECT closing_uuid, status FROM period_closings
		WHERE merchant_id = ? AND period_type = 'MONTHLY' AND period_start = '2026-10-01'`, cmdMerchantID); err != nil {
		t.Fatalf("read period: %v", err)
	}
	if period.ClosingUUID != got.EventUuid || period.Status != "OPEN" {
		t.Errorf("period = %+v, want OPEN and named after the event %s", period, got.EventUuid)
	}
}

func TestDispatchLegacy_PeriodStartedRejected(t *testing.T) {
	test.InitEnums()
	cases := []struct {
		name    string
		aggType enums.AggregateType
		aggID   string
		et      event_types.EventType
		payload string
	}{
		{"ACCOUNT 的 aggregate ID 不是空字串", enums.AggregateAccount.Enum(), "1",
			event_types.EventPeriodMonthStarted.Enum(), cmdOctoberPayload},
		{"直接送出推導的 ('PERIOD', \"1\")", enums.AggregatePeriod.Enum(), "1",
			event_types.EventPeriodMonthStarted.Enum(), cmdOctoberPayload},
		{"年度開帳送到 TRANSACTION", enums.AggregateTransaction.Enum(), "",
			event_types.EventPeriodAnnualStarted.Enum(), `{"year":2026}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := newCommandTestDB(t)
			before := accountCountsOf(t, db)

			got, handled, err := command.NewWriter(db, command.NewEngine(db)).DispatchLegacy(cmdCtx(cmdMerchantID),
				cmd.AppendCmd{AggregateType: tc.aggType, AggregateID: tc.aggID, EventType: tc.et, Payload: json.RawMessage(tc.payload)},
				kevent.EventParams{Source: "api"})

			if !handled || !errors.Is(err, command.ErrLegacyAggregateMismatch) {
				t.Errorf("handled = %v, err = %v; want %v", handled, err, command.ErrLegacyAggregateMismatch)
			}
			if got != nil {
				t.Errorf("result = %+v, want nil", got)
			}
			if after := accountCountsOf(t, db); after != before {
				t.Errorf("rows changed: before %+v, after %+v", before, after)
			}
			var periods int
			if err := db.Get(&periods, `SELECT COUNT(*) FROM period_closings`); err != nil || periods != 0 {
				t.Errorf("periods = %d (err %v), want none", periods, err)
			}
		})
	}
}
