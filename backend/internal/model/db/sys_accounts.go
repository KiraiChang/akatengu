package db

//dbmap:sqlcdb=SysAccount
type SysAccount struct {
	SysCode   string `db:"sys_code"`
	AccountId string `db:"account_id"`
}
