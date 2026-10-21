package handler

import (
	"net/http"
)

// Account IDs below are taken from seeds/accounts.sql; the STOCK accounts and the
// system account mapping follow seeds/account_type_configs.sql and seeds/sys_accounts.sql.
const (
	settingStockBody = `{"realized_gain_account_id":"4203-01","realized_loss_account_id":"5402-11",` +
		`"unrealized_gain_account_id":"4204-01","unrealized_loss_account_id":"5402-21",` +
		`"oci_account_id":null,"fee_account_id":"5402-01","tax_account_id":"5402-05","account_id":null}`
	settingStockPayload = `{"asset_type":"STOCK","realized_gain_account_id":"4203-01","realized_loss_account_id":"5402-11",` +
		`"unrealized_gain_account_id":"4204-01","unrealized_loss_account_id":"5402-21",` +
		`"oci_account_id":null,"fee_account_id":"5402-01","tax_account_id":"5402-05"}`
	settingStockBodyWithBond = `{"asset_type":"BOND","realized_gain_account_id":"4203-01","realized_loss_account_id":"5402-11",` +
		`"unrealized_gain_account_id":"4204-01","unrealized_loss_account_id":"5402-21",` +
		`"oci_account_id":null,"fee_account_id":"5402-01","tax_account_id":"5402-05","account_id":null}`
	settingStockBodyNoFee = `{"realized_gain_account_id":"4203-01","realized_loss_account_id":"5402-11",` +
		`"unrealized_gain_account_id":"4204-01","unrealized_loss_account_id":"5402-21",` +
		`"oci_account_id":null,"tax_account_id":"5402-05","account_id":null}`
	settingSysAccountBody = `{"sys_code":"SYS:EQUITY:EQUITY_OPENING","description":"期初權益","account_id":"3101-01"}`
)

// settingRequest is one call to a config write API. path has no /api prefix.
type settingRequest struct {
	method string
	path   string
	body   string
}

// settingEvent is an event expected in event_store; aggregate_id is always "1".
type settingEvent struct {
	aggregateType string
	eventType     string
	version       int64
	payload       string // compared with MatchJSON
}

type assetTypeConfigRow struct {
	RealizedGainAccountID   string  `db:"realized_gain_account_id"`
	RealizedLossAccountID   string  `db:"realized_loss_account_id"`
	UnrealizedGainAccountID string  `db:"unrealized_gain_account_id"`
	UnrealizedLossAccountID string  `db:"unrealized_loss_account_id"`
	OCIAccountID            *string `db:"oci_account_id"`
	FeeAccountID            string  `db:"fee_account_id"`
	TaxAccountID            string  `db:"tax_account_id"`
	AccountID               *string `db:"account_id"`
}

type sysAccountRow struct {
	Description string `db:"description"`
	AccountID   string `db:"account_id"`
}

var settingStockRow = &assetTypeConfigRow{
	RealizedGainAccountID:   "4203-01",
	RealizedLossAccountID:   "5402-11",
	UnrealizedGainAccountID: "4204-01",
	UnrealizedLossAccountID: "5402-21",
	FeeAccountID:            "5402-01",
	TaxAccountID:            "5402-05",
}

type settingScenario struct {
	given           string
	when            string
	then            string
	existingVersion int64            // version of ('ACCOUNT_CONFIG', "1") stored before the requests (0 = none)
	requests        []settingRequest // sent in order
	wantCode        int              // expected status code of every request
	wantEvents      []settingEvent   // every event in event_store after the requests, in order
	// Read models after the requests; a zero value means the row must not exist.
	wantLedgerTypes map[string]string              // type -> account_id in ledger_account_type_config
	wantAssetTypes  map[string]*assetTypeConfigRow // asset_type -> row in asset_type_account_config
	wantSysAccounts map[string]*sysAccountRow      // sys_code -> row in sys_accounts
}

var settingScenarios = []settingScenario{
	{
		given:    "ACCOUNT_CONFIG 尚無版本紀錄",
		when:     "PUT /setting/ledger-account-type/BANK_ACCOUNT 帶 account_id 1101-02",
		then:     "回傳 200，寫入版本 1 的事件，帳戶類型設定更新為 1101-02",
		requests: []settingRequest{{http.MethodPut, "/setting/ledger-account-type/BANK_ACCOUNT", `{"account_id":"1101-02"}`}},
		wantCode: http.StatusOK,
		wantEvents: []settingEvent{
			{"ACCOUNT_CONFIG", "account_config.ledger_type_updated", 1, `{"type":"BANK_ACCOUNT","account_id":"1101-02"}`},
		},
		wantLedgerTypes: map[string]string{"BANK_ACCOUNT": "1101-02"},
	},
	{
		given:           "ACCOUNT_CONFIG 已在版本 3（已有種子事件的商戶）",
		when:            "PUT /setting/ledger-account-type/BANK_ACCOUNT 帶 account_id 1101-02",
		then:            "回傳 200，寫入版本 4 的事件",
		existingVersion: 3,
		requests:        []settingRequest{{http.MethodPut, "/setting/ledger-account-type/BANK_ACCOUNT", `{"account_id":"1101-02"}`}},
		wantCode:        http.StatusOK,
		wantEvents: []settingEvent{
			{"ACCOUNT_CONFIG", "account_config.ledger_type_updated", 4, `{"type":"BANK_ACCOUNT","account_id":"1101-02"}`},
		},
		wantLedgerTypes: map[string]string{"BANK_ACCOUNT": "1101-02"},
	},
	{
		given:    "ACCOUNT_CONFIG 尚無版本紀錄",
		when:     "PUT /setting/ledger-account-type/BANK_ACCOUNT，body 另帶不存在的 type SAVINGS",
		then:     "以路徑為準：回傳 200，寫入 type 為 BANK_ACCOUNT 的事件",
		requests: []settingRequest{{http.MethodPut, "/setting/ledger-account-type/BANK_ACCOUNT", `{"type":"SAVINGS","account_id":"1101-02"}`}},
		wantCode: http.StatusOK,
		wantEvents: []settingEvent{
			{"ACCOUNT_CONFIG", "account_config.ledger_type_updated", 1, `{"type":"BANK_ACCOUNT","account_id":"1101-02"}`},
		},
		wantLedgerTypes: map[string]string{"BANK_ACCOUNT": "1101-02"},
	},
	{
		given:    "ACCOUNT_CONFIG 尚無版本紀錄",
		when:     "PUT /setting/ledger-account-type/SAVINGS（不存在的帳戶類型）",
		then:     "回傳 400，不寫入事件",
		requests: []settingRequest{{http.MethodPut, "/setting/ledger-account-type/SAVINGS", `{"account_id":"1101-02"}`}},
		wantCode: http.StatusBadRequest,
	},
	{
		given:           "ACCOUNT_CONFIG 尚無版本紀錄",
		when:            "PUT /setting/ledger-account-type/LOAN 帶不存在的科目 9999-99",
		then:            "回傳 400，不寫入事件，設定表沒有 LOAN",
		requests:        []settingRequest{{http.MethodPut, "/setting/ledger-account-type/LOAN", `{"account_id":"9999-99"}`}},
		wantCode:        http.StatusBadRequest,
		wantLedgerTypes: map[string]string{"LOAN": ""},
	},
	{
		given:           "ACCOUNT_CONFIG 尚無版本紀錄",
		when:            "PUT /setting/ledger-account-type/LOAN 帶空的 body {}",
		then:            "回傳 400，不寫入事件，設定表沒有 LOAN",
		requests:        []settingRequest{{http.MethodPut, "/setting/ledger-account-type/LOAN", `{}`}},
		wantCode:        http.StatusBadRequest,
		wantLedgerTypes: map[string]string{"LOAN": ""},
	},
	{
		given:    "ACCOUNT_CONFIG 尚無版本紀錄",
		when:     "PUT /setting/asset-type/STOCK 帶種子的 STOCK 科目，oci_account_id 與 account_id 為 null",
		then:     "回傳 200，寫入版本 1 的事件，資產類型設定與送出的值相同",
		requests: []settingRequest{{http.MethodPut, "/setting/asset-type/STOCK", settingStockBody}},
		wantCode: http.StatusOK,
		wantEvents: []settingEvent{
			{"ACCOUNT_CONFIG", "account_config.asset_type_updated", 1, settingStockPayload},
		},
		wantAssetTypes: map[string]*assetTypeConfigRow{"STOCK": settingStockRow},
	},
	{
		given:    "ACCOUNT_CONFIG 尚無版本紀錄",
		when:     "PUT /setting/asset-type/STOCK，body 另帶不存在的 asset_type BOND",
		then:     "以路徑為準：回傳 200，寫入 asset_type 為 STOCK 的事件",
		requests: []settingRequest{{http.MethodPut, "/setting/asset-type/STOCK", settingStockBodyWithBond}},
		wantCode: http.StatusOK,
		wantEvents: []settingEvent{
			{"ACCOUNT_CONFIG", "account_config.asset_type_updated", 1, settingStockPayload},
		},
		wantAssetTypes: map[string]*assetTypeConfigRow{"STOCK": settingStockRow},
	},
	{
		given:          "ACCOUNT_CONFIG 尚無版本紀錄",
		when:           "PUT /setting/asset-type/STOCK 缺 fee_account_id",
		then:           "回傳 400，不寫入事件，設定表沒有 STOCK",
		requests:       []settingRequest{{http.MethodPut, "/setting/asset-type/STOCK", settingStockBodyNoFee}},
		wantCode:       http.StatusBadRequest,
		wantAssetTypes: map[string]*assetTypeConfigRow{"STOCK": nil},
	},
	{
		given:    "ACCOUNT_CONFIG 尚無版本紀錄",
		when:     "PUT /setting/asset-type/BOND（不存在的資產類型）",
		then:     "回傳 400，不寫入事件",
		requests: []settingRequest{{http.MethodPut, "/setting/asset-type/BOND", settingStockBody}},
		wantCode: http.StatusBadRequest,
	},
	{
		given:    "SYS_CONFIG 尚無版本紀錄",
		when:     "POST /sys/account 帶期初權益的系統科目對應",
		then:     "回傳 200，寫入版本 1 的事件，sys_accounts 寫入該列",
		requests: []settingRequest{{http.MethodPost, "/sys/account", settingSysAccountBody}},
		wantCode: http.StatusOK,
		wantEvents: []settingEvent{
			{"SYS_CONFIG", "sys_config.sys_account_updated", 1, settingSysAccountBody},
		},
		wantSysAccounts: map[string]*sysAccountRow{
			"SYS:EQUITY:EQUITY_OPENING": {Description: "期初權益", AccountID: "3101-01"},
		},
	},
	{
		given:           "SYS_CONFIG 尚無版本紀錄",
		when:            "POST /sys/account 缺 account_id",
		then:            "回傳 400，不寫入事件，sys_accounts 沒有該列",
		requests:        []settingRequest{{http.MethodPost, "/sys/account", `{"sys_code":"SYS:EQUITY:EQUITY_OPENING","description":"期初權益"}`}},
		wantCode:        http.StatusBadRequest,
		wantSysAccounts: map[string]*sysAccountRow{"SYS:EQUITY:EQUITY_OPENING": nil},
	},
	{
		given: "ACCOUNT_CONFIG 尚無版本紀錄",
		when:  "先 PUT /setting/ledger-account-type/BANK_ACCOUNT，再 PUT /setting/asset-type/STOCK",
		then:  "兩個請求都回傳 200，兩筆事件都屬於 ('ACCOUNT_CONFIG', \"1\")，版本依序為 1、2",
		requests: []settingRequest{
			{http.MethodPut, "/setting/ledger-account-type/BANK_ACCOUNT", `{"account_id":"1101-02"}`},
			{http.MethodPut, "/setting/asset-type/STOCK", settingStockBody},
		},
		wantCode: http.StatusOK,
		wantEvents: []settingEvent{
			{"ACCOUNT_CONFIG", "account_config.ledger_type_updated", 1, `{"type":"BANK_ACCOUNT","account_id":"1101-02"}`},
			{"ACCOUNT_CONFIG", "account_config.asset_type_updated", 2, settingStockPayload},
		},
		wantLedgerTypes: map[string]string{"BANK_ACCOUNT": "1101-02"},
		wantAssetTypes:  map[string]*assetTypeConfigRow{"STOCK": settingStockRow},
	},
}

// newMuxSettingScenarios go through the production NewMux routes. Without a
// route the SPA handler answers 200 with index.html, so the events and read
// models are what prove the route exists.
var newMuxSettingScenarios = []settingScenario{
	{
		given:    "已登入並選擇商戶的使用者，ACCOUNT_CONFIG 尚無版本紀錄",
		when:     "經正式路由 PUT /api/setting/ledger-account-type/BANK_ACCOUNT",
		then:     "回傳 200，寫入版本 1 的事件，帳戶類型設定更新為 1101-02",
		requests: []settingRequest{{http.MethodPut, "/setting/ledger-account-type/BANK_ACCOUNT", `{"account_id":"1101-02"}`}},
		wantCode: http.StatusOK,
		wantEvents: []settingEvent{
			{"ACCOUNT_CONFIG", "account_config.ledger_type_updated", 1, `{"type":"BANK_ACCOUNT","account_id":"1101-02"}`},
		},
		wantLedgerTypes: map[string]string{"BANK_ACCOUNT": "1101-02"},
	},
	{
		given:    "已登入並選擇商戶的使用者，ACCOUNT_CONFIG 尚無版本紀錄",
		when:     "經正式路由 PUT /api/setting/asset-type/STOCK",
		then:     "回傳 200，寫入版本 1 的事件，資產類型設定與送出的值相同",
		requests: []settingRequest{{http.MethodPut, "/setting/asset-type/STOCK", settingStockBody}},
		wantCode: http.StatusOK,
		wantEvents: []settingEvent{
			{"ACCOUNT_CONFIG", "account_config.asset_type_updated", 1, settingStockPayload},
		},
		wantAssetTypes: map[string]*assetTypeConfigRow{"STOCK": settingStockRow},
	},
	{
		given:    "已登入並選擇商戶的使用者，SYS_CONFIG 尚無版本紀錄",
		when:     "經正式路由 POST /api/sys/account",
		then:     "回傳 200，寫入版本 1 的事件，sys_accounts 寫入該列",
		requests: []settingRequest{{http.MethodPost, "/sys/account", settingSysAccountBody}},
		wantCode: http.StatusOK,
		wantEvents: []settingEvent{
			{"SYS_CONFIG", "sys_config.sys_account_updated", 1, settingSysAccountBody},
		},
		wantSysAccounts: map[string]*sysAccountRow{
			"SYS:EQUITY:EQUITY_OPENING": {Description: "期初權益", AccountID: "3101-01"},
		},
	},
}
