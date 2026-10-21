package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db"
	"akatengu/internal/model/db/projection"
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

// ErrTxBoundWrite is returned by the write methods of a Repo bound to a
// transaction; such a Repo is only for reads inside the transaction.
var ErrTxBoundWrite = errors.New("query: write through a transaction-bound repo is not allowed")

// NewQueryRepositoryTx returns a Repo whose queries run in tx, so code that
// holds the only connection (SQLite, MaxOpenConns=1) can still query inside its
// transaction. The Repo is valid only while tx is open; the caller owns that
// lifetime (BFSEngine.Run from R1 of the BFS migration plan).
//
// The Repo is read-only. These write methods return ErrTxBoundWrite without
// writing or opening a nested transaction:
//
//	Config:   UpsertLedgerAccountTypeConfig, UpsertAssetTypeAccountConfig
//	Merchant: Create, AddUser, Update, Deactivate
//	User:     Create
//	Template: CreateTemplate, UpdateTemplate, DeleteTemplate
//
// The wrappers below embed the original repos and override only these
// methods. A write method added to one of these repos must be overridden here
// too; otherwise the bound Repo lets the write through.
func NewQueryRepositoryTx(tx *sqlx.Tx) *Repo {
	r := newQueryRepository(sqlcdb.New(tx), tx, nil)
	r.Config = txConfigRepo{r.Config}
	r.Merchant = txMerchantRepo{r.Merchant}
	r.User = txUserRepo{r.User}
	r.Template = txTemplateRepo{r.Template}
	return r
}

type txConfigRepo struct{ ConfigRepo }

func (txConfigRepo) UpsertLedgerAccountTypeConfig(context.Context, enums.LedgerAccountType, string, *string) error {
	return ErrTxBoundWrite
}

func (txConfigRepo) UpsertAssetTypeAccountConfig(context.Context, sqlcdb.UpsertAssetTypeAccountConfigParams) error {
	return ErrTxBoundWrite
}

type txMerchantRepo struct{ MerchantRepo }

func (txMerchantRepo) Create(context.Context, string, string, string) (int64, error) {
	return 0, ErrTxBoundWrite
}

func (txMerchantRepo) AddUser(context.Context, int64, int64, enums.MerchantRoleType) error {
	return ErrTxBoundWrite
}

func (txMerchantRepo) Update(context.Context, int64, string, string, string) error {
	return ErrTxBoundWrite
}

func (txMerchantRepo) Deactivate(context.Context, int64) error {
	return ErrTxBoundWrite
}

type txUserRepo struct{ UserRepo }

func (txUserRepo) Create(context.Context, *db.User) (*db.User, error) {
	return nil, ErrTxBoundWrite
}

type txTemplateRepo struct{ TemplateRepo }

func (txTemplateRepo) CreateTemplate(context.Context, CreateTemplateCmd) (*projection.TransactionTemplate, error) {
	return nil, ErrTxBoundWrite
}

func (txTemplateRepo) UpdateTemplate(context.Context, int64, UpdateTemplateCmd) error {
	return ErrTxBoundWrite
}

func (txTemplateRepo) DeleteTemplate(context.Context, int64) error {
	return ErrTxBoundWrite
}
