package request

import "errors"

type CreateMerchant struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Currency    string `json:"currency"`
}

func (r *CreateMerchant) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if r.DisplayName == "" {
		return errors.New("display_name is required")
	}
	if r.Currency == "" {
		r.Currency = "TWD"
	}
	return nil
}

type SelectMerchant struct {
	MerchantID int64 `json:"merchant_id"`
}

func (r *SelectMerchant) Validate() error {
	if r.MerchantID <= 0 {
		return errors.New("merchant_id is required")
	}
	return nil
}
