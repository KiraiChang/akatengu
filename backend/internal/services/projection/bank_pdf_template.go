package projection

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
)

type BankPdfTemplateProjectionService struct{}

func (s *BankPdfTemplateProjectionService) Name() string {
	return "bank_pdf_template"
}

func (s *BankPdfTemplateProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.Val() {
	case event_types.EventBankPdfTemplateCreated:
		return s.applyCreated(ctx, tx, ct)
	case event_types.EventBankPdfTemplateUpdated:
		return s.applyUpdated(ctx, tx, ct)
	case event_types.EventBankPdfTemplateDeactivated:
		return s.applyDeactivated(ctx, tx, ct)
	}
	return nil
}

func (s *BankPdfTemplateProjectionService) applyCreated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.BankPdfTemplateCreatedPayload](ct)
	if err != nil {
		return err
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	templateID, err := tx.Projection.BankPdfTemplateRepo.InsertBankPdfTemplate(ctx, sqlcdb.InsertBankPdfTemplateParams{
		TemplateUuid: p.TemplateUUID,
		MerchantID:   ct.MerchantID,
		TemplateName: p.TemplateName,
		BankType:     p.BankType,
		UpdatedBy:    updatedBy,
	})
	if err != nil {
		return fmt.Errorf("insert bank pdf template: %w", err)
	}
	for _, l := range p.Ledgers {
		if err := tx.Projection.BankPdfTemplateRepo.InsertBankPdfTemplateLedger(ctx, sqlcdb.InsertBankPdfTemplateLedgerParams{
			TplLedgerUuid: l.TplLedgerUUID,
			TemplateID:    templateID,
			LedgerUuid:    l.LedgerUUID,
			LedgerID:      nil, // ledger_id 在 template 表中留 NULL，查詢時以 ledger_uuid JOIN
			AccountType:   l.AccountType,
			SortOrder:     int64(l.SortOrder),
		}); err != nil {
			return fmt.Errorf("insert bank pdf template ledger %s: %w", l.TplLedgerUUID, err)
		}
	}
	return nil
}

func (s *BankPdfTemplateProjectionService) applyUpdated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.BankPdfTemplateUpdatedPayload](ct)
	if err != nil {
		return err
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	if err := tx.Projection.BankPdfTemplateRepo.UpdateBankPdfTemplate(ctx, sqlcdb.UpdateBankPdfTemplateParams{
		TemplateName: p.TemplateName,
		BankType:     p.BankType,
		UpdatedBy:    updatedBy,
		TemplateUuid: p.TemplateUUID,
		MerchantID:   ct.MerchantID,
	}); err != nil {
		return fmt.Errorf("update bank pdf template: %w", err)
	}
	templateID, err := tx.Projection.BankPdfTemplateRepo.GetIDByUUID(ctx, p.TemplateUUID, ct.MerchantID)
	if err != nil {
		return fmt.Errorf("get bank pdf template id: %w", err)
	}
	if err := tx.Projection.BankPdfTemplateRepo.DeleteBankPdfTemplateLedgers(ctx, templateID); err != nil {
		return fmt.Errorf("delete bank pdf template ledgers: %w", err)
	}
	for _, l := range p.Ledgers {
		if err := tx.Projection.BankPdfTemplateRepo.InsertBankPdfTemplateLedger(ctx, sqlcdb.InsertBankPdfTemplateLedgerParams{
			TplLedgerUuid: l.TplLedgerUUID,
			TemplateID:    templateID,
			LedgerUuid:    l.LedgerUUID,
			LedgerID:      nil,
			AccountType:   l.AccountType,
			SortOrder:     int64(l.SortOrder),
		}); err != nil {
			return fmt.Errorf("insert bank pdf template ledger %s: %w", l.TplLedgerUUID, err)
		}
	}
	return nil
}

func (s *BankPdfTemplateProjectionService) applyDeactivated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.BankPdfTemplateDeactivatedPayload](ct)
	if err != nil {
		return err
	}
	updatedBy := toUpdatedBy(ct.UpdatedBy)
	return tx.Projection.BankPdfTemplateRepo.DeactivateBankPdfTemplate(ctx, sqlcdb.DeactivateBankPdfTemplateParams{
		UpdatedBy:    updatedBy,
		TemplateUuid: p.TemplateUUID,
		MerchantID:   ct.MerchantID,
	})
}
