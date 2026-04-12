package event_types

import (
	"akatengu/internal/pkg/enumx"
)

type EventType = enumx.Enum[eventTypeVal]

// enumx:enum
type eventTypeVal string
