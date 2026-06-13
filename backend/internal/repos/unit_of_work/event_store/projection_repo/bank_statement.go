package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"context"
)

type sqlcdbBankCsvTemplateRepo struct {
	q *sqlcdb.Queries
}

func NewBankCsvTemplateRepo(q *sqlcdb.Queries) BankCsvTemplateRepo {
	return &sqlcdbBankCsvTemplateRepo{q: q}
}

func (r *sqlcdbBankCsvTemplateRepo) InsertBankCsvTemplate(ctx context.Context, p sqlcdb.InsertBankCsvTemplateParams) error {
	_, err := r.q.InsertBankCsvTemplate(ctx, p)
	return err
}

func (r *sqlcdbBankCsvTemplateRepo) UpdateBankCsvTemplate(ctx context.Context, p sqlcdb.UpdateBankCsvTemplateParams) error {
	return r.q.UpdateBankCsvTemplate(ctx, p)
}

func (r *sqlcdbBankCsvTemplateRepo) DeactivateBankCsvTemplate(ctx context.Context, p sqlcdb.DeactivateBankCsvTemplateParams) error {
	return r.q.DeactivateBankCsvTemplate(ctx, p)
}

func (r *sqlcdbBankCsvTemplateRepo) InsertBankCsvTemplateLedger(ctx context.Context, p sqlcdb.InsertBankCsvTemplateLedgerParams) error {
	return r.q.InsertBankCsvTemplateLedger(ctx, p)
}

func (r *sqlcdbBankCsvTemplateRepo) DeleteBankCsvTemplateLedgers(ctx context.Context, templateID int64) error {
	return r.q.DeleteBankCsvTemplateLedgers(ctx, templateID)
}

func (r *sqlcdbBankCsvTemplateRepo) GetIDByUUID(ctx context.Context, uuid string, merchantID int64) (int64, error) {
	return r.q.GetBankCsvTemplateIDByUUID(ctx, sqlcdb.GetBankCsvTemplateIDByUUIDParams{
		TemplateUuid: uuid,
		MerchantID:   merchantID,
	})
}

// ─────────────────────────────────────────────────────────

type sqlcdbBankStatementImportRepo struct {
	q *sqlcdb.Queries
}

func NewBankStatementImportRepo(q *sqlcdb.Queries) BankStatementImportRepo {
	return &sqlcdbBankStatementImportRepo{q: q}
}

func (r *sqlcdbBankStatementImportRepo) InsertBankStatementImport(ctx context.Context, p sqlcdb.InsertBankStatementImportParams) (int64, error) {
	row, err := r.q.InsertBankStatementImport(ctx, p)
	if err != nil {
		return 0, err
	}
	return row.ImportID, nil
}

func (r *sqlcdbBankStatementImportRepo) UpdateBankStatementImportStatus(ctx context.Context, p sqlcdb.UpdateBankStatementImportStatusParams) error {
	return r.q.UpdateBankStatementImportStatus(ctx, p)
}

func (r *sqlcdbBankStatementImportRepo) InsertBankStatementTxn(ctx context.Context, p sqlcdb.InsertBankStatementTxnParams) error {
	_, err := r.q.InsertBankStatementTxn(ctx, p)
	return err
}

func (r *sqlcdbBankStatementImportRepo) UpdateBankTxnMatch(ctx context.Context, p sqlcdb.UpdateBankTxnMatchParams) error {
	return r.q.UpdateBankTxnMatch(ctx, p)
}

func (r *sqlcdbBankStatementImportRepo) UpdateBankTxnStatus(ctx context.Context, p sqlcdb.UpdateBankTxnStatusParams) error {
	return r.q.UpdateBankTxnStatus(ctx, p)
}

func (r *sqlcdbBankStatementImportRepo) UpdateBankTxnCreatedTxn(ctx context.Context, p sqlcdb.UpdateBankTxnCreatedTxnParams) error {
	return r.q.UpdateBankTxnCreatedTxn(ctx, p)
}

func (r *sqlcdbBankStatementImportRepo) ResetNonConfirmedMatches(ctx context.Context, p sqlcdb.ResetNonConfirmedMatchesParams) error {
	return r.q.ResetNonConfirmedMatches(ctx, p)
}

func (r *sqlcdbBankStatementImportRepo) InsertBankStatementImportLedger(ctx context.Context, p sqlcdb.InsertBankStatementImportLedgerParams) error {
	return r.q.InsertBankStatementImportLedger(ctx, p)
}
