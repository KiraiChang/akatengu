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

// cmdCategoryUUID is a category the requests name; it need not exist, the
// dispatch rejects the request first.
const cmdCategoryUUID = "0192f3f4-0000-7000-8000-000000000901"

// TestDispatchLegacy_CategoryRejected: the category events have no legacy
// aggregate, so POST /api/event/append does not accept them; only the
// dedicated APIs write them (BFS R2 section 4).
func TestDispatchLegacy_CategoryRejected(t *testing.T) {
	test.InitEnums()
	const (
		prepaidCreated = `{"name":"保險費","account_id":"1104-01","expense_account_id":"5201-01"}`
		prepaidUpdated = `{"category_uuid":"` + cmdCategoryUUID + `","name":"保險費","account_id":"1104-01","expense_account_id":"5201-01"}`
		fixedCreated   = `{"name":"電腦設備","asset_account_id":"1601-01","accum_depreciation_account_id":"1601-02",` +
			`"depreciation_expense_account_id":"5601-01"}`
		fixedUpdated = `{"category_uuid":"` + cmdCategoryUUID + `","name":"電腦設備","asset_account_id":"1601-01",` +
			`"accum_depreciation_account_id":"1601-02","depreciation_expense_account_id":"5601-01"}`
		deleted = `{"category_uuid":"` + cmdCategoryUUID + `"}`
	)
	cases := []struct {
		aggType enums.AggregateType
		et      event_types.EventType
		payload string
	}{
		{enums.AggregatePrepaidCategory.Enum(), event_types.EventPrepaidCategoryCreated.Enum(), prepaidCreated},
		{enums.AggregatePrepaidCategory.Enum(), event_types.EventPrepaidCategoryUpdated.Enum(), prepaidUpdated},
		{enums.AggregatePrepaidCategory.Enum(), event_types.EventPrepaidCategoryDeleted.Enum(), deleted},
		{enums.AggregateFixedAssetCategory.Enum(), event_types.EventFixedAssetCategoryCreated.Enum(), fixedCreated},
		{enums.AggregateFixedAssetCategory.Enum(), event_types.EventFixedAssetCategoryUpdated.Enum(), fixedUpdated},
		{enums.AggregateFixedAssetCategory.Enum(), event_types.EventFixedAssetCategoryDeleted.Enum(), deleted},
	}
	for _, tc := range cases {
		t.Run(tc.et.String(), func(t *testing.T) {
			db := newCommandTestDB(t)
			before := accountCountsOf(t, db)

			got, handled, err := command.NewWriter(db, command.NewEngine(db)).DispatchLegacy(cmdCtx(cmdMerchantID),
				cmd.AppendCmd{AggregateType: tc.aggType, AggregateID: cmdCategoryUUID, EventType: tc.et, Payload: json.RawMessage(tc.payload)},
				kevent.EventParams{Source: "api"})

			if !handled || !errors.Is(err, command.ErrLegacyNotAccepted) {
				t.Errorf("handled = %v, err = %v; want %v", handled, err, command.ErrLegacyNotAccepted)
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
