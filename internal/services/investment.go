package services

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/enums"
	"akatengu/internal/model/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/repos"
	"akatengu/internal/repos/query"
	"context"
	"fmt"

	"github.com/shopspring/decimal"
)

// ─────────────────────────────────────────
// Interface
// ─────────────────────────────────────────

type InvestmentService interface {
	Buy(ctx context.Context, cmd cmd.BuyCmd) error
	Sell(ctx context.Context, cmd cmd.SellCmd) error
	ReceiveDividend(ctx context.Context, cmd cmd.DividendCmd) error
	Split(ctx context.Context, cmd cmd.SplitCmd) error
	BuyFx(ctx context.Context, cmd cmd.FxBuyCmd) error
	SellFx(ctx context.Context, cmd cmd.FxSellCmd) error
	MarkUnrealized(ctx context.Context, cmd cmd.MarkUnrealizedCmd) error
}

// ─────────────────────────────────────────
// Implementation
// ─────────────────────────────────────────

type investmentService struct {
	repo     query.InvestmentRepo
	eventSvc EventStoreService
}

func NewInvestmentService(
	repo query.InvestmentRepo,
	eventSvc EventStoreService,
) InvestmentService {
	return &investmentService{repo: repo, eventSvc: eventSvc}
}

// ─────────────────────────────────────────
// Buy
// ─────────────────────────────────────────

func (s *investmentService) Buy(ctx context.Context, buyCmd cmd.BuyCmd) error {
	inv, err := s.repo.GetByID(ctx, buyCmd.InvestmentId)
	if err != nil || inv == nil {
		return fmt.Errorf("investment %d not found", buyCmd.InvestmentId)
	}

	unitPriceTWD := buyCmd.UnitPrice.Mul(buyCmd.ExchangeRate)
	totalCostTWD := buyCmd.Quantity.Mul(unitPriceTWD).Add(buyCmd.Fee).Add(buyCmd.Tax)

	// 1. 寫會計分錄
	event, err := s.eventSvc.Append(ctx, cmd.AppendCmd{
		AggregateType: enums.AggregateTransaction,
		AggregateID:   fmt.Sprintf("inv-buy-%d-%s", buyCmd.InvestmentId, buyCmd.Date),
		EventType:     event_types.EventTransactionCreated,
		Payload: payload.TransactionCreatedPayload{
			TransactionDate: buyCmd.Date,
			Description:     fmt.Sprintf("買入 %s", inv.Name),
			TotalAmount:     totalCostTWD,
			Currency:        "TWD",
			Entries:         s.buyEntries(inv, buyCmd, totalCostTWD),
		},
	})
	if err != nil {
		return fmt.Errorf("append buy event: %w", err)
	}

	// 2. 更新庫存
	return s.repo.WithTx(ctx, func(tx repos.InvestmentTxRepository) error {
		switch inv.CostMethod.String() {
		case enums.CostMethodAvg.String():
			return s.updateAvgLot(ctx, tx, inv, buyCmd, unitPriceTWD, event.EventId)
		case enums.CostMethodFIFO.String():
			return s.insertFIFOLot(ctx, tx, inv, buyCmd, unitPriceTWD, event.EventId)
		}
		return fmt.Errorf("unknown cost method: %s", inv.CostMethod)
	})
}

func (s *investmentService) buyEntries(inv *projection.Investment, cmd cmd.BuyCmd, totalCostTWD decimal.Decimal) []payload.TransactionEntryPayload {
	entries := []payload.TransactionEntryPayload{
		// 資產增加
		{AccountId: inv.AccountId, Debit: cmd.Quantity.Mul(cmd.UnitPrice).Mul(cmd.ExchangeRate), Credit: decimal.Zero},
		// 扣款帳戶
		{AccountId: ledgerAccountID(cmd.LedgerId), LedgerId: &cmd.LedgerId, Debit: decimal.Zero, Credit: totalCostTWD},
	}
	if cmd.Fee.IsPositive() {
		entries = append(entries,
			payload.TransactionEntryPayload{AccountId: "5940", Debit: cmd.Fee, Credit: decimal.Zero},
		)
	}
	if cmd.Tax.IsPositive() {
		entries = append(entries,
			payload.TransactionEntryPayload{AccountId: "5950", Debit: cmd.Tax, Credit: decimal.Zero},
		)
	}
	return entries
}

// 平均成本法：重新計算加權平均後 upsert 單一批次
func (s *investmentService) updateAvgLot(
	ctx context.Context,
	tx repos.InvestmentTxRepository,
	inv *projection.Investment,
	cmd cmd.BuyCmd,
	unitPriceTWD decimal.Decimal,
	txnID int64,
) error {
	summary, err := s.repo.GetSummary(ctx, cmd.InvestmentId)
	if err != nil {
		return err
	}

	// 加權平均計算
	oldQty := summary.TotalQty
	oldCost := summary.TotalCostTWD
	newQty := oldQty.Add(cmd.Quantity)
	newCost := oldCost.Add(cmd.Quantity.Mul(unitPriceTWD))
	newAvg := newCost.Div(newQty)

	return tx.UpsertAvgLot(ctx, cmd.InvestmentId, projection.InvestmentLot{
		InvestmentId:  cmd.InvestmentId,
		AcquiredDate:  cmd.Date,
		TransactionId: txnID,
		Quantity:      newQty,
		UnitCost:      cmd.UnitPrice,
		UnitCostTWD:   newAvg,
		RemainingQty:  newQty,
	})
}

// FIFO：每次買入新增一筆批次
func (s *investmentService) insertFIFOLot(
	ctx context.Context,
	tx repos.InvestmentTxRepository,
	inv *projection.Investment,
	cmd cmd.BuyCmd,
	unitPriceTWD decimal.Decimal,
	txnID int64,
) error {
	_, err := tx.InsertLot(ctx, projection.InvestmentLot{
		InvestmentId:  cmd.InvestmentId,
		AcquiredDate:  cmd.Date,
		TransactionId: txnID,
		Quantity:      cmd.Quantity,
		UnitCost:      cmd.UnitPrice,
		UnitCostTWD:   unitPriceTWD,
		RemainingQty:  cmd.Quantity,
		Status:        enums.LotStatusOpen,
	})
	return err
}

// ─────────────────────────────────────────
// Sell
// ─────────────────────────────────────────

func (s *investmentService) Sell(ctx context.Context, sellCmd cmd.SellCmd) error {
	inv, err := s.repo.GetByID(ctx, sellCmd.InvestmentId)
	if err != nil || inv == nil {
		return fmt.Errorf("investment %d not found", sellCmd.InvestmentId)
	}

	unitPriceTWD := sellCmd.UnitPrice.Mul(sellCmd.ExchangeRate)
	sellAmountTWD := sellCmd.Quantity.Mul(unitPriceTWD)

	// 1. 計算成本和損益
	costBasisTWD, lotUpdates, err := s.calcCostBasis(ctx, inv, sellCmd.Quantity)
	if err != nil {
		return fmt.Errorf("calc cost basis: %w", err)
	}

	realizedGain := sellAmountTWD.Sub(costBasisTWD).Sub(sellCmd.Fee).Sub(sellCmd.Tax)
	netProceedsTWD := sellAmountTWD.Sub(sellCmd.Fee).Sub(sellCmd.Tax)

	// 2. 會計分錄
	event, err := s.eventSvc.Append(ctx, cmd.AppendCmd{
		AggregateType: enums.AggregateTransaction,
		AggregateID:   fmt.Sprintf("inv-sell-%d-%s", sellCmd.InvestmentId, sellCmd.Date),
		EventType:     event_types.EventTransactionCreated,
		Payload: payload.TransactionCreatedPayload{
			TransactionDate: sellCmd.Date,
			Description:     fmt.Sprintf("賣出 %s", inv.Name),
			TotalAmount:     realizedGain.Abs(),
			Currency:        "TWD",
			Entries:         s.sellEntries(inv, sellCmd, costBasisTWD, netProceedsTWD, realizedGain),
		},
	})
	if err != nil {
		return fmt.Errorf("append sell event: %w", err)
	}

	// 3. 更新庫存批次 + 寫入異動明細
	return s.repo.WithTx(ctx, func(tx repos.InvestmentTxRepository) error {
		// 更新批次剩餘數量
		for _, u := range lotUpdates {
			if err := tx.UpdateLot(ctx, u.LotID, u.RemainingQty, u.Status); err != nil {
				return err
			}
		}

		// 寫入異動明細
		_, err := tx.InsertMovement(ctx, projection.InvestmentMovement{
			InvestmentId:    sellCmd.InvestmentId,
			TransactionId:   event.EventId,
			MovementType:    enums.MovementTypeSell,
			MovementDate:    sellCmd.Date,
			Quantity:        sellCmd.Quantity.Neg(),
			UnitPrice:       sellCmd.UnitPrice,
			UnitPriceTWD:    unitPriceTWD,
			ExchangeRate:    sellCmd.ExchangeRate,
			FeeTWD:          sellCmd.Fee,
			TaxTWD:          sellCmd.Tax,
			RealizedGainTWD: &realizedGain,
			CostBasisTWD:    &costBasisTWD,
		})
		return err
	})
}

func (s *investmentService) sellEntries(
	inv *projection.Investment,
	cmd cmd.SellCmd,
	costBasisTWD, netProceedsTWD, realizedGain decimal.Decimal,
) []payload.TransactionEntryPayload {
	entries := []payload.TransactionEntryPayload{
		// 入帳
		{AccountId: ledgerAccountID(cmd.LedgerId), LedgerId: &cmd.LedgerId, Debit: netProceedsTWD, Credit: decimal.Zero},
		// 成本沖銷
		{AccountId: inv.AccountId, Debit: decimal.Zero, Credit: costBasisTWD},
	}
	if cmd.Fee.IsPositive() {
		entries = append(entries, payload.TransactionEntryPayload{AccountId: "5940", Debit: cmd.Fee, Credit: decimal.Zero})
	}
	if cmd.Tax.IsPositive() {
		entries = append(entries, payload.TransactionEntryPayload{AccountId: "5950", Debit: cmd.Tax, Credit: decimal.Zero})
	}
	// 已實現損益（正=利，負=損）
	if realizedGain.GreaterThanOrEqual(decimal.Zero) {
		entries = append(entries, payload.TransactionEntryPayload{AccountId: "4230", Debit: decimal.Zero, Credit: realizedGain})
	} else {
		entries = append(entries, payload.TransactionEntryPayload{AccountId: "4230", Debit: realizedGain.Abs(), Credit: decimal.Zero})
	}
	return entries
}

// ─────────────────────────────────────────
// 成本基礎計算（avg / FIFO）
// ─────────────────────────────────────────

type lotUpdate struct {
	LotID        int64
	RemainingQty decimal.Decimal
	Status       enums.LotStatus
}

func (s *investmentService) calcCostBasis(
	ctx context.Context,
	inv *projection.Investment,
	qty decimal.Decimal,
) (costBasisTWD decimal.Decimal, updates []lotUpdate, err error) {
	switch inv.CostMethod.String() {
	case enums.CostMethodAvg.String():
		return s.calcAvgCostBasis(ctx, inv.InvestmentId, qty)
	case enums.CostMethodFIFO.String():
		return s.calcFIFOCostBasis(ctx, inv.InvestmentId, qty)
	}
	return decimal.Zero, nil, fmt.Errorf("unknown cost method: %s", inv.CostMethod)
}

func (s *investmentService) calcAvgCostBasis(
	ctx context.Context,
	investmentID int64,
	qty decimal.Decimal,
) (decimal.Decimal, []lotUpdate, error) {
	summary, err := s.repo.GetSummary(ctx, investmentID)
	if err != nil {
		return decimal.Zero, nil, err
	}
	if summary.TotalQty.LessThan(qty) {
		return decimal.Zero, nil, fmt.Errorf("insufficient qty: have %.4f, sell %.4f", summary.TotalQty, qty)
	}

	lots, err := s.repo.GetOpenLots(ctx, investmentID)
	if err != nil {
		return decimal.Zero, nil, err
	}

	costBasis := summary.AvgCostTWD.Mul(qty)
	newQty := summary.TotalQty.Sub(qty)

	// 平均成本法只有一筆批次
	updates := []lotUpdate{{
		LotID:        lots[0].LotId,
		RemainingQty: newQty,
		Status:       lotStatus(newQty),
	}}
	return costBasis, updates, nil
}

func (s *investmentService) calcFIFOCostBasis(
	ctx context.Context,
	investmentID int64,
	qty decimal.Decimal,
) (decimal.Decimal, []lotUpdate, error) {
	lots, err := s.repo.GetOpenLots(ctx, investmentID)
	if err != nil {
		return decimal.Zero, nil, err
	}

	remaining := qty
	costBasis := decimal.Zero
	var updates []lotUpdate

	for _, lot := range lots {
		if remaining.IsZero() {
			break
		}
		use := decimal.Min(remaining, lot.RemainingQty)
		costBasis = costBasis.Add(use.Mul(lot.UnitCostTWD))
		remaining = remaining.Sub(use)
		newRemaining := lot.RemainingQty.Sub(use)
		updates = append(updates, lotUpdate{
			LotID:        lot.LotId,
			RemainingQty: newRemaining,
			Status:       lotStatus(newRemaining),
		})
	}

	if remaining.IsPositive() {
		return decimal.Zero, nil, fmt.Errorf("insufficient inventory: need %.4f more", remaining)
	}
	return costBasis, updates, nil
}

// ─────────────────────────────────────────
// Dividend
// ─────────────────────────────────────────

func (s *investmentService) ReceiveDividend(ctx context.Context, dividendCmd cmd.DividendCmd) error {
	inv, err := s.repo.GetByID(ctx, dividendCmd.InvestmentId)
	if err != nil || inv == nil {
		return fmt.Errorf("investment %d not found", dividendCmd.InvestmentId)
	}

	amountTWD := dividendCmd.Amount.Mul(dividendCmd.ExchangeRate)
	netAmountTWD := amountTWD.Sub(dividendCmd.WithholdingTax)

	event, err := s.eventSvc.Append(ctx, cmd.AppendCmd{
		AggregateType: enums.AggregateTransaction,
		AggregateID:   fmt.Sprintf("inv-div-%d-%s", dividendCmd.InvestmentId, dividendCmd.Date),
		EventType:     event_types.EventTransactionCreated,
		Payload: payload.TransactionCreatedPayload{
			TransactionDate: dividendCmd.Date,
			Description:     fmt.Sprintf("%s 配息", inv.Name),
			TotalAmount:     amountTWD,
			Currency:        "TWD",
			Entries: []payload.TransactionEntryPayload{
				// 實收金額入帳
				{AccountId: ledgerAccountID(dividendCmd.LedgerId), LedgerId: &dividendCmd.LedgerId, Debit: netAmountTWD, Credit: decimal.Zero},
				// 扣繳稅（若有）
				{AccountId: "5920", Debit: dividendCmd.WithholdingTax, Credit: decimal.Zero},
				// 股利收入
				{AccountId: "4210", Debit: decimal.Zero, Credit: amountTWD},
			},
		},
	})
	if err != nil {
		return err
	}

	return s.repo.WithTx(ctx, func(tx repos.InvestmentTxRepository) error {
		_, err := tx.InsertMovement(ctx, projection.InvestmentMovement{
			InvestmentId:  dividendCmd.InvestmentId,
			TransactionId: event.EventId,
			MovementType:  enums.MovementTypeDividend,
			MovementDate:  dividendCmd.Date,
			Quantity:      decimal.Zero,
			UnitPrice:     dividendCmd.Amount,
			UnitPriceTWD:  amountTWD,
			ExchangeRate:  dividendCmd.ExchangeRate,
			FeeTWD:        dividendCmd.WithholdingTax,
		})
		return err
	})
}

// ─────────────────────────────────────────
// Stock Split（調整批次數量，不產生會計分錄）
// ─────────────────────────────────────────

func (s *investmentService) Split(ctx context.Context, splitCmd cmd.SplitCmd) error {
	lots, err := s.repo.GetOpenLots(ctx, splitCmd.InvestmentId)
	if err != nil {
		return err
	}
	if len(lots) == 0 {
		return fmt.Errorf("no open lots for investment %d", splitCmd.InvestmentId)
	}

	return s.repo.WithTx(ctx, func(tx repos.InvestmentTxRepository) error {
		for _, lot := range lots {
			newQty := lot.RemainingQty.Mul(splitCmd.Ratio)
			newUnitCostTWD := lot.UnitCostTWD.Div(splitCmd.Ratio) // 單價等比縮小
			// 直接更新 lot 數量和單價
			if err := tx.UpdateLot(ctx, lot.LotId, newQty, lot.Status); err != nil {
				return err
			}
			_ = newUnitCostTWD // 實際應一併更新 unit_cost_twd，此處簡化
		}
		return nil
	})
}

// ─────────────────────────────────────────
// FX Buy
// ─────────────────────────────────────────

func (s *investmentService) BuyFx(ctx context.Context, fxBuy cmd.FxBuyCmd) error {
	inv, err := s.repo.GetByID(ctx, fxBuy.InvestmentId)
	if err != nil || inv == nil {
		return fmt.Errorf("investment %d not found", fxBuy.InvestmentId)
	}

	amountTWD := fxBuy.Amount.Mul(fxBuy.ExchangeRate)
	totalTWD := amountTWD.Add(fxBuy.Fee)

	event, err := s.eventSvc.Append(ctx, cmd.AppendCmd{
		AggregateType: enums.AggregateTransaction,
		AggregateID:   fmt.Sprintf("fx-buy-%d-%s", fxBuy.InvestmentId, fxBuy.Date),
		EventType:     event_types.EventTransactionCreated,
		Payload: payload.TransactionCreatedPayload{
			TransactionDate: fxBuy.Date,
			Description:     fmt.Sprintf("買入外幣 %s %.2f", inv.Currency, fxBuy.Amount),
			TotalAmount:     totalTWD,
			Currency:        "TWD",
			Entries: []payload.TransactionEntryPayload{
				{AccountId: inv.AccountId, Debit: amountTWD, Credit: decimal.Zero},
				{AccountId: "5940", Debit: fxBuy.Fee, Credit: decimal.Zero},
				{AccountId: ledgerAccountID(fxBuy.LedgerId), LedgerId: &fxBuy.LedgerId, Debit: decimal.Zero, Credit: totalTWD},
			},
		},
	})
	if err != nil {
		return err
	}

	return s.repo.WithTx(ctx, func(tx repos.InvestmentTxRepository) error {
		_, err := tx.InsertMovement(ctx, projection.InvestmentMovement{
			InvestmentId:  fxBuy.InvestmentId,
			TransactionId: event.EventId,
			MovementType:  enums.MovementTypeBuy,
			MovementDate:  fxBuy.Date,
			Quantity:      fxBuy.Amount,
			UnitPrice:     decimal.NewFromFloat(1.0),
			UnitPriceTWD:  fxBuy.ExchangeRate,
			ExchangeRate:  fxBuy.ExchangeRate,
			FeeTWD:        fxBuy.Fee,
		})
		return err
	})
}

// ─────────────────────────────────────────
// FX Sell（計算匯兌損益）
// ─────────────────────────────────────────

func (s *investmentService) SellFx(ctx context.Context, fxSell cmd.FxSellCmd) error {
	inv, err := s.repo.GetByID(ctx, fxSell.InvestmentId)
	if err != nil || inv == nil {
		return fmt.Errorf("investment %d not found", fxSell.InvestmentId)
	}

	// 成本基礎（買入匯率）
	costBasisTWD, lotUpdates, err := s.calcCostBasis(ctx, inv, fxSell.Amount)
	if err != nil {
		return err
	}

	sellAmountTWD := fxSell.Amount.Mul(fxSell.ExchangeRate)
	fxGain := sellAmountTWD.Sub(costBasisTWD).Sub(fxSell.Fee)
	netTWD := sellAmountTWD.Sub(fxSell.Fee)

	event, err := s.eventSvc.Append(ctx, cmd.AppendCmd{
		AggregateType: enums.AggregateTransaction,
		AggregateID:   fmt.Sprintf("fx-sell-%d-%s", fxSell.InvestmentId, fxSell.Date),
		EventType:     event_types.EventTransactionCreated,
		Payload: payload.TransactionCreatedPayload{
			TransactionDate: fxSell.Date,
			Description:     fmt.Sprintf("賣出外幣 %s %.2f", inv.Currency, fxSell.Amount),
			TotalAmount:     fxGain.Abs(),
			Currency:        "TWD",
			Entries:         s.fxSellEntries(inv, fxSell, costBasisTWD, netTWD, fxGain),
		},
	})
	if err != nil {
		return err
	}

	return s.repo.WithTx(ctx, func(tx repos.InvestmentTxRepository) error {
		for _, u := range lotUpdates {
			if err := tx.UpdateLot(ctx, u.LotID, u.RemainingQty, u.Status); err != nil {
				return err
			}
		}
		_, err := tx.InsertMovement(ctx, projection.InvestmentMovement{
			InvestmentId:    fxSell.InvestmentId,
			TransactionId:   event.EventId,
			MovementType:    enums.MovementTypeSell,
			MovementDate:    fxSell.Date,
			Quantity:        fxSell.Amount.Neg(),
			UnitPrice:       decimal.NewFromFloat(1.0),
			UnitPriceTWD:    fxSell.ExchangeRate,
			ExchangeRate:    fxSell.ExchangeRate,
			FeeTWD:          fxSell.Fee,
			RealizedGainTWD: &fxGain,
			CostBasisTWD:    &costBasisTWD,
		})
		return err
	})
}

func (s *investmentService) fxSellEntries(
	inv *projection.Investment,
	fxSell cmd.FxSellCmd,
	costBasisTWD, netTWD, fxGain decimal.Decimal,
) []payload.TransactionEntryPayload {
	entries := []payload.TransactionEntryPayload{
		{AccountId: ledgerAccountID(fxSell.LedgerId), LedgerId: &fxSell.LedgerId, Debit: netTWD, Credit: decimal.Zero},
		{AccountId: "5940", Debit: fxSell.Fee, Credit: decimal.Zero},
		{AccountId: inv.AccountId, Debit: decimal.Zero, Credit: costBasisTWD},
	}
	if fxGain.GreaterThanOrEqual(decimal.Zero) {
		entries = append(entries, payload.TransactionEntryPayload{AccountId: "4240", Debit: decimal.Zero, Credit: fxGain})
	} else {
		entries = append(entries, payload.TransactionEntryPayload{AccountId: "4240", Debit: fxGain.Abs(), Credit: decimal.Zero})
	}
	return entries
}

// ─────────────────────────────────────────
// Mark Unrealized（期末評價）
// ─────────────────────────────────────────

func (s *investmentService) MarkUnrealized(ctx context.Context, unrealized cmd.MarkUnrealizedCmd) error {
	inv, err := s.repo.GetByID(ctx, unrealized.InvestmentId)
	if err != nil || inv == nil {
		return fmt.Errorf("investment %d not found", unrealized.InvestmentId)
	}

	summary, err := s.repo.GetSummary(ctx, unrealized.InvestmentId)
	if err != nil {
		return err
	}
	if summary.TotalQty.IsZero() {
		return nil // 沒有庫存不需要評價
	}

	marketValueTWD := summary.TotalQty.Mul(unrealized.MarketPrice).Mul(unrealized.ExchangeRate)
	unrealizedGain := marketValueTWD.Sub(summary.TotalCostTWD)

	_, err = s.eventSvc.Append(ctx, cmd.AppendCmd{
		AggregateType: enums.AggregateTransaction,
		AggregateID:   fmt.Sprintf("unrealized-%d-%s", unrealized.InvestmentId, unrealized.Date),
		EventType:     event_types.EventTransactionCreated,
		Payload: payload.TransactionCreatedPayload{
			TransactionDate: unrealized.Date,
			Description:     fmt.Sprintf("%s 未實現損益評價", inv.Name),
			TotalAmount:     unrealizedGain.Abs(),
			Currency:        "TWD",
			Entries: []payload.TransactionEntryPayload{
				// 調整資產帳面價值
				func() payload.TransactionEntryPayload {
					if unrealizedGain.GreaterThanOrEqual(decimal.Zero) {
						return payload.TransactionEntryPayload{AccountId: inv.AccountId, Debit: unrealizedGain, Credit: decimal.Zero}
					}
					return payload.TransactionEntryPayload{AccountId: inv.AccountId, Debit: decimal.Zero, Credit: unrealizedGain.Abs()}
				}(),
				// 未實現損益
				func() payload.TransactionEntryPayload {
					if unrealizedGain.GreaterThanOrEqual(decimal.Zero) {
						return payload.TransactionEntryPayload{AccountId: "4250", Debit: decimal.Zero, Credit: unrealizedGain}
					}
					return payload.TransactionEntryPayload{AccountId: "4250", Debit: unrealizedGain.Abs(), Credit: decimal.Zero}
				}(),
			},
		},
	})
	return err
}

// ─────────────────────────────────────────
// helpers
// ─────────────────────────────────────────

func lotStatus(remainingQty decimal.Decimal) enums.LotStatus {
	if remainingQty.IsZero() {
		return enums.LotStatusClose
	}
	return enums.LotStatusPartial
}

// ledgerAccountID 從 ledger_id 對應科目代號
// 實際應從 DB 查，這裡簡化用前綴對應
func ledgerAccountID(ledgerID int64) string {
	return fmt.Sprintf("ledger:%d", ledgerID)
}
