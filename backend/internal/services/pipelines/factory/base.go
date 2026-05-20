package factory

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/enums/sys_codes"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
)

// ------------------------------
// factory
// ------------------------------

func NewPipelineRegistry(query *query.Repo) pipelines.PipelineRegistry {
	p := &registry{
		rules: make(map[event_types.EventType]pipelines.PipelineRunner),
	}
	p.register(query)
	return p
}

type registry struct {
	rules map[event_types.EventType]pipelines.PipelineRunner
}

func (p *registry) Dispatch(ctx context.Context, cmd cmd.AppendCmd) (*pipelines.Result, error) {
	result, ok := p.rules[cmd.EventType]
	if !ok {
		return nil, fmt.Errorf("registry %s not found", cmd.EventType.String())
	}
	return result.Run(ctx, cmd)
}

func (p *registry) register(query *query.Repo) {
	// Accounts
	p.rules[event_types.EventAccountCreated.Enum()] = NewEventAccountCreatedPipeline()
	p.rules[event_types.EventAccountUpdated.Enum()] = NewEventAccountUpdatedPipeline()
	p.rules[event_types.EventLedgerAccountCreated.Enum()] = NewEventLedgerAccountCreatedPipeline()
	p.rules[event_types.EventLedgerAccountUpdated.Enum()] = NewEventLedgerAccountUpdatedPipeline()

	// Transaction
	p.rules[event_types.EventTransactionCreated.Enum()] = NewEventTransactionCreatedPipeline(query)
	p.rules[event_types.EventTransactionVoided.Enum()] = NewEventTransactionVoidedPipeline(query)
	p.rules[event_types.EventTransactionCorrected.Enum()] = NewEventTransactionCorrectedPipeline(query)

	p.rules[event_types.EventPeriodMonthStarted.Enum()] = NewEventPeriodMonthStartedPipeline(query)
	p.rules[event_types.EventPeriodMonthClosed.Enum()] = NewEventPeriodMonthClosedPipeline(query)
	p.rules[event_types.EventPeriodMonthReopened.Enum()] = NewEventPeriodMonthReopenedPipeline(query)

	p.rules[event_types.EventPeriodAnnualStarted.Enum()] = NewEventPeriodAnnualStartedPipeline(query)
	p.rules[event_types.EventPeriodAnnualClosed.Enum()] = NewEventPeriodAnnualClosedPipeline(query)
	p.rules[event_types.EventPeriodAnnualReopened.Enum()] = NewEventPeriodAnnualReopenedPipeline(query)

	// Investment(Account Agg)
	p.rules[event_types.EventInvestmentCreated.Enum()] = NewEventInvestmentCreatedPipeline()
	p.rules[event_types.EventInvestmentUpdated.Enum()] = NewEventInvestmentUpdatedPipeline()

	// Investment(Transaction Agg)
	p.rules[event_types.EventInvestmentBought.Enum()] = NewEventInvestmentBoughtPipeline(query)
	p.rules[event_types.EventInvestmentSold.Enum()] = NewEventInvestmentSoldPipeline(query)
	p.rules[event_types.EventStockSplit.Enum()] = NewEventStockSplitPipeline(query)
	p.rules[event_types.EventDividendReceived.Enum()] = NewEventDividendReceivedPipeline(query)

	p.rules[event_types.EventRateUpdated.Enum()] = NewEventRateUpdatedPipeline(query)
	p.rules[event_types.EventUnrealizedMarked.Enum()] = NewEventUnrealizedMarkedPipeline(query)

	// Installment
	p.rules[event_types.EventInstallmentCreated.Enum()] = NewEventInstallmentCreatedPipeline(query)
	p.rules[event_types.EventInstallmentPeriodPaid.Enum()] = NewEventInstallmentPeriodPaidPipeline(query)
}

func getSysAccountCode(ctx context.Context, sys query.SysRepo, enum sys_codes.SysAccount) (string, error) {
	return getSysAccountCodeByString(ctx, sys, enum.String())
}

func getSysAccountCodeByString(ctx context.Context, sys query.SysRepo, sysCode string) (string, error) {
	codes, err := sys.GetSysAccount(ctx)
	if err != nil {
		return "", err
	}

	for _, code := range codes {
		if code.SysCode == sysCode {
			return code.AccountId, nil
		}
	}
	return "", fmt.Errorf("sys account not found")
}


func validPeriodMonthlyStatus(ctx context.Context, date string, query *query.Repo, status enums.PeriodTypeStatus) error {
	// 1. 轉換開張日期
	periodStart, err := periodStartDate(date)
	if err != nil {
		return fmt.Errorf("parse date %s, fail: %w", date, err)
	}

	// 2. 檢查是否已結帳
	existing, err := query.Period.GetByPeriod(ctx, enums.PeriodMonthly.Enum(), periodStart)
	if err != nil {
		fmt.Errorf("get period: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("period %s is not exists", periodStart)
	}

	if !existing.Status.Is(status.Val()) {
		return fmt.Errorf("period %s~%s is already closed", existing.PeriodStart, existing.PeriodEnd)
	}
	return nil
}
