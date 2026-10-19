package handler

import (
	"net/http"
)

// Fixture identifiers inserted by insertCFFixture in entry_cf_category_test.go.
const (
	cfTestTxnUUID         = "0192f3a0-0000-7000-8000-000000000001"
	cfTestDebitEntryUUID  = "0192f3a0-0000-7000-8000-000000000011"
	cfTestCreditEntryUUID = "0192f3a0-0000-7000-8000-000000000012"
	cfOtherTxnUUID        = "0192f3a0-0000-7000-8000-000000000002"
	cfOtherEntryUUID      = "0192f3a0-0000-7000-8000-000000000021"
	cfMissingTxnUUID      = "0192f3a0-0000-7000-8000-000000000099"
)

type cfCategoryScenario struct {
	given           string
	when            string
	then            string
	txnUUID         string // path parameter {txn_uuid}
	existingVersion int64  // aggregate version stored before the request (0 = none)
	body            string
	wantCode        int
	wantEventCount  int    // events stored for the transaction aggregate after the request
	wantVersion     int64  // aggregate version of the stored event (0 = not checked)
	wantCategory    string // cf_category of cfTestDebitEntryUUID after the request ("" = no row)
}

var updateCFCategoryScenarios = []cfCategoryScenario{
	{
		given:          "交易尚無現金流分類修改紀錄",
		when:           "PUT /txn/{txn_uuid}/cf-category 帶 expected_version 0 與該交易的分錄",
		then:           "回傳 200，寫入版本 1 的事件，分錄分類更新為已確認的 INVESTING",
		txnUUID:        cfTestTxnUUID,
		body:           `{"expected_version":0,"entries":[{"entry_uuid":"` + cfTestDebitEntryUUID + `","cf_category":"INVESTING"}]}`,
		wantCode:       http.StatusOK,
		wantEventCount: 1,
		wantVersion:    1,
		wantCategory:   "INVESTING",
	},
	{
		given:           "交易的 aggregate 已在版本 1",
		when:            "PUT /txn/{txn_uuid}/cf-category 帶 expected_version 1",
		then:            "回傳 200，寫入版本 2 的事件",
		txnUUID:         cfTestTxnUUID,
		existingVersion: 1,
		body:            `{"expected_version":1,"entries":[{"entry_uuid":"` + cfTestDebitEntryUUID + `","cf_category":"FINANCING"}]}`,
		wantCode:        http.StatusOK,
		wantEventCount:  1,
		wantVersion:     2,
		wantCategory:    "FINANCING",
	},
	{
		given:           "交易的 aggregate 已在版本 1",
		when:            "PUT /txn/{txn_uuid}/cf-category 帶過期的 expected_version 0",
		then:            "回傳 400，不寫入事件，分錄分類不變",
		txnUUID:         cfTestTxnUUID,
		existingVersion: 1,
		body:            `{"expected_version":0,"entries":[{"entry_uuid":"` + cfTestDebitEntryUUID + `","cf_category":"INVESTING"}]}`,
		wantCode:        http.StatusBadRequest,
	},
	{
		given:    "交易尚無現金流分類修改紀錄",
		when:     "PUT /txn/{txn_uuid}/cf-category 帶 expected_version 7",
		then:     "回傳 400，不寫入事件，分錄分類不變",
		txnUUID:  cfTestTxnUUID,
		body:     `{"expected_version":7,"entries":[{"entry_uuid":"` + cfTestDebitEntryUUID + `","cf_category":"INVESTING"}]}`,
		wantCode: http.StatusBadRequest,
	},
	{
		given:    "交易尚無現金流分類修改紀錄",
		when:     "PUT /txn/{txn_uuid}/cf-category 帶 expected_version -1",
		then:     "回傳 400，不寫入事件，分錄分類不變",
		txnUUID:  cfTestTxnUUID,
		body:     `{"expected_version":-1,"entries":[{"entry_uuid":"` + cfTestDebitEntryUUID + `","cf_category":"INVESTING"}]}`,
		wantCode: http.StatusBadRequest,
	},
	{
		given:    "請求中的分錄屬於另一筆交易",
		when:     "PUT /txn/{txn_uuid}/cf-category",
		then:     "回傳 400，不寫入事件",
		txnUUID:  cfTestTxnUUID,
		body:     `{"expected_version":0,"entries":[{"entry_uuid":"` + cfOtherEntryUUID + `","cf_category":"INVESTING"}]}`,
		wantCode: http.StatusBadRequest,
	},
	{
		given:    "路徑中的交易不存在",
		when:     "PUT /txn/{txn_uuid}/cf-category",
		then:     "回傳 400，不寫入事件",
		txnUUID:  cfMissingTxnUUID,
		body:     `{"expected_version":0,"entries":[{"entry_uuid":"` + cfTestDebitEntryUUID + `","cf_category":"INVESTING"}]}`,
		wantCode: http.StatusBadRequest,
	},
}

// newMuxCFCategoryScenarios go through the production NewMux wiring, so a missing
// projector registration there is caught.
var newMuxCFCategoryScenarios = []cfCategoryScenario{
	{
		given:          "已登入並選擇商戶的使用者，交易尚無現金流分類修改紀錄",
		when:           "經正式路由 PUT /api/txn/{txn_uuid}/cf-category 帶 expected_version 0",
		then:           "回傳 200，寫入版本 1 的事件，分錄分類更新為已確認的 OPERATING",
		txnUUID:        cfTestTxnUUID,
		body:           `{"expected_version":0,"entries":[{"entry_uuid":"` + cfTestDebitEntryUUID + `","cf_category":"OPERATING"}]}`,
		wantCode:       http.StatusOK,
		wantEventCount: 1,
		wantVersion:    1,
		wantCategory:   "OPERATING",
	},
}
