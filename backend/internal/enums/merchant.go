package enums

import "akatengu/internal/pkg/enumx"

type MerchantStatusType = enumx.Enum[merchantStatusTypeVal]

//enumx:enum
type merchantStatusTypeVal string

const (
	MerchantStatusTypeActive   merchantStatusTypeVal = "ACTIVE"
	MerchantStatusTypeInactive merchantStatusTypeVal = "INACTIVE"
)

type MerchantRoleType = enumx.Enum[merchantRoleTypeVal]

//enumx:enum
type merchantRoleTypeVal string

const (
	MerchantRoleTypeOwner  merchantRoleTypeVal = "OWNER"
	MerchantRoleTypeAdmin  merchantRoleTypeVal = "ADMIN"
	MerchantRoleTypeMember merchantRoleTypeVal = "MEMBER"
)
