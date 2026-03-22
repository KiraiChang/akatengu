package projection

import (
	"github.com/jmoiron/sqlx"
)

type sqlxTxProjectionRepository struct{ tx *sqlx.Tx }

func NewSqlxTxProjectionRepository(tx *sqlx.Tx) *sqlxTxProjectionRepository {
	return &sqlxTxProjectionRepository{tx: tx}
}
