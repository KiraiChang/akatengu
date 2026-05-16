package db

//dbmap:sqlcdb=SysAccount
//dbmap:sqlcdb=GetSysAccountsRow
//dbmap:sqlcdb=UpdateSysAccountParams
type SysAccount struct {
	SysCode     string `db:"sys_code" json:"sys_code"`
	Description string `db:"description" json:"description"`
	AccountId   string `db:"account_id" json:"account_id"`
}
