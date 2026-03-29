package calculator

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/db/report"
	"akatengu/internal/model/enums"
	"akatengu/internal/model/payload"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

func (e *eventCalculator) calEventPeriodMonthClosed(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p payload.PeriodMonthClosedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("malformed eventCalculator: %w", err)
	}

	existing, err := e.query.Period.GetByID(ctx, p.ClosingId)
	if err != nil {
		return nil, fmt.Errorf("get period: %w", err)
	}

	// 3. 生成期末資產負債表快照
	snapshot, err := e.buildSnapshot(ctx, existing.PeriodEnd)
	if err != nil {
		return nil, fmt.Errorf("build snapshot: %w", err)
	}
	p.Snapshot, err = e.marshalSnapshot(*snapshot)
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot: %w", err)
	}

	if existing.Status.String() == enums.PeriodTypeStatusOpen.String() {
		periodStart, periodEnd, err := e.nextPeriodMonthRange(existing.PeriodEnd)
		if err != nil {
			return nil, fmt.Errorf("next period start: %w", err)
		}

		p.Next = &projection.PeriodClosing{
			PeriodType:  enums.PeriodMonthly,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			Status:      enums.PeriodTypeStatusOpen,
		}
	}

	result, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("marshal period created: %w", err)
	}
	return result, nil
}

func (e *eventCalculator) calEventPeriodAnnualClosed(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p payload.PeriodAnnualClosedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("malformed eventCalculator: %w", err)
	}

	existing, err := e.query.Period.GetByID(ctx, p.ClosingId)
	if err != nil {
		return nil, fmt.Errorf("get period: %w", err)
	}
	t, _ := time.Parse("2006-01-02", existing.PeriodStart)
	year := t.Year()

	periodStart, periodEnd := e.monthRange(year)

	// 3. 計算本年損益
	is, err := e.query.Report.GetIncomeStatement(ctx, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("income statement: %w", err)
	}
	netIncome := is.NetIncome

	// 4. Period Entry：收入、支出結轉 3200 本期損益
	p.ClosedTransaction, err = e.appendClosingEntry(ctx, year, netIncome, is)
	if err != nil {
		return nil, fmt.Errorf("closing entry: %w", err)
	}

	// 5. 期初結轉：3200 → 3100
	p.OpenedTransaction, err = e.appendOpeningEntry(ctx, year, netIncome)
	if err != nil {
		return nil, fmt.Errorf("opening entry: %w", err)
	}

	// 6. 生成期末快照
	snapshot, err := e.buildSnapshot(ctx, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("build snapshot: %w", err)
	}
	p.Snapshot, err = e.marshalSnapshot(*snapshot)
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot: %w", err)
	}

	if existing.Status.String() == enums.PeriodTypeStatusOpen.String() {
		periodStart, periodEnd = e.monthRange(year + 1)
		p.Next = &projection.PeriodClosing{
			PeriodType:  enums.PeriodAnnual,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			Status:      enums.PeriodTypeStatusOpen,
		}
	}

	result, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("marshal period created: %w", err)
	}
	return result, nil
}

func (e *eventCalculator) calEventPeriodAnnualReopened(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p payload.PeriodAnnualReopenedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("malformed eventCalculator: %w", err)
	}

	// 1. 拿到年結紀錄
	closing, err := e.query.Period.GetByID(ctx, p.ClosingId)
	if closing.PeriodType.String() != enums.PeriodAnnual.String() {
		return nil, fmt.Errorf("closing %d is not an annual closing", p.ClosingId)
	}
	if closing.Status.String() != enums.PeriodTypeStatusClosed.String() {
		return nil, fmt.Errorf("closing %d is not closed", p.ClosingId)
	}

	year, _ := strconv.Atoi(closing.PeriodStart[:4])

	// 2. 沖銷 Period Entry
	if closing.ClosingTxnID != nil {
		entries, err := e.query.Entry.GetEntries(ctx, *closing.ClosingTxnID)
		if err != nil {
			return nil, fmt.Errorf("get closing entries: %w", err)
		}
		if len(entries) == 0 {
			return nil, fmt.Errorf("closing entry %d has no journal entries", *closing.ClosingTxnID)
		}

		entriesPayload, totalAmount := e.reverseEntries(entries) // 從原始 Period Entry 反向
		p.ReverseClosedTxn = payload.TransactionCreatedPayload{
			TransactionDate: closing.PeriodEnd,
			Description:     fmt.Sprintf("%d 年度結帳沖銷", year),
			TotalAmount:     totalAmount,
			Entries:         entriesPayload,
			RefTxnId:        closing.ClosingTxnID,
		}
	}

	// 3. 沖銷 Opening Entry（下一年的 1/1）
	nextYear := year + 1
	if closing.OpeningTxnID != nil {
		entries, err := e.query.Entry.GetEntries(ctx, *closing.OpeningTxnID)
		if err != nil {
			return nil, fmt.Errorf("get opening entries: %w", err)
		}
		if len(entries) == 0 {
			return nil, fmt.Errorf("closing entry %d has no journal entries", *closing.ClosingTxnID)
		}

		entriesPayload, totalAmount := e.reverseEntries(entries) // 從原始 Period Entry 反向
		p.ReverseOpenedTxn = payload.TransactionCreatedPayload{
			TransactionDate: fmt.Sprintf("%d-01-01", nextYear),
			Description:     fmt.Sprintf("%d 期初結轉沖銷", year),
			TotalAmount:     totalAmount,
			Entries:         entriesPayload,
			RefTxnId:        closing.OpeningTxnID,
		}
	}

	result, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("marshal period created: %w", err)
	}
	return result, nil
}

func (e *eventCalculator) reverseEntries(entries []projection.Entry) ([]payload.TransactionEntryPayload, decimal.Decimal) {
	var totalDebit, totalCredit decimal.Decimal
	reversed := make([]payload.TransactionEntryPayload, len(entries))
	for i, entry := range entries {
		totalDebit = totalDebit.Add(entry.Debit)
		totalCredit = totalCredit.Add(entry.Credit)
		reversed[i] = payload.TransactionEntryPayload{
			AccountId: entry.AccountId,
			Debit:     entry.Credit, // 借貸對調
			Credit:    entry.Debit,
		}
	}
	// 總金額只丟單邊就好，因為借貸是一樣的
	return reversed, totalDebit
}

// ─────────────────────────────────────────
// Period Entry（年結用）
// 收入貸方沖銷 → 借方
// 支出借方沖銷 → 貸方
// 差額轉入 3200 本期損益
// ─────────────────────────────────────────

func (e *eventCalculator) appendClosingEntry(
	ctx context.Context,
	year int,
	netIncome decimal.Decimal,
	is *report.IncomeStatement,
) (payload.TransactionCreatedPayload, error) {
	result := payload.TransactionCreatedPayload{}
	result.Entries = []payload.TransactionEntryPayload{}

	// 沖銷所有收入科目（收入正常貸方 → 借方沖銷）
	for _, row := range is.Income {
		if row.Amount.IsZero() {
			continue
		}
		result.Entries = append(result.Entries, payload.TransactionEntryPayload{
			AccountId: row.AccountID,
			Debit:     row.Amount,
			Credit:    decimal.Zero,
		})
	}

	// 沖銷所有支出科目（支出正常借方 → 貸方沖銷）
	for _, row := range is.Expenses {
		if row.Amount.IsZero() {
			continue
		}
		result.Entries = append(result.Entries, payload.TransactionEntryPayload{
			AccountId: row.AccountID,
			Debit:     decimal.Zero,
			Credit:    row.Amount,
		})
	}

	// 差額轉入 3200 本期損益
	// 淨利 → 3200 貸方；淨損 → 3200 借方
	if netIncome.IsPositive() {
		result.Entries = append(result.Entries, payload.TransactionEntryPayload{
			AccountId: "3200",
			Debit:     decimal.Zero,
			Credit:    netIncome,
		})
	} else {
		result.Entries = append(result.Entries, payload.TransactionEntryPayload{
			AccountId: "3200",
			Debit:     netIncome.Abs(),
			Credit:    decimal.Zero,
		})
	}

	result.TransactionDate = fmt.Sprintf("%d-12-31", year)
	result.Description = fmt.Sprintf("%d 年度結帳", year)
	result.TotalAmount = netIncome.Abs()
	result.Currency = "TWD"
	return result, nil
}

// ─────────────────────────────────────────
// Opening Entry（年結後期初結轉）
// 3200 本期損益 → 3100 期初淨資產
// ─────────────────────────────────────────

func (e *eventCalculator) appendOpeningEntry(ctx context.Context, year int, netIncome decimal.Decimal) (payload.TransactionCreatedPayload, error) {
	// 3200 結轉到 3100
	// 本期有淨利：3200 借方，3100 貸方
	// 本期有淨損：3200 貸方，3100 借方
	result := payload.TransactionCreatedPayload{}
	result.Entries = []payload.TransactionEntryPayload{}
	if netIncome.IsPositive() {
		result.Entries = []payload.TransactionEntryPayload{
			{AccountId: "3200", Debit: netIncome, Credit: decimal.Zero},
			{AccountId: "3100", Debit: decimal.Zero, Credit: netIncome},
		}
	} else {
		abs := netIncome.Abs()
		result.Entries = []payload.TransactionEntryPayload{
			{AccountId: "3200", Debit: decimal.Zero, Credit: abs},
			{AccountId: "3100", Debit: abs, Credit: decimal.Zero},
		}
	}
	result.TransactionDate = fmt.Sprintf("%d-01-01", year+1)
	result.Description = fmt.Sprintf("%d 期初餘額結轉", year+1)
	result.TotalAmount = netIncome.Abs()
	result.Currency = "TWD"
	return result, nil
}

func (e *eventCalculator) buildSnapshot(ctx context.Context, date string) (*projection.SnapshotData, error) {
	bs, err := e.query.Report.GetBalanceSheet(ctx, date)
	if err != nil {
		return nil, err
	}

	snap := &projection.SnapshotData{
		TotalAssets:      bs.TotalAssets,
		TotalLiabilities: bs.TotalLiabilities,
		NetWorth:         bs.NetWorth,
	}
	for _, r := range bs.Assets {
		snap.Assets = append(snap.Assets, projection.BalanceSheetEntry{
			AccountId: r.AccountId,
			Name:      r.Name,
			Balance:   r.Balance,
		})
	}
	for _, r := range bs.Liabilities {
		snap.Liabilities = append(snap.Liabilities, projection.BalanceSheetEntry{
			AccountId: r.AccountId,
			Name:      r.Name,
			Balance:   r.Balance,
		})
	}
	for _, r := range bs.Equity {
		snap.Equity = append(snap.Equity, projection.BalanceSheetEntry{
			AccountId: r.AccountId,
			Name:      r.Name,
			Balance:   r.Balance,
		})
	}
	return snap, nil
}

func (e *eventCalculator) marshalSnapshot(data projection.SnapshotData) (string, error) {
	data.GeneratedAt = time.Now().Format(time.RFC3339)
	b, err := json.Marshal(data)
	return string(b), err
}

func (e *eventCalculator) marshalTransaction(data payload.TransactionCreatedPayload) (string, error) {
	b, err := json.Marshal(data)
	return string(b), err
}

func (e *eventCalculator) monthRange(year int) (startDate, endDate string) {
	// 當月第一天
	first := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

	// 當月最後一天（下個月第一天 -1 天）
	last := first.AddDate(1, 0, -1)

	return first.Format("2006-01-02"), last.Format("2006-01-02")
}

func (e *eventCalculator) nextPeriodMonthRange(date string) (startDate, endDate string, err error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", "", fmt.Errorf("invalid date %q: %w", date, err)
	}

	// 下個月第一天
	first := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, time.UTC)

	// 下個月最後一天
	last := first.AddDate(0, 1, -1)

	return first.Format("2006-01-02"), last.Format("2006-01-02"), nil
}
