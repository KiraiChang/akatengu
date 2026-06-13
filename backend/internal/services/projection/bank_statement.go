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
	if err := tx.Projection.BankCsvTemplateRepo.InsertBankCsvTemplate(ctx, sqlcdb.InsertBankCsvTemplateParams{
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
	}); err != nil {
		return err
	}
	if len(p.Ledgers) > 0 {
		templateID, err := tx.Projection.BankCsvTemplateRepo.GetIDByUUID(ctx, ct.Event.EventUuid, ct.MerchantID)
		if err != nil {
			return fmt.Errorf("get bank csv template id: %w", err)
		}
		for _, l := range p.Ledgers {
			if err := tx.Projection.BankCsvTemplateRepo.InsertBankCsvTemplateLedger(ctx, sqlcdb.InsertBankCsvTemplateLedgerParams{
				TplLedgerUuid: l.TplLedgerUUID,
				TemplateID:    templateID,
				LedgerUuid:    l.LedgerUUID,
				LedgerID:      nil,
				AccountType:   l.AccountType,
				SortOrder:     int64(l.SortOrder),
			}); err != nil {
				return fmt.Errorf("insert bank csv template ledger %s: %w", l.TplLedgerUUID, err)
			}
		}
	}
	return nil
}

func (s *BankCsvTemplateProjectionService) applyUpdated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.BankCsvTemplateUpdatedPayload](ct)
	if err != nil {
		return err
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	if err := tx.Projection.BankCsvTemplateRepo.UpdateBankCsvTemplate(ctx, sqlcdb.UpdateBankCsvTemplateParams{
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
	}); err != nil {
		return err
	}
	templateID, err := tx.Projection.BankCsvTemplateRepo.GetIDByUUID(ctx, p.TemplateUUID, ct.MerchantID)
	if err != nil {
		return fmt.Errorf("get bank csv template id: %w", err)
	}
	if err := tx.Projection.BankCsvTemplateRepo.DeleteBankCsvTemplateLedgers(ctx, templateID); err != nil {
		return fmt.Errorf("delete bank csv template ledgers: %w", err)
	}
	for _, l := range p.Ledgers {
		if err := tx.Projection.BankCsvTemplateRepo.InsertBankCsvTemplateLedger(ctx, sqlcdb.InsertBankCsvTemplateLedgerParams{
			TplLedgerUuid: l.TplLedgerUUID,
			TemplateID:    templateID,
			LedgerUuid:    l.LedgerUUID,
			LedgerID:      nil,
			AccountType:   l.AccountType,
			SortOrder:     int64(l.SortOrder),
		}); err != nil {
			return fmt.Errorf("insert bank csv template ledger %s: %w", l.TplLedgerUUID, err)
		}
	}
	return nil
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
	importSource := p.ImportSource
	if importSource == "" {
		importSource = "CSV"
	}
	importID, err := tx.Projection.BankStatementImportRepo.InsertBankStatementImport(ctx, sqlcdb.InsertBankStatementImportParams{
		ImportUuid:      ct.Event.EventUuid,
		MerchantID:      ct.MerchantID,
		LedgerID:        p.LedgerID,
		TemplateID:      p.TemplateID,
		TemplateUuid:    p.TemplateUUID,
		PdfTemplateID:   nil,
		PdfTemplateUuid: p.PdfTemplateUUID,
		BankType:        p.BankType,
		StatementDate:   p.StatementDate,
		ImportSource:    importSource,
		Filename:        p.Filename,
		Note:            p.Note,
		UpdatedBy:       updatedBy,
	})
	if err != nil {
		return fmt.Errorf("insert bank statement import: %w", err)
	}

	for _, l := range p.Ledgers {
		if err := tx.Projection.BankStatementImportRepo.InsertBankStatementImportLedger(ctx, sqlcdb.InsertBankStatementImportLedgerParams{
			ImportLedgerUuid: l.ImportLedgerUUID,
			ImportID:         importID,
			LedgerUuid:       l.LedgerUUID,
			LedgerID:         nil,
			AccountType:      l.AccountType,
		}); err != nil {
			return fmt.Errorf("insert bank statement import ledger %s: %w", l.ImportLedgerUUID, err)
		}
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
			LedgerUuid:  txn.LedgerUUID,
			LedgerID:    nil,
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
