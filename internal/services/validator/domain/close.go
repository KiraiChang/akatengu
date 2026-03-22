package domain

import (
	"akatengu/internal/model/enums"
	"akatengu/internal/model/payload"
	"context"
	"encoding/json"
	"fmt"
)

func (v *domainValidator) validatePeriodClosed(ctx context.Context, raw json.RawMessage) error {
	var p payload.PeriodClosedPayload
	err := json.Unmarshal(raw, &p)
	if err != nil {
		return err
	}

	// 1. 檢查是否已結帳
	existing, err := v.query.Closing.GetByPeriod(ctx, enums.PeriodMonthly, p.PeriodStart)
	if err != nil {
		return fmt.Errorf("get period: %w", err)
	}
	if existing != nil && existing.Status.String() == enums.ClosingStatusClosed.String() {
		return fmt.Errorf("period %s~%s is already closed", p.PeriodStart, p.PeriodEnd)
	}

	return v.query.Closing.AssertNoUnresolvedAdjustments(ctx, p.PeriodStart, p.PeriodEnd)
}

//
// helper
//
