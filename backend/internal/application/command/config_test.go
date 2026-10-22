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

// TestDispatchLegacy_ConfigRejected: the config events have no legacy
// aggregate, so POST /api/event/append does not accept them, even with the
// aggregate the dedicated APIs write to (BFS R2 section 3).
func TestDispatchLegacy_ConfigRejected(t *testing.T) {
	test.InitEnums()
	cases := []struct {
		name    string
		aggType enums.AggregateType
		et      event_types.EventType
		payload string
	}{
		{
			name:    "sys_config.sys_account_updated",
			aggType: enums.AggregateSysConfig.Enum(),
			et:      event_types.EventSysAccountUpdated.Enum(),
			payload: `{"sys_code":"SYS:EQUITY:EQUITY_OPENING","description":"期初權益","account_id":"3101-01"}`,
		},
		{
			name:    "account_config.ledger_type_updated",
			aggType: enums.AggregateAccountConfig.Enum(),
			et:      event_types.EventLedgerAccountTypeConfigUpdated.Enum(),
			payload: `{"type":"BANK_ACCOUNT","account_id":"1101-02"}`,
		},
		{
			name:    "account_config.asset_type_updated",
			aggType: enums.AggregateAccountConfig.Enum(),
			et:      event_types.EventAssetTypeAccountConfigUpdated.Enum(),
			payload: `{"asset_type":"STOCK","realized_gain_account_id":"4203-01","realized_loss_account_id":"5402-11",` +
				`"unrealized_gain_account_id":"4204-01","unrealized_loss_account_id":"5402-21",` +
				`"oci_account_id":null,"fee_account_id":"5402-01","tax_account_id":"5402-05"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := newCommandTestDB(t)
			before := accountCountsOf(t, db)

			got, handled, err := command.NewWriter(db, command.NewEngine(db)).DispatchLegacy(cmdCtx(cmdMerchantID),
				cmd.AppendCmd{AggregateType: tc.aggType, AggregateID: "1", EventType: tc.et, Payload: json.RawMessage(tc.payload)},
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
