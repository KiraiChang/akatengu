package domain

import (
	"akatengu/internal/model/enums"
	"akatengu/internal/model/payload"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func (d *domainValidator) validateEventPeriodMonthClosed(ctx context.Context, raw json.RawMessage) error {
	var p payload.PeriodMonthClosedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed domainValidator: %w", err)
	}

	// 1. 檢查是否已結帳
	existing, err := d.query.Period.GetByID(ctx, p.ClosingId)
	if err != nil {
		fmt.Errorf("get period: %w", err)
	}
	if existing != nil && existing.Status.String() == enums.PeriodTypeStatusClosed.String() {
		return fmt.Errorf("period %s~%s is already closed", existing.PeriodStart, existing.PeriodEnd)
	}

	// 2. 檢查有無未解決的對帳差異
	return d.query.Period.AssertNoUnresolvedAdjustments(ctx, existing.PeriodStart, existing.PeriodEnd)
}

func (d *domainValidator) validateEventPeriodAnnualClosed(ctx context.Context, raw json.RawMessage) error {
	var p payload.PeriodAnnualClosedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed domainValidator: %w", err)
	}

	// 1. 檢查是否已年結 (要設定資料)
	existing, err := d.query.Period.GetByID(ctx, p.ClosingId)
	if err != nil {
		return fmt.Errorf("get period: %w", err)
	}
	t, _ := time.Parse("2006-01-02", existing.PeriodStart)
	year := t.Year()
	if existing != nil && existing.Status.String() == enums.PeriodTypeStatusClosed.String() {
		return fmt.Errorf("year %d is already closed", year)
	}

	// 2. 確認 12 個月都已月結
	if err := d.assertAllMonthsClosed(ctx, year); err != nil {
		return err
	}

	return nil
}

func (d *domainValidator) assertAllMonthsClosed(ctx context.Context, year int) error {
	for m := time.January; m <= time.December; m++ {
		start, _ := d.monthRange(year, m)
		c, err := d.query.Period.GetByPeriod(ctx, enums.PeriodMonthly, start)
		if err != nil {
			return err
		}
		if c == nil || c.Status.String() != enums.PeriodTypeStatusClosed.String() {
			return fmt.Errorf("month %d-%02d is not closed yet", year, m)
		}
	}
	return nil
}

func (d *domainValidator) monthRange(year int, month time.Month) (start, end string) {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	return first.Format("2006-01-02"), last.Format("2006-01-02")
}
