package request

import "errors"

type SysAccount struct {
	SysCode   string `json:"sys_code"`
	AccountId string `json:"account_id"`
}

func (s *SysAccount) Validate() error {
	if s.SysCode == "" {
		return errors.New("sys_code is required")
	}
	if s.AccountId == "" {
		return errors.New("account_id is required")
	}
	return nil
}
