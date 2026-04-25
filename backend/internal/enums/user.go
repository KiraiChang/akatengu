package enums

import "akatengu/internal/pkg/enumx"

type UserStatusType = enumx.Enum[userStatusTypeVal]

//enumx:enum
type userStatusTypeVal string

const (
	UserStatusTypeActive   userStatusTypeVal = "ACTIVE"
	UserStatusTypeInactive userStatusTypeVal = "INACTIVE"
)
