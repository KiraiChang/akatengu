package db

import (
	"akatengu/internal/enums"
)

//dbmap:sqlcdb=User
type User struct {
	UserId   int64                `db:"user_id"`
	Username string               `db:"username"`
	Password string               `db:"password"`
	Status   enums.UserStatusType `db:"status"`
}
