package services_test

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/shared/utils/test"
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

// periodRow is a period_closings row as the period projection writes it.
type periodRow struct {
	MerchantID  int64  `db:"merchant_id"`
	PeriodType  string `db:"period_type"`
	PeriodStart string `db:"period_start"`
	PeriodEnd   string `db:"period_end"`
	Status      string `db:"status"`
}

// periodState is what opening a period writes: an event, the version of the
// merchant's ACCOUNT aggregate with an empty ID, and a period row.
type periodState struct {
	Events  int   `db:"events"`
	Version int64 `db:"version"`
	Periods int   `db:"periods"`
}

func periodStateOf(t *testing.T, db *sqlx.DB, merchantID int64) periodState {
	t.Helper()
	var s periodState
	if err := db.Get(&s, `SELECT
		(SELECT COUNT(*) FROM event_store) AS events,
		(SELECT COALESCE(MAX(current_version), 0) FROM aggregate_versions
		 WHERE aggregate_type = 'ACCOUNT' AND aggregate_id = '' AND merchant_id = ?) AS version,
		(SELECT COUNT(*) FROM period_closings) AS periods`, merchantID); err != nil {
		t.Fatalf("read period state: %v", err)
	}
	return s
}

func TestPeriodStarted_PerMerchant(t *testing.T) {
	test.InitEnums()
	month := event_types.EventPeriodMonthStarted.Enum()
	annual := event_types.EventPeriodAnnualStarted.Enum()
	october := payload.PeriodMonthStartedPayload{PeriodStart: "2026-10"}
	year2026 := payload.PeriodAnnualStartedPayload{Year: 2026}

	type periodStart struct {
		merchantID int64
		eventType  event_types.EventType
		payload    any
		wantErr    string // "" = the period is opened
	}
	cases := []struct {
		name  string
		steps []periodStart
		want  []periodRow
	}{
		{
			name: "商戶 1、商戶 2 各自開立 2026-10",
			steps: []periodStart{
				{merchantID: 1, eventType: month, payload: october},
				{merchantID: 2, eventType: month, payload: october},
			},
			want: []periodRow{
				{MerchantID: 1, PeriodType: "MONTHLY", PeriodStart: "2026-10-01", PeriodEnd: "2026-10-31", Status: "OPEN"},
				{MerchantID: 2, PeriodType: "MONTHLY", PeriodStart: "2026-10-01", PeriodEnd: "2026-10-31", Status: "OPEN"},
			},
		},
		{
			name: "商戶 1、商戶 2 各自開立 2026 年度",
			steps: []periodStart{
				{merchantID: 1, eventType: annual, payload: year2026},
				{merchantID: 2, eventType: annual, payload: year2026},
			},
			want: []periodRow{
				{MerchantID: 1, PeriodType: "ANNUAL", PeriodStart: "2026-01-01", PeriodEnd: "2026-12-31", Status: "OPEN"},
				{MerchantID: 2, PeriodType: "ANNUAL", PeriodStart: "2026-01-01", PeriodEnd: "2026-12-31", Status: "OPEN"},
			},
		},
		{
			name: "同一商戶重複開立 2026-10",
			steps: []periodStart{
				{merchantID: 1, eventType: month, payload: october},
				{merchantID: 1, eventType: month, payload: october, wantErr: "UNIQUE constraint failed: period_closings."},
			},
			want: []periodRow{
				{MerchantID: 1, PeriodType: "MONTHLY", PeriodStart: "2026-10-01", PeriodEnd: "2026-10-31", Status: "OPEN"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := test.NewTestDB(t)
			// Every new connection to ":memory:" is an empty database; keep a single one.
			db.SetMaxOpenConns(1)
			svc := newSvc(db)

			for i, step := range tc.steps {
				before := periodStateOf(t, db, step.merchantID)
				ctx := context.WithValue(context.Background(), ctxkey.MerchantID, step.merchantID)
				// Same envelope as frontside/src/api/period.ts: ('ACCOUNT', '') at its current version.
				_, err := svc.Append(ctx, cmd.AppendCmd{
					AggregateType:   enums.AggregateAccount.Enum(),
					AggregateID:     "",
					ExpectedVersion: before.Version,
					EventType:       step.eventType,
					Payload:         mustMarshal(t, step.payload),
				})
				after := periodStateOf(t, db, step.merchantID)

				if step.wantErr != "" {
					if err == nil || !strings.Contains(err.Error(), step.wantErr) {
						t.Fatalf("step %d: err = %v, want %q", i, err, step.wantErr)
					}
					if after != before {
						t.Fatalf("step %d: state changed: before %+v, after %+v", i, before, after)
					}
					continue
				}
				if err != nil {
					t.Fatalf("step %d: Append: %v", i, err)
				}
				want := periodState{Events: before.Events + 1, Version: before.Version + 1, Periods: before.Periods + 1}
				if after != want {
					t.Fatalf("step %d: state = %+v, want %+v", i, after, want)
				}
			}

			var got []periodRow
			if err := db.Select(&got, `SELECT merchant_id, period_type, period_start, period_end, status
				FROM period_closings ORDER BY merchant_id, period_type, period_start`); err != nil {
				t.Fatalf("read periods: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("periods = %+v, want %+v", got, tc.want)
			}
		})
	}
}
