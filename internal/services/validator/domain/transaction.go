package domain

import (
	"akatengu/internal/model/enums"
	"akatengu/internal/model/payload"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func (d *domainValidator) validateEventTransactionCreated(ctx context.Context, raw json.RawMessage) error {
	var p payload.TransactionCreatedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed domainValidator: %w", err)
	}
	// 1. 轉換開張日期
	periodStart, err := d.periodStartDate(p.TransactionDate)
	if err != nil {
		return fmt.Errorf("parse date %s, fail: %w", p.TransactionDate, err)
	}

	// 2. 檢查是否已結帳
	existing, err := d.query.Period.GetByPeriod(ctx, enums.PeriodMonthly, periodStart)
	if err != nil {
		fmt.Errorf("get period: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("period %s is not exists", periodStart)
	}

	if existing.Status.String() != enums.PeriodTypeStatusOpen.String() {
		return fmt.Errorf("period %s~%s is already closed", existing.PeriodStart, existing.PeriodEnd)
	}
	return nil
}

func (d *domainValidator) periodStartDate(date string) (string, error) {
	datetime, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", err
	}
	first := time.Date(datetime.Year(), datetime.Month(), 1, 0, 0, 0, 0, time.UTC)
	return first.Format("2006-01-02"), nil
}
