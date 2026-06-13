package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
	"context"
	"database/sql"
)

type BankPdfTemplateRepo interface {
	GetBankPdfTemplates(ctx context.Context) ([]projection.BankPdfTemplate, error)
	GetBankPdfTemplateByID(ctx context.Context, id int64) (*projection.BankPdfTemplate, error)
	GetBankPdfTemplateByUUID(ctx context.Context, uuid string) (*projection.BankPdfTemplate, error)
	GetBankPdfTemplateLedgers(ctx context.Context, templateID int64) ([]projection.BankPdfTemplateLedger, error)
}

type sqlcdbBankPdfTemplateRepo struct {
	q *sqlcdb.Queries
}

func newBankPdfTemplateRepo(q *sqlcdb.Queries) BankPdfTemplateRepo {
	return &sqlcdbBankPdfTemplateRepo{q: q}
}

func (r *sqlcdbBankPdfTemplateRepo) GetBankPdfTemplates(ctx context.Context) ([]projection.BankPdfTemplate, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetBankPdfTemplates(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.BankPdfTemplate, len(rows))
	for i, row := range rows {
		result[i] = projection.BankPdfTemplateFromBankPdfTemplate(row)
	}
	return result, nil
}

func (r *sqlcdbBankPdfTemplateRepo) GetBankPdfTemplateByID(ctx context.Context, id int64) (*projection.BankPdfTemplate, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetBankPdfTemplateByID(ctx, sqlcdb.GetBankPdfTemplateByIDParams{
		TemplateID: id,
		MerchantID: merchantID,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.BankPdfTemplatePtrFromBankPdfTemplate(row), nil
}

func (r *sqlcdbBankPdfTemplateRepo) GetBankPdfTemplateByUUID(ctx context.Context, uuid string) (*projection.BankPdfTemplate, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetBankPdfTemplateByUUID(ctx, sqlcdb.GetBankPdfTemplateByUUIDParams{
		TemplateUuid: uuid,
		MerchantID:   merchantID,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.BankPdfTemplatePtrFromBankPdfTemplate(row), nil
}

func (r *sqlcdbBankPdfTemplateRepo) GetBankPdfTemplateLedgers(ctx context.Context, templateID int64) ([]projection.BankPdfTemplateLedger, error) {
	rows, err := r.q.GetBankPdfTemplateLedgers(ctx, templateID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.BankPdfTemplateLedger, len(rows))
	for i, row := range rows {
		result[i] = projection.BankPdfTemplateLedgerFromBankPdfTemplateLedger(row)
	}
	return result, nil
}
