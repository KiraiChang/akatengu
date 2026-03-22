package services

import (
	"akatengu/internal/model/db"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/db/report"
	"akatengu/internal/model/enums"
	"akatengu/internal/model/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/repos/query"
	"strconv"

	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

// ─────────────────────────────────────────
// Interface
// ─────────────────────────────────────────

type ClosingService interface {
	// CloseMonth 月結
	CloseMonth(ctx context.Context, year int, month time.Month) (*db.PeriodClosing, error)

	// CloseYear 年結（包含 closing entry + 期初結轉）
	CloseYear(ctx context.Context, year int) (*db.PeriodClosing, error)

	// Reopen 重開已結帳期間（發現問題時）
	Reopen(ctx context.Context, closingID int64, reason string) error

	// AssertNotClosed 新增交易前檢查是否在已結帳期間
	AssertNotClosed(ctx context.Context, txnDate string) error

	// ReopenYear 再開帳
	ReopenYear(ctx context.Context, closingID int64, reason string) error
}

// ─────────────────────────────────────────
// Implementation
// ─────────────────────────────────────────

type closingService struct {
	db        *sqlx.DB
	closing   query.ClosingRepo
	reportSvc ReportService
	eventSvc  EventStoreService
}

func NewClosingService(
	db *sqlx.DB,
	closing query.ClosingRepo,
	reportSvc ReportService,
	eventSvc EventStoreService,
) ClosingService {
	return &closingService{
		db:        db,
		closing:   closing,
		reportSvc: reportSvc,
		eventSvc:  eventSvc,
	}
}

// ─────────────────────────────────────────
// CloseMonth
// ─────────────────────────────────────────

func (s *closingService) CloseMonth(ctx context.Context, year int, month time.Month) (*db.PeriodClosing, error) {
	periodStart, periodEnd := monthRange(year, month)

	// 1. 檢查是否已結帳
	existing, err := s.closing.GetByPeriod(ctx, enums.PeriodMonthly, periodStart)
	if err != nil {
		return nil, fmt.Errorf("get period: %w", err)
	}
	if existing != nil && existing.Status.String() == enums.ClosingStatusClosed.String() {
		return nil, fmt.Errorf("period %s~%s is already closed", periodStart, periodEnd)
	}

	// 2. 檢查有無未解決的對帳差異
	if err := s.assertNoUnresolvedAdjustments(ctx, periodStart, periodEnd); err != nil {
		return nil, err
	}

	// 3. 生成期末資產負債表快照
	snapshot, err := s.buildSnapshot(ctx, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("build snapshot: %w", err)
	}
	snapshotJSON, err := MarshalSnapshot(*snapshot)
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot: %w", err)
	}

	// 4. 寫入結帳紀錄

	now := time.Now().Format(time.RFC3339)
	// db 操作全部在 WithTx 裡，service 看不到 sqlx
	err = s.closing.WithTx(ctx, func(tx query.ClosingTxRepository) error {
		var closingID int64

		if existing == nil {
			closingID, err = tx.Insert(ctx, db.PeriodClosing{
				PeriodType:  enums.PeriodMonthly,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				Status:      enums.ClosingStatusClosed,
			})
			if err != nil {
				return fmt.Errorf("insert closing: %w", err)
			}
		} else {
			closingID = existing.ClosingId
			if err := tx.UpdateStatus(ctx, closingID, enums.ClosingStatusClosed, &now); err != nil {
				return fmt.Errorf("update status: %w", err)
			}
		}

		return tx.UpdateSnapshot(ctx, closingID, snapshotJSON)
	})
	if err != nil {
		return nil, err
	}

	return s.closing.GetByPeriod(ctx, enums.PeriodMonthly, periodStart)
}

// ─────────────────────────────────────────
// CloseYear
// ─────────────────────────────────────────

func (s *closingService) CloseYear(ctx context.Context, year int) (*db.PeriodClosing, error) {
	periodStart := fmt.Sprintf("%d-01-01", year)
	periodEnd := fmt.Sprintf("%d-12-31", year)

	// 1. 確認 12 個月都已月結
	if err := s.assertAllMonthsClosed(ctx, year); err != nil {
		return nil, err
	}

	// 2. 檢查是否已年結
	existing, err := s.closing.GetByPeriod(ctx, enums.PeriodAnnual, periodStart)
	if err != nil {
		return nil, fmt.Errorf("get period: %w", err)
	}
	if existing != nil && existing.Status.String() == enums.ClosingStatusClosed.String() {
		return nil, fmt.Errorf("year %d is already closed", year)
	}

	// 3. 計算本年損益
	is, err := s.reportSvc.GetIncomeStatement(ctx, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("income statement: %w", err)
	}
	netIncome := is.NetIncome

	// 4. Closing Entry：收入、支出結轉 3200 本期損益
	closingTxnID, err := s.appendClosingEntry(ctx, year, netIncome, is)
	if err != nil {
		return nil, fmt.Errorf("closing entry: %w", err)
	}

	// 5. 期初結轉：3200 → 3100
	openingTxnID, err := s.appendOpeningEntry(ctx, year, netIncome)
	if err != nil {
		return nil, fmt.Errorf("opening entry: %w", err)
	}

	// 6. 生成期末快照
	snapshot, err := s.buildSnapshot(ctx, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("build snapshot: %w", err)
	}
	snapshotJSON, err := MarshalSnapshot(*snapshot)
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot: %w", err)
	}

	// 7. 寫入年結紀錄
	now := time.Now().Format(time.RFC3339)
	err = s.closing.WithTx(ctx, func(tx query.ClosingTxRepository) error {
		var closingID int64

		if existing == nil {
			closingID, err = tx.Insert(ctx, db.PeriodClosing{
				PeriodType:  enums.PeriodAnnual,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				Status:      enums.ClosingStatusClosed,
			})
			if err != nil {
				return fmt.Errorf("insert closing: %w", err)
			}
		} else {
			closingID = existing.ClosingId
			if err := tx.UpdateStatus(ctx, closingID, enums.ClosingStatusClosed, &now); err != nil {
				return fmt.Errorf("update status: %w", err)
			}
		}

		if err := tx.UpdateClosingTxn(ctx, closingID, openingTxnID, closingTxnID); err != nil {
			return fmt.Errorf("update closing txn: %w", err)
		}

		return tx.UpdateSnapshot(ctx, closingID, snapshotJSON)
	})
	if err != nil {
		return nil, err
	}

	return s.closing.GetByPeriod(ctx, enums.PeriodAnnual, periodStart)
}

// ─────────────────────────────────────────
// Reopen
// ─────────────────────────────────────────

func (s *closingService) Reopen(ctx context.Context, closingID int64, reason string) error {
	return s.closing.WithTx(ctx, func(tx query.ClosingTxRepository) error {
		return tx.UpdateStatus(ctx, closingID, enums.ClosingStatusReopened, nil)
	})
}

// ─────────────────────────────────────────
// AssertNotClosed
// ─────────────────────────────────────────

func (s *closingService) AssertNotClosed(ctx context.Context, txnDate string) error {
	closed, err := s.closing.IsDateInClosedPeriod(ctx, txnDate)
	if err != nil {
		return fmt.Errorf("check closed period: %w", err)
	}
	if closed {
		return fmt.Errorf("date %s is in a closed period, cannot add transactions", txnDate)
	}
	return nil
}

// ─────────────────────────────────────────
// ReopenYear 重新開帳
// ─────────────────────────────────────────

func (s *closingService) ReopenYear(ctx context.Context, closingID int64, reason string) error {
	// 1. 拿到年結紀錄
	closing, err := s.closing.GetByID(ctx, closingID)
	if err != nil {
		return err
	}
	if closing.PeriodType.String() != enums.PeriodAnnual.String() {
		return fmt.Errorf("closing %d is not an annual closing", closingID)
	}
	if closing.Status.String() != enums.ClosingStatusClosed.String() {
		return fmt.Errorf("closing %d is not closed", closingID)
	}

	year, _ := strconv.Atoi(closing.PeriodStart[:4])
	entries, err := s.closing.GetClosingEntries(ctx, *closing.ClosingTxnID)
	if err != nil {
		return fmt.Errorf("get closing entries: %w", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("closing entry %d has no journal entries", *closing.ClosingTxnID)
	}

	// 2. 沖銷 Closing Entry
	if closing.ClosingTxnID != nil {
		_, err = s.eventSvc.Append(ctx, cmd.AppendCmd{
			AggregateType: enums.AggregateTransaction,
			AggregateID:   fmt.Sprintf("void-closing-%s", year),
			EventType:     event_types.EventTransactionCreated,
			Payload: payload.TransactionCreatedPayload{
				TransactionDate: closing.PeriodEnd,
				Description:     fmt.Sprintf("%s 年度結帳沖銷", year),
				// 從原始 Closing Entry 反向
				Entries: s.reverseEntries(entries),
			},
		})
		if err != nil {
			return fmt.Errorf("void closing entry: %w", err)
		}
	}

	// 3. 沖銷 Opening Entry（下一年的 1/1）
	nextYear := year + 1
	if closing.OpeningTxnID != nil {
		openingEntries, err := s.closing.GetClosingEntries(ctx, *closing.OpeningTxnID)
		if err != nil {
			return fmt.Errorf("get opening entries: %w", err)
		}
		_, err = s.eventSvc.Append(ctx, cmd.AppendCmd{
			AggregateType: enums.AggregateTransaction,
			AggregateID:   fmt.Sprintf("void-opening-%s", nextYear),
			EventType:     event_types.EventTransactionCreated,
			Payload: payload.TransactionCreatedPayload{
				TransactionDate: fmt.Sprintf("%s-01-01", nextYear),
				Description:     fmt.Sprintf("%s 期初結轉沖銷", nextYear),
				Entries:         s.reverseEntries(openingEntries),
			},
		})
		if err != nil {
			return fmt.Errorf("void opening entry: %w", err)
		}
	}

	// 4. 更新年結狀態為 reopened
	return s.closing.WithTx(ctx, func(tx query.ClosingTxRepository) error {
		return tx.UpdateStatus(ctx, closingID, enums.ClosingStatusReopened, nil)
	})
}

func (s *closingService) reverseEntries(entries []projection.Entry) []payload.TransactionEntryPayload {
	reversed := make([]payload.TransactionEntryPayload, len(entries))
	for i, e := range entries {
		reversed[i] = payload.TransactionEntryPayload{
			AccountId: e.AccountId,
			Debit:     e.Credit, // 借貸對調
			Credit:    e.Debit,
		}
	}
	return reversed
}

// ─────────────────────────────────────────
// Closing Entry（年結用）
// 收入貸方沖銷 → 借方
// 支出借方沖銷 → 貸方
// 差額轉入 3200 本期損益
// ─────────────────────────────────────────

func (s *closingService) appendClosingEntry(
	ctx context.Context,
	year int,
	netIncome decimal.Decimal,
	is *report.IncomeStatement,
) (int64, error) {
	entries := []payload.TransactionEntryPayload{}

	// 沖銷所有收入科目（收入正常貸方 → 借方沖銷）
	for _, row := range is.Income {
		if row.Amount.IsZero() {
			continue
		}
		entries = append(entries, payload.TransactionEntryPayload{
			AccountId: row.AccountID,
			Debit:     row.Amount,
			Credit:    decimal.Zero,
		})
	}

	// 沖銷所有支出科目（支出正常借方 → 貸方沖銷）
	for _, row := range is.Expenses {
		if row.Amount == decimal.Zero {
			continue
		}
		entries = append(entries, payload.TransactionEntryPayload{
			AccountId: row.AccountID,
			Debit:     decimal.Zero,
			Credit:    row.Amount,
		})
	}

	// 差額轉入 3200 本期損益
	// 淨利 → 3200 貸方；淨損 → 3200 借方
	if netIncome.IsPositive() {
		entries = append(entries, payload.TransactionEntryPayload{
			AccountId: "3200",
			Debit:     decimal.Zero,
			Credit:    netIncome,
		})
	} else {
		entries = append(entries, payload.TransactionEntryPayload{
			AccountId: "3200",
			Debit:     netIncome.Abs(),
			Credit:    decimal.Zero,
		})
	}

	event, err := s.eventSvc.Append(ctx, cmd.AppendCmd{
		AggregateType: enums.AggregateTransaction,
		AggregateID:   fmt.Sprintf("closing-%d", year),
		EventType:     event_types.EventTransactionCreated,
		Payload: payload.TransactionCreatedPayload{
			TransactionDate: fmt.Sprintf("%d-12-31", year),
			Description:     fmt.Sprintf("%d 年度結帳 Closing Entry", year),
			TotalAmount:     netIncome.Abs(),
			Currency:        "TWD",
			Entries:         entries,
		},
	})
	if err != nil {
		return 0, err
	}
	return event.EventId, nil
}

// ─────────────────────────────────────────
// Opening Entry（年結後期初結轉）
// 3200 本期損益 → 3100 期初淨資產
// ─────────────────────────────────────────

func (s *closingService) appendOpeningEntry(ctx context.Context, year int, netIncome decimal.Decimal) (int64, error) {
	// 3200 結轉到 3100
	// 本期有淨利：3200 借方，3100 貸方
	// 本期有淨損：3200 貸方，3100 借方
	var entries []payload.TransactionEntryPayload
	if netIncome.IsPositive() {
		entries = []payload.TransactionEntryPayload{
			{AccountId: "3200", Debit: netIncome, Credit: decimal.Zero},
			{AccountId: "3100", Debit: decimal.Zero, Credit: netIncome},
		}
	} else {
		abs := netIncome.Abs()
		entries = []payload.TransactionEntryPayload{
			{AccountId: "3200", Debit: decimal.Zero, Credit: abs},
			{AccountId: "3100", Debit: abs, Credit: decimal.Zero},
		}
	}

	event, err := s.eventSvc.Append(ctx, cmd.AppendCmd{
		AggregateType: enums.AggregateTransaction,
		AggregateID:   fmt.Sprintf("opening-%d", year+1),
		EventType:     event_types.EventTransactionCreated,
		Payload: payload.TransactionCreatedPayload{
			TransactionDate: fmt.Sprintf("%d-01-01", year+1),
			Description:     fmt.Sprintf("%d 期初餘額結轉", year+1),
			TotalAmount:     netIncome.Abs(),
			Currency:        "TWD",
			Entries:         entries,
		},
	})
	if err != nil {
		return 0, err
	}
	return event.EventId, nil
}

// ─────────────────────────────────────────
// helpers
// ─────────────────────────────────────────

func (s *closingService) buildSnapshot(ctx context.Context, date string) (*db.SnapshotData, error) {
	bs, err := s.reportSvc.GetBalanceSheet(ctx, date)
	if err != nil {
		return nil, err
	}

	snap := &db.SnapshotData{
		TotalAssets:      bs.TotalAssets,
		TotalLiabilities: bs.TotalLiabilities,
		NetWorth:         bs.NetWorth,
	}
	for _, r := range bs.Assets {
		snap.Assets = append(snap.Assets, db.BalanceSheetEntry{
			AccountId: r.AccountID,
			Name:      r.Name,
			Balance:   r.Balance,
		})
	}
	for _, r := range bs.Liabilities {
		snap.Liabilities = append(snap.Liabilities, db.BalanceSheetEntry{
			AccountId: r.AccountID,
			Name:      r.Name,
			Balance:   r.Balance,
		})
	}
	for _, r := range bs.Equity {
		snap.Equity = append(snap.Equity, db.BalanceSheetEntry{
			AccountId: r.AccountID,
			Name:      r.Name,
			Balance:   r.Balance,
		})
	}
	return snap, nil
}

func (s *closingService) assertNoUnresolvedAdjustments(ctx context.Context, start, end string) error {
	// 查 v_unresolved_adjustments 是否有在這個期間的未解決差異
	var count int
	err := s.db.QueryRowxContext(ctx, `
		SELECT COUNT(*)
		FROM v_unresolved_adjustments va
		JOIN reconciliations r ON va.recon_id = r.recon_id
		WHERE r.recon_date BETWEEN ? AND ?`,
		start, end,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("check unresolved: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("period has %d unresolved reconciliation adjustments, please resolve before closing", count)
	}
	return nil
}

func (s *closingService) assertAllMonthsClosed(ctx context.Context, year int) error {
	for m := time.January; m <= time.December; m++ {
		start, _ := monthRange(year, m)
		c, err := s.closing.GetByPeriod(ctx, enums.PeriodMonthly, start)
		if err != nil {
			return err
		}
		if c == nil || c.Status.String() != enums.ClosingStatusClosed.String() {
			return fmt.Errorf("month %d-%02d is not closed yet", year, m)
		}
	}
	return nil
}

func monthRange(year int, month time.Month) (start, end string) {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	return first.Format("2006-01-02"), last.Format("2006-01-02")
}

// SnapshotJSON 是給外部用的輔助函式
func UnmarshalSnapshot(raw string) (*db.SnapshotData, error) {
	var snap db.SnapshotData
	if err := json.Unmarshal([]byte(raw), &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

func MarshalSnapshot(data db.SnapshotData) (string, error) {
	data.GeneratedAt = time.Now().Format(time.RFC3339)
	b, err := json.Marshal(data)
	return string(b), err
}
