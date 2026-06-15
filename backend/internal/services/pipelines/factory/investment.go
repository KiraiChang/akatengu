package factory

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// ------------------------------
// EventRateUpdated
// ------------------------------

type eventRateUpdatedProjector struct {
	query *query.Repo
}

func (e eventRateUpdatedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.RateUpdatedPayload]) error {
	return ct.Payload.Validate()
}

func NewEventRateUpdatedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.RateUpdatedPayload] {
	return pipelines.NewTypeWithNoState[payload.RateUpdatedPayload](&eventRateUpdatedProjector{query})
}

// ------------------------------
// EventInvestmentBought
// ------------------------------

type eventInvestmentBoughtProjector struct {
	query *query.Repo
}

func (e eventInvestmentBoughtProjector) Project(ctx context.Context, ct *pipelines.Context[state.InvestmentBoughtState, payload.InvestmentBoughtPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload

	if err := validPeriodMonthlyStatus(ctx, p.Date, e.query, enums.PeriodTypeStatusOpen.Enum()); err != nil {
		return err
	}

	inv, err := e.query.Investment.GetByCreationEventUuid(ctx, p.InvestmentUUID)
	if err != nil {
		return err
	}

	if inv == nil {
		return fmt.Errorf("investment %s not found", p.InvestmentUUID)
	}
	ledger, err := e.query.Account.GetLedger(ctx, p.LedgerId)
	if err != nil {
		return err
	}

	ct.State.Ledger = *ledger
	ct.State.Investment = *inv
	ct.State.Movement = projection.InvestmentMovement{
		InvestmentId: inv.InvestmentId,
		MovementType: enums.MovementTypeBuy.Enum(),
		MovementDate: p.Date,
		Quantity:     p.Quantity,
		UnitPrice:    p.UnitPrice,
		UnitPriceTWD: p.UnitPrice.Mul(p.ExchangeRate),
		ExchangeRate: p.ExchangeRate,
		Fee:          p.Fee,
		Tax:          p.Tax,
	}
	if inv.CostMethod.Is(enums.CostMethodAvg) {
		ct.State.Position = projection.InvestmentPosition{
			InvestmentId:  inv.InvestmentId,
			TotalQuantity: p.Quantity,
			TotalCost:     p.UnitPrice.Mul(p.Quantity),
		}
	} else {
		ct.State.Lot = projection.InvestmentLot{
			InvestmentId: inv.InvestmentId,
			AcquiredDate: p.Date,
			Quantity:     p.Quantity,
			UnitCost:     p.UnitPrice.Mul(p.ExchangeRate),
			TotalCost:    p.UnitPrice.Mul(p.Quantity),
			RemainingQty: p.Quantity,
			Status:       enums.LotStatusOpen.Enum(),
		}
	}

	entries, err := e.buyEntries(ctx, ct.State, &p)
	if err != nil {
		return err
	}
	ct.State.Transaction = payload.TransactionCreatedPayload{
		TransactionDate: p.Date,
		Description:     fmt.Sprintf("買入 %s", inv.Name),
		Currency:        "TWD",
		Entries:         entries,
	}

	return nil
}

func (e eventInvestmentBoughtProjector) buyEntries(ctx context.Context, st *state.InvestmentBoughtState, p *payload.InvestmentBoughtPayload) ([]payload.TransactionEntryPayload, error) {
	config, err := e.query.Config.GetAssetTypeAccountConfig(ctx, st.Investment.AssetType)
	if err != nil {
		return nil, fmt.Errorf("get asset type config: %w", err)
	}
	cost := p.Quantity.Mul(p.UnitPrice).Mul(p.ExchangeRate)
	totalCost := cost.Add(p.Fee).Add(p.Tax)
	entries := []payload.TransactionEntryPayload{
		{AccountId: st.Investment.AccountId, Debit: cost, Credit: decimal.Zero},
		{AccountId: st.Ledger.AccountId, LedgerId: &p.LedgerId, Debit: decimal.Zero, Credit: totalCost},
	}
	if p.Fee.IsPositive() {
		entries = append(entries, payload.TransactionEntryPayload{AccountId: config.FeeAccountID, Debit: p.Fee, Credit: decimal.Zero})
	}
	if p.Tax.IsPositive() {
		entries = append(entries, payload.TransactionEntryPayload{AccountId: config.TaxAccountID, Debit: p.Tax, Credit: decimal.Zero})
	}
	return entries, nil
}

func NewEventInvestmentBoughtPipeline(query *query.Repo) *pipelines.TypedPipeline[state.InvestmentBoughtState, payload.InvestmentBoughtPayload] {
	return pipelines.NewType[state.InvestmentBoughtState, payload.InvestmentBoughtPayload](&eventInvestmentBoughtProjector{query}, func() *state.InvestmentBoughtState {
		return &state.InvestmentBoughtState{}
	})
}

// ------------------------------
// EventInvestmentSold
// ------------------------------

type eventInvestmentSoldProjector struct {
	query *query.Repo
}

func (e eventInvestmentSoldProjector) Project(ctx context.Context, ct *pipelines.Context[state.InvestmentSoldState, payload.InvestmentSoldPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload

	if err := validPeriodMonthlyStatus(ctx, p.Date, e.query, enums.PeriodTypeStatusOpen.Enum()); err != nil {
		return err
	}

	inv, err := e.query.Investment.GetByCreationEventUuid(ctx, p.InvestmentUUID)
	if err != nil {
		return err
	}

	if inv == nil {
		return fmt.Errorf("investment %s not found", p.InvestmentUUID)
	}
	ledger, err := e.query.Account.GetLedger(ctx, p.LedgerId)
	if err != nil {
		return err
	}

	ct.State.Ledger = *ledger
	ct.State.Investment = *inv

	unitPrice := p.UnitPrice.Mul(p.ExchangeRate)
	sellAmount := p.Quantity.Mul(unitPrice)

	// 1. 計算成本和損益
	var originalCostBasis decimal.Decimal

	switch inv.CostMethod.Val() {
	case enums.CostMethodAvg:
		position, err := e.query.Investment.GetPosition(ctx, inv.InvestmentId)
		if err != nil {
			return fmt.Errorf("calc avg cost fail: %w", err)
		}
		originalCostBasis = position.AvgCost.Mul(p.Quantity)
		ct.State.CostBasis = originalCostBasis
		ct.State.Position = projection.InvestmentPosition{
			InvestmentId:  inv.InvestmentId,
			TotalQuantity: p.Quantity,
			TotalCost:     originalCostBasis,
		}
		// FVTPL/FVOCI：若曾評價則以 FV 沖銷資產帳，並記錄需沖回的累積未實現
		switch inv.IFRSCategory.Val() {
		case enums.IFRSCategoryFVTPL, enums.IFRSCategoryFVOCI:
			if position.MarketPriceTWD.IsPositive() {
				fvBasis := position.MarketPriceTWD.Mul(p.Quantity)
				ct.State.AccumulatedUnrealizedTWD = fvBasis.Sub(originalCostBasis)
				ct.State.CostBasis = fvBasis
			}
		}

	case enums.CostMethodFIFO:
		result, err := e.calcFIFOCostBasis(ctx, inv.InvestmentId, p)
		if err != nil {
			return fmt.Errorf("calc fifo cost fail: %w", err)
		}
		ct.State.LotDisposals = result.Disposals
		originalCostBasis = result.OriginalCost
		// FVTPL/FVOCI：以批次 FV 沖銷（unrealized_unit_twd > 0 才生效）
		switch inv.IFRSCategory.Val() {
		case enums.IFRSCategoryFVTPL, enums.IFRSCategoryFVOCI:
			ct.State.CostBasis = result.FVCost
			ct.State.AccumulatedUnrealizedTWD = result.AccumulatedUnrealized
		default:
			ct.State.CostBasis = result.OriginalCost
		}
	}

	ct.State.RealizedGain = sellAmount.Sub(originalCostBasis)
	ct.State.NetProceeds = sellAmount.Sub(p.Fee).Sub(p.Tax)

	ct.State.Movement = projection.InvestmentMovement{
		InvestmentId: inv.InvestmentId,
		MovementType: enums.MovementTypeSell.Enum(),
		MovementDate: p.Date,
		Quantity:     p.Quantity.Neg(),
		UnitPrice:    p.UnitPrice,
		UnitPriceTWD: p.UnitPrice.Mul(p.ExchangeRate),
		ExchangeRate: p.ExchangeRate,
		Fee:          p.Fee,
		Tax:          p.Tax,
		RealizedGain: decimal.NewNullDecimal(ct.State.RealizedGain),
		CostBasis:    decimal.NewNullDecimal(ct.State.CostBasis),
	}

	entries, err := e.sellEntries(ctx, ct.State, &p)
	if err != nil {
		return err
	}
	ct.State.Transaction = payload.TransactionCreatedPayload{
		TransactionDate: p.Date,
		Description:     fmt.Sprintf("賣出 %s", inv.Name),
		Currency:        "TWD",
		Entries:         entries,
	}

	return nil
}

type fifoCalcResult struct {
	OriginalCost          decimal.Decimal
	FVCost                decimal.Decimal // 若批次曾評價則為 FV，否則同 OriginalCost
	AccumulatedUnrealized decimal.Decimal
	Disposals             []projection.InvestmentLotDisposals
}

func (e *eventInvestmentSoldProjector) calcFIFOCostBasis(
	ctx context.Context,
	investmentID int64,
	p payload.InvestmentSoldPayload,
) (fifoCalcResult, error) {
	lots, err := e.query.Investment.GetNotCloseLots(ctx, investmentID)
	if err != nil {
		return fifoCalcResult{}, err
	}

	remaining := p.Quantity
	originalCost := decimal.Zero
	fvCost := decimal.Zero
	accUnrealized := decimal.Zero
	var updates []projection.InvestmentLotDisposals

	for _, lot := range lots {
		if remaining.IsZero() {
			break
		}
		use := decimal.Min(remaining, lot.RemainingQty)
		remaining = remaining.Sub(use)

		lotOriginal := lot.UnitCost.Mul(use)
		lotUnrealized := lot.UnrealizedUnitTWD.Mul(use)
		lotFV := lotOriginal.Add(lotUnrealized) // = (unit_cost + unrealized_unit) × use

		originalCost = originalCost.Add(lotOriginal)
		fvCost = fvCost.Add(lotFV)
		accUnrealized = accUnrealized.Add(lotUnrealized)

		day, err := e.holdingDays(lot.AcquiredDate, p.Date)
		if err != nil {
			return fifoCalcResult{}, err
		}
		updates = append(updates, projection.InvestmentLotDisposals{
			LotId:             lot.LotId,
			LotUuid:           lot.LotUuid,
			Quantity:          use,
			CostBasis:         lotOriginal,
			SaleProceeds:      p.UnitPrice.Mul(use),
			CapitalGain:       p.UnitPrice.Mul(use).Sub(lotOriginal),
			DisposalDate:      p.Date,
			HoldingPeriodDays: day,
		})
	}

	if remaining.IsPositive() {
		return fifoCalcResult{}, fmt.Errorf("insufficient inventory: need %s more", remaining.StringFixed(4))
	}
	return fifoCalcResult{
		OriginalCost:          originalCost,
		FVCost:                fvCost,
		AccumulatedUnrealized: accUnrealized,
		Disposals:             updates,
	}, nil
}

func (e *eventInvestmentSoldProjector) holdingDays(buyStr, sellStr string) (int64, error) {
	layout := "2006-01-02"

	buy, err := time.Parse(layout, buyStr)
	if err != nil {
		return 0, err
	}

	sell, err := time.Parse(layout, sellStr)
	if err != nil {
		return 0, err
	}

	return int64(sell.Sub(buy).Hours() / 24), nil
}

func (e *eventInvestmentSoldProjector) sellEntries(
	ctx context.Context,
	ct *state.InvestmentSoldState,
	p *payload.InvestmentSoldPayload,
) ([]payload.TransactionEntryPayload, error) {
	config, err := e.query.Config.GetAssetTypeAccountConfig(ctx, ct.Investment.AssetType)
	if err != nil {
		return nil, fmt.Errorf("get asset type config: %w", err)
	}
	entries := []payload.TransactionEntryPayload{
		{AccountId: ct.Ledger.AccountId, LedgerId: &p.LedgerId, Debit: ct.NetProceeds, Credit: decimal.Zero},
		{AccountId: ct.Investment.AccountId, Debit: decimal.Zero, Credit: ct.CostBasis},
	}
	if p.Fee.IsPositive() {
		entries = append(entries, payload.TransactionEntryPayload{AccountId: config.FeeAccountID, Debit: p.Fee, Credit: decimal.Zero})
	}
	if p.Tax.IsPositive() {
		entries = append(entries, payload.TransactionEntryPayload{AccountId: config.TaxAccountID, Debit: p.Tax, Credit: decimal.Zero})
	}
	if ct.RealizedGain.IsPositive() {
		entries = append(entries, payload.TransactionEntryPayload{AccountId: config.RealizedGainAccountID, Debit: decimal.Zero, Credit: ct.RealizedGain})
	} else if ct.RealizedGain.IsNegative() {
		entries = append(entries, payload.TransactionEntryPayload{AccountId: config.RealizedLossAccountID, Debit: ct.RealizedGain.Abs(), Credit: decimal.Zero})
	}
	// 沖回累積未實現損益
	if !ct.AccumulatedUnrealizedTWD.IsZero() {
		absAcc := ct.AccumulatedUnrealizedTWD.Abs()
		switch ct.Investment.IFRSCategory.Val() {
		case enums.IFRSCategoryFVTPL:
			if ct.AccumulatedUnrealizedTWD.IsPositive() {
				entries = append(entries, payload.TransactionEntryPayload{AccountId: config.UnrealizedGainAccountID, Debit: absAcc, Credit: decimal.Zero})
			} else {
				entries = append(entries, payload.TransactionEntryPayload{AccountId: config.UnrealizedLossAccountID, Debit: decimal.Zero, Credit: absAcc})
			}
		case enums.IFRSCategoryFVOCI:
			if config.OciAccountID == nil {
				return nil, fmt.Errorf("oci_account_id not configured for asset type %s", ct.Investment.AssetType.String())
			}
			if ct.AccumulatedUnrealizedTWD.IsPositive() {
				entries = append(entries, payload.TransactionEntryPayload{AccountId: *config.OciAccountID, Debit: absAcc, Credit: decimal.Zero})
			} else {
				entries = append(entries, payload.TransactionEntryPayload{AccountId: *config.OciAccountID, Debit: decimal.Zero, Credit: absAcc})
			}
		}
	}
	return entries, nil
}

func NewEventInvestmentSoldPipeline(query *query.Repo) *pipelines.TypedPipeline[state.InvestmentSoldState, payload.InvestmentSoldPayload] {
	return pipelines.NewType[state.InvestmentSoldState, payload.InvestmentSoldPayload](&eventInvestmentSoldProjector{query}, func() *state.InvestmentSoldState {
		return &state.InvestmentSoldState{}
	})
}

// ------------------------------
// EventStockSplit
// ------------------------------

type eventStockSplitProjector struct {
	query *query.Repo
}

func (e eventStockSplitProjector) Project(ctx context.Context, ct *pipelines.Context[state.StockSplitState, payload.StockSplitPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload

	inv, err := e.query.Investment.GetByCreationEventUuid(ctx, p.InvestmentUUID)
	if err != nil {
		return err
	}

	if inv == nil {
		return fmt.Errorf("investment %s not found", p.InvestmentUUID)
	}
	ct.State.Investment = *inv
	ct.State.Movement = projection.InvestmentMovement{
		InvestmentId: inv.InvestmentId,
		MovementType: enums.MovementTypeSplit.Enum(),
		MovementDate: p.Date,
		SplitRatio:   decimal.NewNullDecimal(p.Ratio),
	}

	return nil
}

func NewEventStockSplitPipeline(query *query.Repo) *pipelines.TypedPipeline[state.StockSplitState, payload.StockSplitPayload] {
	return pipelines.NewType[state.StockSplitState, payload.StockSplitPayload](&eventStockSplitProjector{query}, func() *state.StockSplitState {
		return &state.StockSplitState{}
	})
}

// ------------------------------
// EventDividendReceived
// ------------------------------

type eventDividendReceivedProjector struct {
	query *query.Repo
}

func (e eventDividendReceivedProjector) Project(ctx context.Context, ct *pipelines.Context[state.DividendReceivedState, payload.DividendReceivedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload

	inv, err := e.query.Investment.GetByCreationEventUuid(ctx, p.InvestmentUUID)
	if err != nil {
		return err
	}

	if inv == nil {
		return fmt.Errorf("investment %s not found", p.InvestmentUUID)
	}

	ledger, err := e.query.Account.GetLedger(ctx, p.LedgerId)
	if err != nil {
		return err
	}

	ct.State.Investment = *inv
	ct.State.Movement = projection.InvestmentMovement{
		InvestmentId: inv.InvestmentId,
		MovementType: enums.MovementTypeDividend.Enum(),
		MovementDate: p.Date,
		ExchangeRate: p.ExchangeRate,
		SplitRatio:   decimal.NewNullDecimal(p.Ratio),
	}
	amountTWD := decimal.Zero
	netAmountTWD := decimal.Zero
	if p.Amount.GreaterThan(decimal.Zero) {

		if err := validPeriodMonthlyStatus(ctx, p.Date, e.query, enums.PeriodTypeStatusOpen.Enum()); err != nil {
			return err
		}

		amountTWD = p.Amount.Mul(p.ExchangeRate)
		netAmountTWD = amountTWD.Sub(p.WithholdingTax)
		ct.State.Movement.GrossAmount = decimal.NewNullDecimal(amountTWD)
		entries := []payload.TransactionEntryPayload{{AccountId: "4210", Debit: decimal.Zero, Credit: amountTWD}}
		if p.WithholdingTax.GreaterThan(decimal.Zero) {
			ct.State.Movement.WithholdingTax = decimal.NewNullDecimal(p.WithholdingTax)
			entries = append(entries, payload.TransactionEntryPayload{AccountId: "5920", Debit: p.WithholdingTax, Credit: decimal.Zero})
		}
		ct.State.Movement.NetAmount = decimal.NewNullDecimal(netAmountTWD)
		entries = append(entries, payload.TransactionEntryPayload{AccountId: ledger.AccountId, LedgerId: &p.LedgerId, Debit: netAmountTWD, Credit: decimal.Zero})
		ct.State.Transaction = &payload.TransactionCreatedPayload{
			TransactionDate: p.Date,
			Description:     fmt.Sprintf("%s 配息", inv.Name),
			TotalAmount:     amountTWD,
			Currency:        "TWD",
			Entries:         entries,
		}
	}

	return nil
}

func NewEventDividendReceivedPipeline(query *query.Repo) *pipelines.TypedPipeline[state.DividendReceivedState, payload.DividendReceivedPayload] {
	return pipelines.NewType[state.DividendReceivedState, payload.DividendReceivedPayload](&eventDividendReceivedProjector{query}, func() *state.DividendReceivedState {
		return &state.DividendReceivedState{}
	})
}

// ------------------------------
// EventUnrealizedMarked
// ------------------------------

type eventUnrealizedMarkedProjector struct {
	query *query.Repo
}

func (e *eventUnrealizedMarkedProjector) Project(ctx context.Context, ct *pipelines.Context[state.UnrealizedMarkedState, payload.UnrealizedMarkedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}
	p := ct.Payload

	inv, err := e.query.Investment.GetByCreationEventUuid(ctx, p.InvestmentUUID)
	if err != nil {
		return err
	}
	if inv == nil {
		return fmt.Errorf("investment %s not found", p.InvestmentUUID)
	}
	ct.State.Investment = *inv

	newPriceTWD := p.MarketPrice.Mul(p.ExchangeRate)

	var totalQty, carryingTWD decimal.Decimal
	var lotUnrealizedUpdates []state.LotUnrealizedUpdate

	switch inv.CostMethod.Val() {
	case enums.CostMethodAvg:
		pos, err := e.query.Investment.GetPosition(ctx, inv.InvestmentId)
		if err != nil {
			return err
		}
		if pos == nil {
			return fmt.Errorf("investment %s has no position", p.InvestmentUUID)
		}
		totalQty = pos.TotalQuantity
		if pos.MarketPriceTWD.IsPositive() {
			carryingTWD = pos.MarketPriceTWD.Mul(totalQty)
		} else {
			carryingTWD = pos.TotalCost
		}

	case enums.CostMethodFIFO:
		lots, err := e.query.Investment.GetNotCloseLots(ctx, inv.InvestmentId)
		if err != nil {
			return err
		}
		for _, lot := range lots {
			totalQty = totalQty.Add(lot.RemainingQty)
			// 前期批次帳面值 = (原始成本 + 前期每單位未實現) × 剩餘數量
			prevCarrying := lot.UnitCost.Add(lot.UnrealizedUnitTWD).Mul(lot.RemainingQty)
			carryingTWD = carryingTWD.Add(prevCarrying)
			lotUnrealizedUpdates = append(lotUnrealizedUpdates, state.LotUnrealizedUpdate{
				LotId:             lot.LotId,
				UnrealizedUnitTWD: newPriceTWD.Sub(lot.UnitCost),
			})
		}
	}

	if totalQty.IsZero() {
		return fmt.Errorf("investment %s has zero position, cannot mark fair value", p.InvestmentUUID)
	}

	newValueTWD := newPriceTWD.Mul(totalQty)
	adjustmentTWD := newValueTWD.Sub(carryingTWD)

	ct.State.TotalQty = totalQty
	ct.State.CarryingAmountTWD = carryingTWD
	ct.State.NewMarketValueTWD = newValueTWD
	ct.State.AdjustmentTWD = adjustmentTWD
	ct.State.NewMarketPriceTWD = newPriceTWD
	ct.State.LotUnrealizedUpdates = lotUnrealizedUpdates

	ct.State.Movement = projection.InvestmentMovement{
		InvestmentId: inv.InvestmentId,
		MovementType: enums.MovementTypeMark.Enum(),
		MovementDate: p.Date,
		Quantity:     decimal.Zero,
		UnitPrice:    p.MarketPrice,
		UnitPriceTWD: newPriceTWD,
		ExchangeRate: p.ExchangeRate,
	}

	// 產生分錄（調整額為 0 時不產生交易）
	if !adjustmentTWD.IsZero() {
		entries, err := e.fvEntries(ctx, inv, adjustmentTWD)
		if err != nil {
			return err
		}
		desc := fmt.Sprintf("%s 公允價值評估", inv.Name)
		ct.State.Transaction = &payload.TransactionCreatedPayload{
			TransactionDate: p.Date,
			Description:     desc,
			Currency:        "TWD",
			Entries:         entries,
		}
	}

	return nil
}

func (e *eventUnrealizedMarkedProjector) fvEntries(
	ctx context.Context,
	inv *projection.Investment,
	adjustment decimal.Decimal,
) ([]payload.TransactionEntryPayload, error) {
	config, err := e.query.Config.GetAssetTypeAccountConfig(ctx, inv.AssetType)
	if err != nil {
		return nil, fmt.Errorf("get asset type config: %w", err)
	}
	absAdj := adjustment.Abs()

	switch inv.IFRSCategory.Val() {
	case enums.IFRSCategoryFVTPL:
		if adjustment.IsPositive() {
			return []payload.TransactionEntryPayload{
				{AccountId: inv.AccountId, Debit: absAdj, Credit: decimal.Zero},
				{AccountId: config.UnrealizedGainAccountID, Debit: decimal.Zero, Credit: absAdj},
			}, nil
		}
		return []payload.TransactionEntryPayload{
			{AccountId: config.UnrealizedLossAccountID, Debit: absAdj, Credit: decimal.Zero},
			{AccountId: inv.AccountId, Debit: decimal.Zero, Credit: absAdj},
		}, nil

	case enums.IFRSCategoryFVOCI:
		if config.OciAccountID == nil {
			return nil, fmt.Errorf("oci_account_id not configured for asset type %s", inv.AssetType.String())
		}
		// FVOCI goes to equity (OCI), not through P&L.
		if adjustment.IsPositive() {
			return []payload.TransactionEntryPayload{
				{AccountId: inv.AccountId, Debit: absAdj, Credit: decimal.Zero},
				{AccountId: *config.OciAccountID, Debit: decimal.Zero, Credit: absAdj},
			}, nil
		}
		return []payload.TransactionEntryPayload{
			{AccountId: *config.OciAccountID, Debit: absAdj, Credit: decimal.Zero},
			{AccountId: inv.AccountId, Debit: decimal.Zero, Credit: absAdj},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported ifrs_category %s for fair value marking", inv.IFRSCategory.String())
	}
}

func NewEventUnrealizedMarkedPipeline(query *query.Repo) *pipelines.TypedPipeline[state.UnrealizedMarkedState, payload.UnrealizedMarkedPayload] {
	return pipelines.NewType[state.UnrealizedMarkedState, payload.UnrealizedMarkedPayload](&eventUnrealizedMarkedProjector{query}, func() *state.UnrealizedMarkedState {
		return &state.UnrealizedMarkedState{}
	})
}
