package factory

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/sys_codes"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/db/report"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/pipelines"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// ------------------------------
// EventPeriodMonthStarted
// ------------------------------
type eventPeriodMonthStartedProjector struct {
	query *query.Repo
}

func (e eventPeriodMonthStartedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.PeriodMonthStartedPayload]) error {
	return ct.Payload.Validate()
}

func NewEventPeriodMonthStartedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.PeriodMonthStartedPayload] {
	return pipelines.NewTypeWithNoState[payload.PeriodMonthStartedPayload](&eventPeriodMonthStartedProjector{query})
}

// ------------------------------
// EventPeriodMonthClosed
// ------------------------------
type eventPeriodMonthClosedProjector struct {
	query *query.Repo
}

func (e eventPeriodMonthClosedProjector) Project(ctx context.Context, ct *pipelines.Context[state.PeriodMonthClosedState, payload.PeriodMonthClosedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload
	s := ct.State

	// 1. 檢查是否已結帳
	existing, err := e.query.Period.GetByID(ctx, p.ClosingId)
	if err != nil {
		fmt.Errorf("get period: %w", err)
	}

	if existing == nil {
		return fmt.Errorf("closing_id %d is not exists", p.ClosingId)
	}

	if existing.Status.Is(enums.PeriodTypeStatusClosed) {
		return fmt.Errorf("period %s~%s is already closed", existing.PeriodStart, existing.PeriodEnd)
	}

	// 2. 檢查有無未解決的對帳差異
	if err := e.query.Period.AssertNoUnresolvedAdjustments(ctx, existing.PeriodStart, existing.PeriodEnd); err != nil {
		return err
	}

	// 3. 處理 state
	existing, err = e.query.Period.GetByID(ctx, p.ClosingId)
	if err != nil {
		return fmt.Errorf("get period: %w", err)
	}

	// 3. 生成期末資產負債表快照
	snapshot, err := buildSnapshot(ctx, existing.PeriodEnd, e.query)
	if err != nil {
		return fmt.Errorf("build snapshot: %w", err)
	}
	s.Snapshot, err = marshalSnapshot(*snapshot)
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	if existing.Status.Is(enums.PeriodTypeStatusOpen) {
		periodStart, periodEnd, err := nextPeriodMonthRange(existing.PeriodEnd)
		if err != nil {
			return fmt.Errorf("next period start: %w", err)
		}

		s.Next = &projection.PeriodClosing{
			PeriodType:  enums.PeriodMonthly.Enum(),
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			Status:      enums.PeriodTypeStatusOpen.Enum(),
		}
	}
	return nil
}

func NewEventPeriodMonthClosedPipeline(query *query.Repo) *pipelines.TypedPipeline[state.PeriodMonthClosedState, payload.PeriodMonthClosedPayload] {
	return pipelines.NewType[state.PeriodMonthClosedState, payload.PeriodMonthClosedPayload](&eventPeriodMonthClosedProjector{query}, func() *state.PeriodMonthClosedState {
		return &state.PeriodMonthClosedState{}
	})
}

// ------------------------------
// EventPeriodMonthReopened
// ------------------------------
type eventPeriodMonthReopenedProjector struct {
	query *query.Repo
}

func (e eventPeriodMonthReopenedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.PeriodMonthReopenedPayload]) error {
	return ct.Payload.Validate()
}

func NewEventPeriodMonthReopenedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.PeriodMonthReopenedPayload] {
	return pipelines.NewTypeWithNoState[payload.PeriodMonthReopenedPayload](&eventPeriodMonthReopenedProjector{query})
}

// ------------------------------
// EventPeriodAnnualStarted
// ------------------------------
type eventPeriodAnnualStartedProjector struct {
	query *query.Repo
}

func (e eventPeriodAnnualStartedProjector) Project(ctx context.Context, ct *pipelines.Context[pipelines.NoState, payload.PeriodAnnualStartedPayload]) error {
	return ct.Payload.Validate()
}

func NewEventPeriodAnnualStartedPipeline(query *query.Repo) *pipelines.TypedPipeline[pipelines.NoState, payload.PeriodAnnualStartedPayload] {
	return pipelines.NewTypeWithNoState[payload.PeriodAnnualStartedPayload](&eventPeriodAnnualStartedProjector{query})
}

// ------------------------------
// EventPeriodAnnualClosed
// ------------------------------
type eventPeriodAnnualClosedProjector struct {
	query *query.Repo
}

func (e eventPeriodAnnualClosedProjector) Project(ctx context.Context, ct *pipelines.Context[state.PeriodAnnualClosedState, payload.PeriodAnnualClosedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload
	s := ct.State

	existing, err := e.query.Period.GetByID(ctx, p.ClosingId)
	if err != nil {
		return fmt.Errorf("get period: %w", err)
	}
	t, _ := time.Parse("2006-01-02", existing.PeriodStart)
	year := t.Year()
	if existing != nil && existing.Status.Is(enums.PeriodTypeStatusClosed) {
		return fmt.Errorf("year %d is already closed", year)
	}

	// 2. 確認 12 個月都已月結
	if err := assertAllMonthsClosed(ctx, year, e.query); err != nil {
		return err
	}

	periodStart, periodEnd := monthRange(year)

	// 3. 計算本年損益
	is, err := e.query.Report.GetIncomeStatement(ctx, periodStart, periodEnd)
	if err != nil {
		return fmt.Errorf("income statement: %w", err)
	}
	netIncome := is.NetIncome

	closing, err := getSysAccountCode(ctx, e.query.Sys, sys_codes.SysAccountEquityCloseNetIncome.Enum())
	if err != nil {
		return err
	}

	// 4. Period Entry：收入、支出結轉 3102-01 本期損益
	s.ClosedTxn, err = appendClosingEntry(ctx, year, netIncome, is, closing)
	if err != nil {
		return fmt.Errorf("closing entry: %w", err)
	}

	opening, err := getSysAccountCode(ctx, e.query.Sys, sys_codes.SysAccountEquityEquityOpening.Enum())
	if err != nil {
		return err
	}

	// 5. 期初結轉：3102-01 → 3101-01
	s.OpenedTxn, err = appendOpeningEntry(ctx, year, netIncome, closing, opening)
	if err != nil {
		return fmt.Errorf("opening entry: %w", err)
	}

	// 6. 生成期末快照
	snapshot, err := buildSnapshot(ctx, periodEnd, e.query)
	if err != nil {
		return fmt.Errorf("build snapshot: %w", err)
	}
	s.Snapshot, err = marshalSnapshot(*snapshot)
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	if existing.Status.Is(enums.PeriodTypeStatusOpen) {
		periodStart, periodEnd = monthRange(year + 1)
		s.Next = &projection.PeriodClosing{
			PeriodType:  enums.PeriodAnnual.Enum(),
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			Status:      enums.PeriodTypeStatusOpen.Enum(),
		}
	}

	return nil
}

func NewEventPeriodAnnualClosedPipeline(query *query.Repo) *pipelines.TypedPipeline[state.PeriodAnnualClosedState, payload.PeriodAnnualClosedPayload] {
	return pipelines.NewType[state.PeriodAnnualClosedState, payload.PeriodAnnualClosedPayload](&eventPeriodAnnualClosedProjector{query}, func() *state.PeriodAnnualClosedState {
		return &state.PeriodAnnualClosedState{}
	})
}

// ------------------------------
// EventPeriodAnnualReopened
// ------------------------------
type eventPeriodAnnualReopenedProjector struct {
	query *query.Repo
}

func (e eventPeriodAnnualReopenedProjector) Project(ctx context.Context, ct *pipelines.Context[state.PeriodAnnualReopenedState, payload.PeriodAnnualReopenedPayload]) error {
	if err := ct.Payload.Validate(); err != nil {
		return err
	}

	p := ct.Payload
	s := ct.State

	// 1. 檢查是否已年結
	existing, err := e.query.Period.GetByID(ctx, p.ClosingId)
	if err != nil {
		return fmt.Errorf("get period: %w", err)
	}
	t, _ := time.Parse("2006-01-02", existing.PeriodStart)
	year := t.Year()
	if existing != nil && !existing.Status.Is(enums.PeriodTypeStatusClosed) {
		return fmt.Errorf("year %d is not closed", year)
	}

	// 2. 沖銷 Period Entry
	if existing.ClosingTxnID != nil {
		entries, err := e.query.Entry.GetEntries(ctx, *existing.ClosingTxnID)
		if err != nil {
			return fmt.Errorf("get closing entries: %w", err)
		}
		if len(entries) == 0 {
			return fmt.Errorf("closing entry %d has no journal entries", *existing.ClosingTxnID)
		}

		entriesPayload, totalAmount := reverseEntries(entries) // 從原始 Period Entry 反向
		s.ReverseClosedTxn = payload.TransactionCreatedPayload{
			TransactionDate: existing.PeriodEnd,
			Description:     fmt.Sprintf("%d 年度結帳沖銷", year),
			TotalAmount:     totalAmount,
			Entries:         entriesPayload,
			RefTxnId:        existing.ClosingTxnID,
		}
	}

	// 3. 沖銷 Opening Entry（下一年的 1/1）
	nextYear := year + 1
	if existing.OpeningTxnID != nil {
		entries, err := e.query.Entry.GetEntries(ctx, *existing.OpeningTxnID)
		if err != nil {
			return fmt.Errorf("get opening entries: %w", err)
		}
		if len(entries) == 0 {
			return fmt.Errorf("closing entry %d has no journal entries", *existing.ClosingTxnID)
		}

		entriesPayload, totalAmount := reverseEntries(entries) // 從原始 Period Entry 反向
		s.ReverseOpenedTxn = payload.TransactionCreatedPayload{
			TransactionDate: fmt.Sprintf("%d-01-01", nextYear),
			Description:     fmt.Sprintf("%d 期初結轉沖銷", year),
			TotalAmount:     totalAmount,
			Entries:         entriesPayload,
			RefTxnId:        existing.OpeningTxnID,
		}
	}

	return nil
}

func NewEventPeriodAnnualReopenedPipeline(query *query.Repo) *pipelines.TypedPipeline[state.PeriodAnnualReopenedState, payload.PeriodAnnualReopenedPayload] {
	return pipelines.NewType[state.PeriodAnnualReopenedState, payload.PeriodAnnualReopenedPayload](&eventPeriodAnnualReopenedProjector{query}, func() *state.PeriodAnnualReopenedState {
		return &state.PeriodAnnualReopenedState{}
	})
}

// ------------------------------
// Helper
// ------------------------------
func reverseEntries(entries []projection.Entry) ([]payload.TransactionEntryPayload, decimal.Decimal) {
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

func appendClosingEntry(
	ctx context.Context,
	year int,
	netIncome decimal.Decimal,
	is *report.IncomeStatement,
	netIncomeAccountId string,
) (payload.TransactionCreatedPayload, error) {
	result := payload.TransactionCreatedPayload{}
	result.Entries = []payload.TransactionEntryPayload{}

	// 沖銷所有收入科目（收入正常貸方 → 借方沖銷）；只沖葉科目，摘要科目為匯總顯示用，不重複沖銷
	for _, row := range is.Income {
		if row.HasChild || row.Amount.IsZero() {
			continue
		}
		result.Entries = append(result.Entries, payload.TransactionEntryPayload{
			AccountId: row.AccountID,
			Debit:     row.Amount,
			Credit:    decimal.Zero,
		})
	}

	// 沖銷所有支出科目（支出正常借方 → 貸方沖銷）；只沖葉科目，同上
	for _, row := range is.Expenses {
		if row.HasChild || row.Amount.IsZero() {
			continue
		}
		result.Entries = append(result.Entries, payload.TransactionEntryPayload{
			AccountId: row.AccountID,
			Debit:     decimal.Zero,
			Credit:    row.Amount,
		})
	}

	// 差額轉入 3102-01 本期損益
	// 淨利 → 3102-01 貸方；淨損 → 3102-01 借方
	if netIncome.IsPositive() {
		result.Entries = append(result.Entries, payload.TransactionEntryPayload{
			AccountId: netIncomeAccountId,
			Debit:     decimal.Zero,
			Credit:    netIncome,
		})
	} else {
		result.Entries = append(result.Entries, payload.TransactionEntryPayload{
			AccountId: netIncomeAccountId,
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
// 3102-01 本期損益 → 3101-01 期初淨資產
// ─────────────────────────────────────────

func appendOpeningEntry(ctx context.Context, year int,
	netIncome decimal.Decimal,
	netIncomeAccountId string,
	openingAccountId string,
) (payload.TransactionCreatedPayload, error) {
	// 3102-01 結轉到 3101-01
	// 本期有淨利：3102-01 借方，3101-01 貸方
	// 本期有淨損：3102-01 貸方，3101-01 借方
	result := payload.TransactionCreatedPayload{}
	result.Entries = []payload.TransactionEntryPayload{}
	if netIncome.IsPositive() {
		result.Entries = []payload.TransactionEntryPayload{
			{AccountId: netIncomeAccountId, Debit: netIncome, Credit: decimal.Zero},
			{AccountId: openingAccountId, Debit: decimal.Zero, Credit: netIncome},
		}
	} else {
		abs := netIncome.Abs()
		result.Entries = []payload.TransactionEntryPayload{
			{AccountId: netIncomeAccountId, Debit: decimal.Zero, Credit: abs},
			{AccountId: openingAccountId, Debit: abs, Credit: decimal.Zero},
		}
	}
	result.TransactionDate = fmt.Sprintf("%d-01-01", year+1)
	result.Description = fmt.Sprintf("%d 期初餘額結轉", year+1)
	result.TotalAmount = netIncome.Abs()
	result.Currency = "TWD"
	return result, nil
}

func buildSnapshot(ctx context.Context, date string, query *query.Repo) (*projection.SnapshotData, error) {
	bs, err := query.Report.GetBalanceSheet(ctx, date)
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

func marshalSnapshot(data projection.SnapshotData) (string, error) {
	data.GeneratedAt = time.Now().Format(time.RFC3339)
	b, err := json.Marshal(data)
	return string(b), err
}

func marshalTransaction(data payload.TransactionCreatedPayload) (string, error) {
	b, err := json.Marshal(data)
	return string(b), err
}

// ─────────────────────────────────────────
// 日期計算
// ─────────────────────────────────────────

func periodStartDate(date string) (string, error) {
	datetime, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", err
	}
	first := time.Date(datetime.Year(), datetime.Month(), 1, 0, 0, 0, 0, time.UTC)
	return first.Format("2006-01-02"), nil
}

func monthRange(year int) (startDate, endDate string) {
	// 當月第一天
	first := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

	// 當月最後一天（下個月第一天 -1 天）
	last := first.AddDate(1, 0, -1)

	return first.Format("2006-01-02"), last.Format("2006-01-02")
}

func nextPeriodMonthRange(date string) (startDate, endDate string, err error) {
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

func assertAllMonthsClosed(ctx context.Context, year int, query *query.Repo) error {
	for m := time.January; m <= time.December; m++ {
		start, _ := monthRangeByYearMonth(year, m)
		c, err := query.Period.GetByPeriod(ctx, enums.PeriodMonthly.Enum(), start)
		if err != nil {
			return err
		}
		if c == nil || !c.Status.Is(enums.PeriodTypeStatusClosed) {
			return fmt.Errorf("month %d-%02d is not closed yet", year, m)
		}
	}
	return nil
}

func monthRangeByYearMonth(year int, month time.Month) (start, end string) {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	return first.Format("2006-01-02"), last.Format("2006-01-02")
}
