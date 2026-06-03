package projection

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"

	"github.com/shopspring/decimal"
)

// ─────────────────────────────────────────
// BankCsvTemplate Projection
// ─────────────────────────────────────────

type BankCsvTemplateProjectionService struct{}

func (s *BankCsvTemplateProjectionService) Name() string {
	return "bank_csv_template"
}

func (s *BankCsvTemplateProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.Val() {
	case event_types.EventBankCsvTemplateCreated:
		return s.applyCreated(ctx, tx, ct)
	case event_types.EventBankCsvTemplateUpdated:
		return s.applyUpdated(ctx, tx, ct)
	case event_types.EventBankCsvTemplateDeactivated:
		return s.applyDeactivated(ctx, tx, ct)
	}
	return nil
}

func (s *BankCsvTemplateProjectionService) applyCreated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.BankCsvTemplateCreatedPayload](ct)
	if err != nil {
		return err
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	encoding := p.Encoding
	if encoding == "" {
		encoding = "UTF-8"
	}
	return tx.Projection.BankCsvTemplateRepo.InsertBankCsvTemplate(ctx, sqlcdb.InsertBankCsvTemplateParams{
		TemplateUuid:      ct.Event.EventUuid,
		MerchantID:        ct.MerchantID,
		TemplateName:      p.TemplateName,
		Encoding:          encoding,
		SkipRows:          p.SkipRows,
		DateColumn:        p.DateColumn,
		DateFormat:        p.DateFormat,
		DescriptionColumn: p.DescriptionColumn,
		DebitColumn:       p.DebitColumn,
		CreditColumn:      p.CreditColumn,
		AmountColumn:      p.AmountColumn,
		BalanceColumn:     p.BalanceColumn,
		ReferenceColumn:   p.ReferenceColumn,
		Note:              p.Note,
		UpdatedBy:         updatedBy,
	})
}

func (s *BankCsvTemplateProjectionService) applyUpdated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.BankCsvTemplateUpdatedPayload](ct)
	if err != nil {
		return err
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	return tx.Projection.BankCsvTemplateRepo.UpdateBankCsvTemplate(ctx, sqlcdb.UpdateBankCsvTemplateParams{
		TemplateName:      p.TemplateName,
		Encoding:          p.Encoding,
		SkipRows:          p.SkipRows,
		DateColumn:        p.DateColumn,
		DateFormat:        p.DateFormat,
		DescriptionColumn: p.DescriptionColumn,
		DebitColumn:       p.DebitColumn,
		CreditColumn:      p.CreditColumn,
		AmountColumn:      p.AmountColumn,
		BalanceColumn:     p.BalanceColumn,
		ReferenceColumn:   p.ReferenceColumn,
		Note:              p.Note,
		UpdatedBy:         updatedBy,
		TemplateUuid:      p.TemplateUUID,
		MerchantID:        ct.MerchantID,
	})
}

func (s *BankCsvTemplateProjectionService) applyDeactivated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.BankCsvTemplateDeactivatedPayload](ct)
	if err != nil {
		return err
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	return tx.Projection.BankCsvTemplateRepo.DeactivateBankCsvTemplate(ctx, sqlcdb.DeactivateBankCsvTemplateParams{
		UpdatedBy:    updatedBy,
		TemplateUuid: p.TemplateUUID,
		MerchantID:   ct.MerchantID,
	})
}

// ─────────────────────────────────────────
// BankStatement Projection
// ─────────────────────────────────────────

type BankStatementProjectionService struct{}

func (s *BankStatementProjectionService) Name() string {
	return "bank_statement"
}

func (s *BankStatementProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.Val() {
	case event_types.EventBankStatementImported:
		return s.applyImported(ctx, tx, ct)
	case event_types.EventBankStatementTxnMatched:
		return s.applyTxnMatched(ctx, tx, ct)
	case event_types.EventBankStatementTxnIgnored:
		return s.applyTxnIgnored(ctx, tx, ct)
	case event_types.EventBankStatementAdjustmentApproved:
		return s.applyAdjustmentApproved(ctx, tx, ct)
	case event_types.EventBankStatementCompleted:
		return s.applyCompleted(ctx, tx, ct)
	}
	return nil
}

func (s *BankStatementProjectionService) applyImported(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.BankStatementImportedPayload](ct)
	if err != nil {
		return err
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	importID, err := tx.Projection.BankStatementImportRepo.InsertBankStatementImport(ctx, sqlcdb.InsertBankStatementImportParams{
		ImportUuid:    ct.Event.EventUuid,
		MerchantID:    ct.MerchantID,
		LedgerID:      p.LedgerID,
		TemplateID:    p.TemplateID,
		StatementDate: p.StatementDate,
		ImportSource:  "CSV",
		Filename:      p.Filename,
		Note:          p.Note,
		UpdatedBy:     updatedBy,
	})
	if err != nil {
		return fmt.Errorf("insert bank statement import: %w", err)
	}

	for _, txn := range p.Transactions {
		var bal decimal.NullDecimal
		if txn.Balance != nil {
			bal = decimal.NewNullDecimal(*txn.Balance)
		}
		if err := tx.Projection.BankStatementImportRepo.InsertBankStatementTxn(ctx, sqlcdb.InsertBankStatementTxnParams{
			BankTxnUuid: txn.BankTxnUUID,
			ImportID:    importID,
			MerchantID:  ct.MerchantID,
			TxnDate:     txn.TxnDate,
			Description: txn.Description,
			Debit:       txn.Debit,
			Credit:      txn.Credit,
			Balance:     bal,
			ReferenceNo: txn.ReferenceNo,
		}); err != nil {
			return fmt.Errorf("insert bank statement txn %s: %w", txn.BankTxnUUID, err)
		}
	}
	return nil
}

func (s *BankStatementProjectionService) applyTxnMatched(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.BankStatementTxnMatchedPayload](ct)
	if err != nil {
		return err
	}
	return tx.Projection.BankStatementImportRepo.UpdateBankTxnMatch(ctx, sqlcdb.UpdateBankTxnMatchParams{
		MerchantID:      ct.MerchantID,
		BankTxnID:       p.BankTxnID,
		MatchStatus:     p.MatchStatus,
		MatchedEntryID:  p.MatchedEntryID,
		MatchConfidence: &p.MatchConfidence,
	})
}

func (s *BankStatementProjectionService) applyTxnIgnored(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.BankStatementTxnIgnoredPayload](ct)
	if err != nil {
		return err
	}
	return tx.Projection.BankStatementImportRepo.UpdateBankTxnStatus(ctx, sqlcdb.UpdateBankTxnStatusParams{
		MerchantID:  ct.MerchantID,
		BankTxnID:   p.BankTxnID,
		MatchStatus: "IGNORED",
	})
}

func (s *BankStatementProjectionService) applyAdjustmentApproved(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.BankStatementAdjustmentApprovedPayload](ct)
	if err != nil {
		return err
	}
	return tx.Projection.BankStatementImportRepo.UpdateBankTxnCreatedTxn(ctx, sqlcdb.UpdateBankTxnCreatedTxnParams{
		MerchantID:    ct.MerchantID,
		BankTxnID:     p.BankTxnID,
		CreatedTxnID:  &p.CreatedTxnID,
	})
}

func (s *BankStatementProjectionService) applyCompleted(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	st, err := checkAndGetState[state.BankStatementCompletedState](ct)
	if err != nil {
		return err
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	return tx.Projection.BankStatementImportRepo.UpdateBankStatementImportStatus(ctx, sqlcdb.UpdateBankStatementImportStatusParams{
		MerchantID: ct.MerchantID,
		ImportID:   st.ImportID,
		Status:     "COMPLETED",
		UpdatedBy:  updatedBy,
	})
}
