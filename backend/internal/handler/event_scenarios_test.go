package handler

import (
	"net/http"
)

// eventAppendNewAccountID is sent with account.created; not in seeds/accounts.sql.
const eventAppendNewAccountID = "1101-92"

type eventAppendScenario struct {
	given         string
	when          string
	then          string
	body          string
	wantCode      int
	wantDetail    string // substring of the problem detail ("" = success response)
	wantEventType string // event type counted in event_store after the request
	wantEvents    int    // events of wantEventType stored after the request
	wantCategory  bool   // entry_cf_categories has a row for cfTestDebitEntryUUID
	wantBFS       bool   // the stored events carry BFS execution metadata (source api)
}

// eventAppendScenarios go through the production NewMux wiring.
var eventAppendScenarios = []eventAppendScenario{
	{
		given: "交易與分錄已存在，使用者已選擇商戶",
		when:  "POST /api/event/append 送出已遷移到 BFS 的 transaction.cf_category_updated",
		then:  "BFS 派送拒絕（沒有舊映射），回傳 500 與 problem 格式；不寫入事件，分錄沒有分類",
		body: `{"aggregate_type":"TRANSACTION","aggregate_id":"` + cfTestTxnUUID + `","expected_version":0,` +
			`"event_type":"transaction.cf_category_updated",` +
			`"payload":{"txn_uuid":"` + cfTestTxnUUID + `","entries":[{"entry_uuid":"` + cfTestDebitEntryUUID + `","cf_category":"INVESTING"}]}}`,
		wantCode:      http.StatusInternalServerError,
		wantDetail:    "is not accepted by POST /api/event/append",
		wantEventType: "transaction.cf_category_updated",
		wantEvents:    0,
	},
	{
		given: "交易與分錄已存在，使用者已選擇商戶",
		when:  "POST /api/event/append 送出 R2 段 2 遷移到 BFS 的 account.created",
		then:  "BFS 派送成功，回傳 200，寫入 1 筆 account.created 事件",
		body: `{"aggregate_type":"ACCOUNT","aggregate_id":"` + eventAppendNewAccountID + `","expected_version":0,` +
			`"event_type":"account.created",` +
			`"payload":{"account_id":"` + eventAppendNewAccountID + `","parent_id":"1101","name":"F7 對照科目",` +
			`"type":"ASSET","normal_balance":"DEBIT","currency":"TWD","is_active":true}}`,
		wantCode:      http.StatusOK,
		wantEventType: "account.created",
		wantEvents:    1,
		wantBFS:       true,
	},
	{
		given: "交易與分錄已存在，使用者已選擇商戶",
		when:  "POST /api/event/append 送出 account.created，aggregate ID 與 payload.account_id 不同",
		then:  "BFS 派送拒絕（不符合舊映射），回傳 500 與 problem 格式；不寫入事件",
		body: `{"aggregate_type":"ACCOUNT","aggregate_id":"1101","expected_version":0,` +
			`"event_type":"account.created",` +
			`"payload":{"account_id":"` + eventAppendNewAccountID + `","parent_id":"1101","name":"F7 對照科目",` +
			`"type":"ASSET","normal_balance":"DEBIT","currency":"TWD","is_active":true}}`,
		wantCode:      http.StatusInternalServerError,
		wantDetail:    "aggregate does not match the legacy aggregate",
		wantEventType: "account.created",
		wantEvents:    0,
	},
	{
		given: "交易與分錄已存在，使用者已選擇商戶",
		when:  "POST /api/event/append 送出未遷移的 investment.created",
		then:  "交給舊路徑，回傳 200，寫入 1 筆 investment.created 事件",
		body: `{"aggregate_type":"TRANSACTION","aggregate_id":"","expected_version":0,` +
			`"event_type":"investment.created",` +
			`"payload":{"account_id":"1203-02","asset_type":"STOCK","currency":"TWD","symbol":"R2-CONTROL",` +
			`"name":"R2 對照投資","cost_method":"FIFO","ifrs_category":"FVOCI","is_active":true}}`,
		wantCode:      http.StatusOK,
		wantEventType: "investment.created",
		wantEvents:    1,
	},
}
