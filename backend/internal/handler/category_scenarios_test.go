package handler

// categoryAPIScenario creates, updates and deletes one category through the
// production NewMux routes. The accounts come from seeds/accounts.sql, the same
// as services/category_test.go.
type categoryAPIScenario struct {
	given         string
	when          string
	then          string
	path          string // the create route without /api; update and delete add /{category_id}
	table         string // the read model
	aggregateType string
	createBody    string
	updateBody    string // sends expected_version 1, the read-model version after the creation
	updatedName   string
	eventPrefix   string // the event types are <eventPrefix>.created, .updated and .deleted
}

// categoryDeleteBody sends the read-model version after one update.
const categoryDeleteBody = `{"expected_version":2}`

var categoryAPIScenarios = []categoryAPIScenario{
	{
		given:         "已登入並選擇商戶的使用者",
		when:          "經正式路由 POST、PUT、DELETE /api/prepaid/category，修改與刪除送出畫面讀到的版本",
		then:          "都回傳 200；3 筆事件都寫在 ('PREPAID_CATEGORY', category_uuid)，版本 1～3，category_uuid 為建立事件的 UUID；類別停用，讀模型版本 3",
		path:          "/prepaid/category",
		table:         "prepaid_categories",
		aggregateType: "PREPAID_CATEGORY",
		createBody:    `{"name":"保險費","account_id":"1104-01","expense_account_id":"5201-01"}`,
		updateBody:    `{"expected_version":1,"name":"保險費（修正）","account_id":"1104-01","expense_account_id":"5201-01"}`,
		updatedName:   "保險費（修正）",
		eventPrefix:   "prepaid_category",
	},
	{
		given:         "已登入並選擇商戶的使用者",
		when:          "經正式路由 POST、PUT、DELETE /api/fixed_asset/category，修改與刪除送出畫面讀到的版本",
		then:          "都回傳 200；3 筆事件都寫在 ('FIXED_ASSET_CATEGORY', category_uuid)，版本 1～3，category_uuid 為建立事件的 UUID；類別停用，讀模型版本 3",
		path:          "/fixed_asset/category",
		table:         "fixed_asset_categories",
		aggregateType: "FIXED_ASSET_CATEGORY",
		createBody: `{"name":"電腦設備","asset_account_id":"1601-01","accum_depreciation_account_id":"1601-02",` +
			`"depreciation_expense_account_id":"5601-01"}`,
		updateBody: `{"expected_version":1,"name":"電腦設備（修正）","asset_account_id":"1601-01",` +
			`"accum_depreciation_account_id":"1601-02","depreciation_expense_account_id":"5601-01"}`,
		updatedName: "電腦設備（修正）",
		eventPrefix: "fixed_asset_category",
	},
}
