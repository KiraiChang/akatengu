package event_test

import (
	domainevent "akatengu/internal/domain/event"
	"akatengu/internal/enums"
	"akatengu/internal/model/payload"
	"akatengu/internal/shared/utils/test"
	"testing"
)

// TestConfigPayloads_LegacyJSON is the contract of BFS R2 D12 for the config
// events; see TestAccountPayloads_LegacyJSON. The account IDs follow
// seeds/account_type_configs.sql and seeds/sys_accounts.sql.
func TestConfigPayloads_LegacyJSON(t *testing.T) {
	test.InitEnums()

	t.Run("sys_config.sys_account_updated", func(t *testing.T) {
		expectSameJSON(t,
			domainevent.SysAccountUpdatedPayload{SysCode: "SYS:EQUITY:EQUITY_OPENING", Description: "期初權益", AccountID: "3101-01"},
			payload.SysAccountUpdatedPayload{SysCode: "SYS:EQUITY:EQUITY_OPENING", Description: "期初權益", AccountID: "3101-01"})
	})
	t.Run("account_config.ledger_type_updated", func(t *testing.T) {
		expectSameJSON(t,
			domainevent.LedgerAccountTypeConfigUpdatedPayload{Type: enums.LedgerAccountTypeBankAccount.Enum(), AccountID: "1101-02"},
			payload.LedgerAccountTypeConfigUpdatedPayload{Type: enums.LedgerAccountTypeBankAccount.Enum(), AccountID: "1101-02"})
	})

	oci, account := "3102-02", "1203-02"
	cases := []struct {
		name    string
		oci     *string
		account *string
	}{
		{name: "account_config.asset_type_updated", oci: &oci, account: &account},
		// account_id is omitempty and oci_account_id is not; nil checks both tags.
		{name: "account_config.asset_type_updated 的選填科目為 null"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expectSameJSON(t, domainevent.AssetTypeAccountConfigUpdatedPayload{
				AssetType: enums.AssetTypeStock.Enum(), RealizedGainAccountID: "4203-01", RealizedLossAccountID: "5402-11",
				UnrealizedGainAccountID: "4204-01", UnrealizedLossAccountID: "5402-21", OCIAccountID: tc.oci,
				FeeAccountID: "5402-01", TaxAccountID: "5402-05", AccountID: tc.account,
			}, payload.AssetTypeAccountConfigUpdatedPayload{
				AssetType: enums.AssetTypeStock.Enum(), RealizedGainAccountID: "4203-01", RealizedLossAccountID: "5402-11",
				UnrealizedGainAccountID: "4204-01", UnrealizedLossAccountID: "5402-21", OCIAccountID: tc.oci,
				FeeAccountID: "5402-01", TaxAccountID: "5402-05", AccountID: tc.account,
			})
		})
	}
}
