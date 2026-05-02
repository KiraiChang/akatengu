package db

//dbmap:sqlcdb=SysAccount
//dbmap:sqlcdb=UpdateSysAccountParams
type SysAccount struct {
	SysCode   string `db:"sys_code" json:"sys_code"`
	AccountId string `db:"account_id" json:"account_id"`
}
