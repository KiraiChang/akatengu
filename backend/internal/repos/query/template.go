package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type TemplateRepo interface {
	GetTemplates(ctx context.Context, search string) ([]projection.TransactionTemplate, error)
	GetTemplate(ctx context.Context, id int64) (*projection.TransactionTemplateDetail, error)
	CreateTemplate(ctx context.Context, cmd CreateTemplateCmd) (*projection.TransactionTemplate, error)
	UpdateTemplate(ctx context.Context, id int64, cmd UpdateTemplateCmd) error
	DeleteTemplate(ctx context.Context, id int64) error
}

type TemplateEntryInput struct {
	SortOrder        int64
	AccountID        string
	LedgerID         *int64
	Debit            decimal.Decimal
	Credit           decimal.Decimal
	Note             *string
	CashFlowCategory enums.CashFlowCategory
}

type CreateTemplateCmd struct {
	Name        string
	Description *string
	Tag         *string
	UpdatedBy   *string
	Entries     []TemplateEntryInput
}

type UpdateTemplateCmd struct {
	Name        string
	Description *string
	Tag         *string
	UpdatedBy   *string
	Entries     []TemplateEntryInput
}

type sqlxTemplateRepo struct {
	q  *sqlcdb.Queries
	db *sqlx.DB
}

func newTemplateRepo(q *sqlcdb.Queries, db *sqlx.DB) TemplateRepo {
	return &sqlxTemplateRepo{q: q, db: db}
}

func NewTemplateRepo(db *sqlx.DB) TemplateRepo {
	return &sqlxTemplateRepo{q: sqlcdb.New(db), db: db}
}

func (r *sqlxTemplateRepo) GetTemplates(ctx context.Context, search string) ([]projection.TransactionTemplate, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetTemplates(ctx, sqlcdb.GetTemplatesParams{
		MerchantID: merchantID,
		Search:     search,
	})
	if err != nil {
		return nil, err
	}
	result := make([]projection.TransactionTemplate, len(rows))
	for i, row := range rows {
		result[i] = projection.TransactionTemplateFromTransactionTemplate(row)
	}
	return result, nil
}

func (r *sqlxTemplateRepo) GetTemplate(ctx context.Context, id int64) (*projection.TransactionTemplateDetail, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetTemplateByID(ctx, sqlcdb.GetTemplateByIDParams{
		ID:         id,
		MerchantID: merchantID,
	})
	if err != nil {
		return nil, err
	}
	entryRows, err := r.q.GetTemplateEntriesByTemplateID(ctx, sqlcdb.GetTemplateEntriesByTemplateIDParams{
		TemplateID: id,
		MerchantID: merchantID,
	})
	if err != nil {
		return nil, err
	}
	entries := make([]projection.TransactionTemplateEntry, len(entryRows))
	for i, e := range entryRows {
		entries[i] = projection.TransactionTemplateEntryFromTransactionTemplateEntry(e)
	}
	return &projection.TransactionTemplateDetail{
		TransactionTemplate: projection.TransactionTemplateFromTransactionTemplate(row),
		Entries:             entries,
	}, nil
}

func (r *sqlxTemplateRepo) CreateTemplate(ctx context.Context, cmd CreateTemplateCmd) (*projection.TransactionTemplate, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	tx, err := r.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	qtx := r.q.WithTx(tx)

	tmpl, err := qtx.InsertTemplate(ctx, sqlcdb.InsertTemplateParams{
		MerchantID:  merchantID,
		Name:        cmd.Name,
		Description: cmd.Description,
		Tag:         cmd.Tag,
		UpdatedBy:   cmd.UpdatedBy,
	})
	if err != nil {
		return nil, err
	}
	if err = insertTemplateEntries(ctx, qtx, merchantID, tmpl.ID, cmd.Entries); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	result := projection.TransactionTemplateFromTransactionTemplate(tmpl)
	return &result, nil
}

func (r *sqlxTemplateRepo) UpdateTemplate(ctx context.Context, id int64, cmd UpdateTemplateCmd) error {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return err
	}
	tx, err := r.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	qtx := r.q.WithTx(tx)

	if err = qtx.UpdateTemplate(ctx, sqlcdb.UpdateTemplateParams{
		ID:          id,
		MerchantID:  merchantID,
		Name:        cmd.Name,
		Description: cmd.Description,
		Tag:         cmd.Tag,
		UpdatedBy:   cmd.UpdatedBy,
	}); err != nil {
		return err
	}
	if err = qtx.DeleteTemplateEntriesByTemplateID(ctx, sqlcdb.DeleteTemplateEntriesByTemplateIDParams{
		TemplateID: id,
		MerchantID: merchantID,
	}); err != nil {
		return err
	}
	if err = insertTemplateEntries(ctx, qtx, merchantID, id, cmd.Entries); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *sqlxTemplateRepo) DeleteTemplate(ctx context.Context, id int64) error {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return err
	}
	return r.q.DeleteTemplate(ctx, sqlcdb.DeleteTemplateParams{
		ID:         id,
		MerchantID: merchantID,
	})
}

func insertTemplateEntries(ctx context.Context, q *sqlcdb.Queries, merchantID, templateID int64, entries []TemplateEntryInput) error {
	for _, e := range entries {
		if err := q.InsertTemplateEntry(ctx, sqlcdb.InsertTemplateEntryParams{
			MerchantID:       merchantID,
			TemplateID:       templateID,
			SortOrder:        e.SortOrder,
			AccountID:        e.AccountID,
			LedgerID:         e.LedgerID,
			Debit:            e.Debit,
			Credit:           e.Credit,
			Note:             e.Note,
			CashFlowCategory: e.CashFlowCategory,
		}); err != nil {
			return err
		}
	}
	return nil
}
