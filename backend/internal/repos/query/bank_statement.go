package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
	"context"
	"database/sql"
)

// ─────────────────────────────────────────
// BankCsvTemplate
// ─────────────────────────────────────────

type BankCsvTemplateRepo interface {
	GetBankCsvTemplates(ctx context.Context) ([]projection.BankCsvTemplate, error)
	GetBankCsvTemplateByID(ctx context.Context, id int64) (*projection.BankCsvTemplate, error)
	GetBankCsvTemplateByUUID(ctx context.Context, uuid string) (*projection.BankCsvTemplate, error)
	GetBankCsvTemplateLedgers(ctx context.Context, templateID int64) ([]projection.BankCsvTemplateLedger, error)
}

type sqlcdbBankCsvTemplateRepo struct {
	q *sqlcdb.Queries
}

func newBankCsvTemplateRepo(q *sqlcdb.Queries) BankCsvTemplateRepo {
	return &sqlcdbBankCsvTemplateRepo{q: q}
}

func (r *sqlcdbBankCsvTemplateRepo) GetBankCsvTemplates(ctx context.Context) ([]projection.BankCsvTemplate, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetBankCsvTemplates(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.BankCsvTemplate, len(rows))
	for i, row := range rows {
		result[i] = projection.BankCsvTemplateFromBankCsvTemplate(row)
	}
	return result, nil
}

func (r *sqlcdbBankCsvTemplateRepo) GetBankCsvTemplateByID(ctx context.Context, id int64) (*projection.BankCsvTemplate, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetBankCsvTemplateByID(ctx, sqlcdb.GetBankCsvTemplateByIDParams{
		TemplateID: id,
		MerchantID: merchantID,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.BankCsvTemplatePtrFromBankCsvTemplate(row), nil
}

func (r *sqlcdbBankCsvTemplateRepo) GetBankCsvTemplateByUUID(ctx context.Context, uuid string) (*projection.BankCsvTemplate, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetBankCsvTemplateByUUID(ctx, sqlcdb.GetBankCsvTemplateByUUIDParams{
		TemplateUuid: uuid,
		MerchantID:   merchantID,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.BankCsvTemplatePtrFromBankCsvTemplate(row), nil
}

func (r *sqlcdbBankCsvTemplateRepo) GetBankCsvTemplateLedgers(ctx context.Context, templateID int64) ([]projection.BankCsvTemplateLedger, error) {
	rows, err := r.q.GetBankCsvTemplateLedgers(ctx, templateID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.BankCsvTemplateLedger, len(rows))
	for i, row := range rows {
		result[i] = projection.BankCsvTemplateLedgerFromBankCsvTemplateLedger(row)
	}
	return result, nil
}

// ─────────────────────────────────────────
// BankStatementImport
// ─────────────────────────────────────────

type BankStatementImportRepo interface {
	GetBankStatementImportByID(ctx context.Context, id int64) (*projection.BankStatementImport, error)
	GetBankStatementImportByUUID(ctx context.Context, uuid string) (*projection.BankStatementImport, error)
	GetBankStatementImportsPaged(ctx context.Context, ledgerID int64, pageSize, offset int64) ([]projection.BankStatementImport, int64, error)
	GetBankStatementTxnsByImport(ctx context.Context, importID int64) ([]projection.BankStatementTxn, error)
	GetBankStatementTxnByID(ctx context.Context, id int64) (*projection.BankStatementTxn, error)
	GetUnmatchedBankTxns(ctx context.Context, importID int64) ([]projection.BankStatementTxn, error)
	GetBankTxnsForReview(ctx context.Context, importID int64) ([]projection.BankStatementTxn, error)
	CountBankTxnsByStatus(ctx context.Context, importID int64, status string) (int64, error)
	GetLedgerEntriesForMatching(ctx context.Context, ledgerID int64, dateFrom, dateTo string) ([]sqlcdb.GetLedgerEntriesForMatchingRow, error)
}

type sqlcdbBankStatementImportRepo struct {
	q *sqlcdb.Queries
}

func newBankStatementImportRepo(q *sqlcdb.Queries) BankStatementImportRepo {
	return &sqlcdbBankStatementImportRepo{q: q}
}

func (r *sqlcdbBankStatementImportRepo) GetBankStatementImportByID(ctx context.Context, id int64) (*projection.BankStatementImport, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetBankStatementImportByID(ctx, sqlcdb.GetBankStatementImportByIDParams{
		ImportID:   id,
		MerchantID: merchantID,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.BankStatementImportPtrFromBankStatementImport(row), nil
}

func (r *sqlcdbBankStatementImportRepo) GetBankStatementImportByUUID(ctx context.Context, uuid string) (*projection.BankStatementImport, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetBankStatementImportByUUID(ctx, sqlcdb.GetBankStatementImportByUUIDParams{
		ImportUuid: uuid,
		MerchantID: merchantID,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.BankStatementImportPtrFromBankStatementImport(row), nil
}

func (r *sqlcdbBankStatementImportRepo) GetBankStatementImportsPaged(ctx context.Context, ledgerID int64, pageSize, offset int64) ([]projection.BankStatementImport, int64, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.GetBankStatementImportsPaged(ctx, sqlcdb.GetBankStatementImportsPagedParams{
		MerchantID: merchantID,
		LedgerID:   ledgerID,
		PageSize:   pageSize,
		Offset:     offset,
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := r.q.CountBankStatementImports(ctx, sqlcdb.CountBankStatementImportsParams{
		MerchantID: merchantID,
		LedgerID:   ledgerID,
	})
	if err != nil {
		return nil, 0, err
	}
	result := make([]projection.BankStatementImport, len(rows))
	for i, row := range rows {
		result[i] = projection.BankStatementImportFromBankStatementImport(row)
	}
	return result, total, nil
}

func (r *sqlcdbBankStatementImportRepo) GetBankStatementTxnsByImport(ctx context.Context, importID int64) ([]projection.BankStatementTxn, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetBankStatementTxnsByImport(ctx, sqlcdb.GetBankStatementTxnsByImportParams{
		ImportID:   importID,
		MerchantID: merchantID,
	})
	if err != nil {
		return nil, err
	}
	result := make([]projection.BankStatementTxn, len(rows))
	for i, row := range rows {
		result[i] = projection.BankStatementTxnFromBankStatementTxn(row)
	}
	return result, nil
}

func (r *sqlcdbBankStatementImportRepo) GetBankStatementTxnByID(ctx context.Context, id int64) (*projection.BankStatementTxn, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetBankStatementTxnByID(ctx, sqlcdb.GetBankStatementTxnByIDParams{
		BankTxnID:  id,
		MerchantID: merchantID,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.BankStatementTxnPtrFromBankStatementTxn(row), nil
}

func (r *sqlcdbBankStatementImportRepo) GetUnmatchedBankTxns(ctx context.Context, importID int64) ([]projection.BankStatementTxn, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetUnmatchedBankTxns(ctx, sqlcdb.GetUnmatchedBankTxnsParams{
		ImportID:   importID,
		MerchantID: merchantID,
	})
	if err != nil {
		return nil, err
	}
	result := make([]projection.BankStatementTxn, len(rows))
	for i, row := range rows {
		result[i] = projection.BankStatementTxnFromBankStatementTxn(row)
	}
	return result, nil
}

func (r *sqlcdbBankStatementImportRepo) GetBankTxnsForReview(ctx context.Context, importID int64) ([]projection.BankStatementTxn, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetBankTxnsForReview(ctx, sqlcdb.GetBankTxnsForReviewParams{
		ImportID:   importID,
		MerchantID: merchantID,
	})
	if err != nil {
		return nil, err
	}
	result := make([]projection.BankStatementTxn, len(rows))
	for i, row := range rows {
		result[i] = projection.BankStatementTxnFromBankStatementTxn(row)
	}
	return result, nil
}

func (r *sqlcdbBankStatementImportRepo) CountBankTxnsByStatus(ctx context.Context, importID int64, status string) (int64, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return 0, err
	}
	return r.q.CountBankTxnsByStatus(ctx, sqlcdb.CountBankTxnsByStatusParams{
		ImportID:    importID,
		MerchantID:  merchantID,
		MatchStatus: status,
	})
}

func (r *sqlcdbBankStatementImportRepo) GetLedgerEntriesForMatching(ctx context.Context, ledgerID int64, dateFrom, dateTo string) ([]sqlcdb.GetLedgerEntriesForMatchingRow, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	return r.q.GetLedgerEntriesForMatching(ctx, sqlcdb.GetLedgerEntriesForMatchingParams{
		LedgerID:   &ledgerID,
		MerchantID: merchantID,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
	})
}
