package validator

import (
	"akatengu/internal/model/enums/event_types"
	"context"
	"encoding/json"
)

type Validator interface {
	Validate(ctx context.Context, eventType event_types.EventType, payload json.RawMessage) error
}
