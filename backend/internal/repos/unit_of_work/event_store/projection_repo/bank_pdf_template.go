package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"context"
)

type sqlcdbBankPdfTemplateRepo struct {
	q *sqlcdb.Queries
}

func NewBankPdfTemplateRepo(q *sqlcdb.Queries) BankPdfTemplateRepo {
	return &sqlcdbBankPdfTemplateRepo{q: q}
}

func (r *sqlcdbBankPdfTemplateRepo) InsertBankPdfTemplate(ctx context.Context, p sqlcdb.InsertBankPdfTemplateParams) (int64, error) {
	row, err := r.q.InsertBankPdfTemplate(ctx, p)
	if err != nil {
		return 0, err
	}
	return row.TemplateID, nil
}

func (r *sqlcdbBankPdfTemplateRepo) UpdateBankPdfTemplate(ctx context.Context, p sqlcdb.UpdateBankPdfTemplateParams) error {
	return r.q.UpdateBankPdfTemplate(ctx, p)
}

func (r *sqlcdbBankPdfTemplateRepo) DeactivateBankPdfTemplate(ctx context.Context, p sqlcdb.DeactivateBankPdfTemplateParams) error {
	return r.q.DeactivateBankPdfTemplate(ctx, p)
}

func (r *sqlcdbBankPdfTemplateRepo) InsertBankPdfTemplateLedger(ctx context.Context, p sqlcdb.InsertBankPdfTemplateLedgerParams) error {
	return r.q.InsertBankPdfTemplateLedger(ctx, p)
}

func (r *sqlcdbBankPdfTemplateRepo) DeleteBankPdfTemplateLedgers(ctx context.Context, templateID int64) error {
	return r.q.DeleteBankPdfTemplateLedgers(ctx, templateID)
}

func (r *sqlcdbBankPdfTemplateRepo) GetIDByUUID(ctx context.Context, uuid string, merchantID int64) (int64, error) {
	return r.q.GetBankPdfTemplateIDByUUID(ctx, sqlcdb.GetBankPdfTemplateIDByUUIDParams{
		TemplateUuid: uuid,
		MerchantID:   merchantID,
	})
}
