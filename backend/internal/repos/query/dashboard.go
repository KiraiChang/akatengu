package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
	"context"

	"github.com/shopspring/decimal"

	"github.com/jmoiron/sqlx"
)

type DashboardRepo interface {
	GetSummary(ctx context.Context, today string) (projection.DashboardSummary, error)
	GetMonthlyTrend(ctx context.Context, fromMonth, toMonth string) ([]projection.MonthlyTrendItem, error)
	GetLedgerBalances(ctx context.Context) ([]projection.LedgerBalance, error)
}

type sqlxDashboardRepo struct {
	q  *sqlcdb.Queries
	db *sqlx.DB
}

func newDashboardRepo(q *sqlcdb.Queries, db *sqlx.DB) DashboardRepo {
	return &sqlxDashboardRepo{q: q, db: db}
}

func NewDashboardRepo(db *sqlx.DB) DashboardRepo {
	return &sqlxDashboardRepo{q: sqlcdb.New(db), db: db}
}

func (r *sqlxDashboardRepo) GetSummary(ctx context.Context, today string) (projection.DashboardSummary, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return projection.DashboardSummary{}, err
	}

	accounts, err := r.q.GetAllAccounts(ctx, merchantID)
	if err != nil {
		return projection.DashboardSummary{}, err
	}

	runningBalances, err := r.q.GetAllAccountRunningBalance(ctx, merchantID)
	if err != nil {
		return projection.DashboardSummary{}, err
	}

	type balEntry struct {
		debit  decimal.Decimal
		credit decimal.Decimal
	}
	balanceMap := make(map[string]balEntry, len(runningBalances))
	for _, b := range runningBalances {
		balanceMap[b.AccountID] = balEntry{debit: b.DebitTotal, credit: b.CreditTotal}
	}

	currentMonth := today[:7]
	trendRows, err := r.q.GetMonthlyTrend(ctx, sqlcdb.GetMonthlyTrendParams{
		MerchantID: merchantID,
		FromMonth:  currentMonth,
		ToMonth:    currentMonth,
	})
	if err != nil {
		return projection.DashboardSummary{}, err
	}

	var monthIncome, monthExpense decimal.Decimal
	if len(trendRows) > 0 {
		monthIncome = trendRows[0].Income
		monthExpense = trendRows[0].Expense
	}

	var totalAssets, totalLiabilities, totalEquity, cashBalance decimal.Decimal
	for _, a := range accounts {
		if a.IsSummary || !a.IsActive {
			continue
		}
		bal, ok := balanceMap[a.AccountID]
		if !ok {
			continue
		}

		var accountBalance decimal.Decimal
		if a.NormalBalance.Is(enums.BalanceDebit) {
			accountBalance = bal.debit.Sub(bal.credit)
		} else {
			accountBalance = bal.credit.Sub(bal.debit)
		}

		switch {
		case a.Type.Is(enums.AccountAsset):
			totalAssets = totalAssets.Add(accountBalance)
			if a.CashFlowCategory.Is(enums.CashFlowCategoryCash) {
				cashBalance = cashBalance.Add(accountBalance)
			}
		case a.Type.Is(enums.AccountLiability):
			totalLiabilities = totalLiabilities.Add(accountBalance)
		case a.Type.Is(enums.AccountEquity):
			totalEquity = totalEquity.Add(accountBalance)
		}
	}

	return projection.DashboardSummary{
		AsOfDate:         today,
		Month:            currentMonth,
		MonthIncome:      monthIncome,
		MonthExpense:     monthExpense,
		TotalAssets:      totalAssets,
		TotalLiabilities: totalLiabilities,
		TotalEquity:      totalEquity,
		CashBalance:      cashBalance,
	}, nil
}

func (r *sqlxDashboardRepo) GetMonthlyTrend(ctx context.Context, fromMonth, toMonth string) ([]projection.MonthlyTrendItem, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := r.q.GetMonthlyTrend(ctx, sqlcdb.GetMonthlyTrendParams{
		MerchantID: merchantID,
		FromMonth:  fromMonth,
		ToMonth:    toMonth,
	})
	if err != nil {
		return nil, err
	}

	result := make([]projection.MonthlyTrendItem, len(rows))
	for i, row := range rows {
		result[i] = projection.MonthlyTrendItem{
			Month:   row.Month,
			Income:  row.Income,
			Expense: row.Expense,
			Net:     row.Income.Sub(row.Expense),
		}
	}
	return result, nil
}

func (r *sqlxDashboardRepo) GetLedgerBalances(ctx context.Context) ([]projection.LedgerBalance, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}

	ledgers, err := r.q.GetAllLedgers(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	runningBalances, err := r.q.GetAllLedgerRunningBalance(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	type balEntry struct {
		debit         decimal.Decimal
		credit        decimal.Decimal
		normalBalance enums.NormalBalance
	}
	balanceMap := make(map[int64]balEntry, len(runningBalances))
	for _, b := range runningBalances {
		balanceMap[b.LedgerID] = balEntry{
			debit:         b.DebitTotal,
			credit:        b.CreditTotal,
			normalBalance: b.NormalBalance,
		}
	}

	result := make([]projection.LedgerBalance, 0, len(ledgers))
	for _, l := range ledgers {
		if !l.IsActive {
			continue
		}
		bal, ok := balanceMap[l.LedgerID]
		var balance decimal.Decimal
		if ok {
			if bal.normalBalance.Is(enums.BalanceDebit) {
				balance = bal.debit.Sub(bal.credit)
			} else {
				balance = bal.credit.Sub(bal.debit)
			}
		}
		result = append(result, projection.LedgerBalance{
			LedgerID:    l.LedgerID,
			Name:        l.Name,
			Institution: l.Institution,
			Type:        l.Type.String(),
			Balance:     balance,
		})
	}
	return result, nil
}
