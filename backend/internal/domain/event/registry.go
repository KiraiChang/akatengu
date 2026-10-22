package event

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	kerrors "akatengu/internal/kernel/errors"
	kevent "akatengu/internal/kernel/event"
	"encoding/json"
	"fmt"
	"sync"
)

// The event registry lists every event type migrated to the BFS engine
// (ADR-027 rule 1). R2's dispatch of POST /api/event/append and R3's rebuild
// both read it. It only declares, decodes and derives identities, so it must
// not depend on the engine, the mediator or persistence.

// VersionPolicy is how the version of an event type's root event is chosen
// (ADR-027 rule 4).
type VersionPolicy string

const (
	// VersionNew creates a new aggregate; an existing one is a conflict.
	VersionNew VersionPolicy = "new"
	// VersionValidate uses and checks the version the client sends (ADR-023 A2 rule 1).
	VersionValidate VersionPolicy = "validate"
	// VersionInTx writes at the current version inside the transaction. Only for
	// event types whose legacy aggregate is not the derived one.
	VersionInTx VersionPolicy = "in_tx"
)

// LegacyAggregate is the aggregate POST /api/event/append accepts for an event
// type. The aggregate written is always the one Spec.Aggregate derives
// (ADR-027 rule 3).
type LegacyAggregate struct {
	AggregateType enums.AggregateType
	AggregateID   func(payload any) (string, error)
}

// Spec registers one event type.
type Spec struct {
	EventType event_types.EventType
	// Decode turns the payload JSON stored in event_store into the payload value
	// type that mediator handlers and projectors assert.
	Decode func(raw json.RawMessage) (any, error)
	// Aggregate derives the aggregate the event belongs to (ADR-024 A5, A7).
	Aggregate func(payload any) (enums.AggregateType, string, error)
	Version   VersionPolicy
	// Legacy is nil when POST /api/event/append does not accept the event type.
	Legacy *LegacyAggregate
}

// Specs lists every registered event type. It is built on each call, after the
// enums are initialized, so it never runs at package initialization.
func Specs() []Spec {
	return []Spec{
		newSpec(func(p TransactionCFCategoryUpdatedPayload) (enums.AggregateType, string) {
			return enums.AggregateTransaction.Enum(), p.TxnUUID
		}, VersionValidate, nil),
	}
}

var (
	registryOnce sync.Once
	registry     map[event_types.EventType]Spec
)

// Lookup returns the Spec of event type t.
func Lookup(t event_types.EventType) (Spec, bool) {
	registryOnce.Do(func() {
		specs := Specs()
		registry = make(map[event_types.EventType]Spec, len(specs))
		for _, s := range specs {
			registry[s.EventType] = s
		}
	})
	s, ok := registry[t]
	return s, ok
}

// Decode turns the payload JSON of an event of type t, as stored in
// event_store, into its payload value.
func Decode(t event_types.EventType, raw json.RawMessage) (any, error) {
	s, ok := Lookup(t)
	if !ok {
		return nil, kerrors.NewContractError(kerrors.ErrUnknownEventType, kevent.Event{EventType: t}, nil)
	}
	p, err := s.Decode(raw)
	if err != nil {
		return nil, kerrors.NewContractError(kerrors.ErrDecodePayload, kevent.Event{EventType: t}, err)
	}
	return p, nil
}

// newSpec builds a Spec for payload type T; the event type and Decode come
// from T, so they cannot disagree with it.
func newSpec[T kevent.Payload](aggregate func(T) (enums.AggregateType, string), version VersionPolicy, legacy *LegacyAggregate) Spec {
	var zero T
	return Spec{
		EventType: zero.EventType(),
		Decode: func(raw json.RawMessage) (any, error) {
			var p T
			if err := json.Unmarshal(raw, &p); err != nil {
				return nil, err
			}
			return p, nil
		},
		Aggregate: func(payload any) (enums.AggregateType, string, error) {
			p, ok := payload.(T)
			if !ok {
				return enums.AggregateType{}, "", fmt.Errorf("payload %T is not %T", payload, zero)
			}
			aggType, aggID := aggregate(p)
			return aggType, aggID, nil
		},
		Version: version,
		Legacy:  legacy,
	}
}
