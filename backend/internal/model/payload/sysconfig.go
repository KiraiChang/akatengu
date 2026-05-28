package payload

// SysAccountUpdatedPayload 更新系統設定科目映射。
// Description 為識別標籤，由事件本身儲存以確保 replay 時可完整重建（含初始描述）。
type SysAccountUpdatedPayload struct {
	SysCode     string `json:"sys_code"`
	Description string `json:"description"`
	AccountID   string `json:"account_id"`
}

func (p SysAccountUpdatedPayload) Validate() error {
	var errs []string
	if p.SysCode == "" {
		errs = append(errs, "sys_code is required")
	}
	if p.AccountID == "" {
		errs = append(errs, "account_id is required")
	}
	return joinErrors(errs)
}
