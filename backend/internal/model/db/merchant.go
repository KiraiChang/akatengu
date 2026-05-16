package db

import "akatengu/internal/enums"

//dbmap:sqlcdb=Merchant
//dbmap:sqlcdb=GetUserMerchantsRow
type Merchant struct {
	MerchantID  int64                    `db:"merchant_id" json:"merchant_id"`
	Name        string                   `db:"name" json:"name"`
	DisplayName string                   `db:"display_name" json:"display_name"`
	Currency    string                   `db:"currency" json:"currency"`
	Status      enums.MerchantStatusType `db:"status" json:"status"`
	CreatedAt   string                   `db:"created_at" json:"created_at"`
}
