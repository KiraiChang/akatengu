package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type AccountAnalysisRepo interface {
	GetAccountDirectChildrenWithBalance(ctx context.Context, accountID string) ([]projection.AccountChildBalance, error)
	GetAccountJournalEntriesPaged(ctx context.Context, accountID, fromDate, toDate string, req model.PaginationParams) ([]projection.AccountJournalEntryRow, int64, error)
	GetAccountMonthlyBalances(ctx context.Context, accountID, fromMonth, toMonth string) ([]projection.AccountMonthlyBalance, error)
}

type sqlxAccountAnalysisRepo struct {
	q *sqlcdb.Queries
}

func newAccountAnalysisRepo(q *sqlcdb.Queries) AccountAnalysisRepo {
	return &sqlxAccountAnalysisRepo{q: q}
}

func NewAccountAnalysisRepo(db *sqlx.DB) AccountAnalysisRepo {
	return &sqlxAccountAnalysisRepo{q: sqlcdb.New(db)}
}

func (r *sqlxAccountAnalysisRepo) GetAccountDirectChildrenWithBalance(ctx context.Context, accountID string) ([]projection.AccountChildBalance, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}

	children, err := r.q.GetChildrenAccount(ctx, sqlcdb.GetChildrenAccountParams{
		ParentID:   &accountID,
		MerchantID: merchantID,
	})
	if err != nil {
		return nil, err
	}
	// Fetch all running balances and index by account_id
	balances, err := r.q.GetAllAccountRunningBalance(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	type totals struct {
		debit  decimal.Decimal
		credit decimal.Decimal
	}
	balanceMap := make(map[string]totals, len(balances))
	for _, b := range balances {
		balanceMap[b.AccountID] = totals{debit: b.DebitTotal, credit: b.CreditTotal}
	}

	// 葉科目（無子科目）：回傳自身以利前端顯示餘額
	if len(children) == 0 {
		self, err := r.q.GetAccount(ctx, sqlcdb.GetAccountParams{
			AccountID:  accountID,
			MerchantID: merchantID,
		})
		if err != nil {
			return nil, err
		}
		bal := balanceMap[accountID]
		return []projection.AccountChildBalance{{
			AccountID:   self.AccountID,
			ParentID:    self.ParentID,
			Name:        self.Name,
			IsSummary:   self.IsSummary,
			HasChild:    false,
			DebitTotal:  bal.debit,
			CreditTotal: bal.credit,
		}}, nil
	}

	result := make([]projection.AccountChildBalance, len(children))
	for i, c := range children {
		bal := balanceMap[c.AccountID]
		result[i] = projection.AccountChildBalance{
			AccountID:   c.AccountID,
			ParentID:    c.ParentID,
			Name:        c.Name,
			IsSummary:   c.IsSummary,
			HasChild:    c.HasChild,
			DebitTotal:  bal.debit,
			CreditTotal: bal.credit,
		}
	}
	return result, nil
}

func (r *sqlxAccountAnalysisRepo) GetAccountJournalEntriesPaged(ctx context.Context, accountID, fromDate, toDate string, req model.PaginationParams) ([]projection.AccountJournalEntryRow, int64, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.q.GetAccountJournalEntriesPaged(ctx, sqlcdb.GetAccountJournalEntriesPagedParams{
		AccountID:  accountID,
		MerchantID: merchantID,
		FromDate:   fromDate,
		ToDate:     toDate,
		Offset:     req.Offset,
		Limit:      req.Limit,
	})
	if err != nil {
		return nil, 0, err
	}

	total := int64(0)
	if len(rows) > 0 {
		total = rows[0].Total
	}

	result := make([]projection.AccountJournalEntryRow, len(rows))
	for i, row := range rows {
		result[i] = projection.AccountJournalEntryRow{
			EntryID:     row.EntryID,
			TxnID:       row.TxnID,
			AccountID:   row.AccountID,
			LedgerID:    row.LedgerID,
			Debit:       row.Debit,
			Credit:      row.Credit,
			Note:        row.Note,
			TxnDate:     row.TxnDate,
			Description: row.Description,
		}
	}
	return result, total, nil
}

func (r *sqlxAccountAnalysisRepo) GetAccountMonthlyBalances(ctx context.Context, accountID, fromMonth, toMonth string) ([]projection.AccountMonthlyBalance, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}

	from, err := time.Parse("2006-01", fromMonth)
	if err != nil {
		return nil, fmt.Errorf("invalid from month: %w", err)
	}
	to, err := time.Parse("2006-01", toMonth)
	if err != nil {
		return nil, fmt.Errorf("invalid to month: %w", err)
	}
	if to.Before(from) {
		return nil, errors.New("to month must be >= from month")
	}

	fromDate := from.Format("2006-01-02")
	toDate := lastDayOfMonth(to).Format("2006-01-02")

	months := enumerateMonths(from, to)

	entries, err := r.q.GetAccountJournalEntriesInDateRange(ctx, sqlcdb.GetAccountJournalEntriesInDateRangeParams{
		AccountID:  accountID,
		MerchantID: merchantID,
		FromDate:   fromDate,
		ToDate:     toDate,
	})
	if err != nil {
		return nil, err
	}

	closedPeriods, err := r.q.GetClosedMonthlyPeriodsInRange(ctx, sqlcdb.GetClosedMonthlyPeriodsInRangeParams{
		MerchantID: merchantID,
		FromDate:   fromDate,
		ToDate:     toDate,
	})
	if err != nil {
		return nil, err
	}

	closedSet := make(map[string]bool, len(closedPeriods))
	for _, cp := range closedPeriods {
		if len(cp.PeriodEnd) >= 7 {
			closedSet[cp.PeriodEnd[:7]] = true
		}
	}

	type monthTotals struct {
		debit  decimal.Decimal
		credit decimal.Decimal
	}
	totals := make(map[string]*monthTotals, len(months))
	for _, m := range months {
		totals[m] = &monthTotals{}
	}
	for _, e := range entries {
		if len(e.TxnDate) >= 7 {
			ym := e.TxnDate[:7]
			if t, ok := totals[ym]; ok {
				t.debit = t.debit.Add(e.Debit)
				t.credit = t.credit.Add(e.Credit)
			}
		}
	}

	result := make([]projection.AccountMonthlyBalance, len(months))
	for i, m := range months {
		t := totals[m]
		result[i] = projection.AccountMonthlyBalance{
			Month:       m,
			DebitTotal:  t.debit,
			CreditTotal: t.credit,
			HasSnapshot: closedSet[m],
		}
	}
	return result, nil
}

func enumerateMonths(from, to time.Time) []string {
	var months []string
	cur := from
	for !cur.After(to) {
		months = append(months, cur.Format("2006-01"))
		cur = cur.AddDate(0, 1, 0)
	}
	return months
}

func lastDayOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location())
}
