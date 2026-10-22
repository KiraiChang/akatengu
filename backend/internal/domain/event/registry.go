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
	// Aggregate derives the aggregate the event belongs to (ADR-024 A5, A7)
	// from its payload, or from the event itself when the payload has no
	// identity, such as an aggregate named after the event UUID.
	Aggregate func(evt kevent.Event) (enums.AggregateType, string, error)
	Version   VersionPolicy
	// Legacy is nil when POST /api/event/append does not accept the event type.
	Legacy *LegacyAggregate
}

// Specs lists every registered event type. It is built on each call, after the
// enums are initialized, so it never runs at package initialization.
func Specs() []Spec {
	return []Spec{
		newSpec(func(_ kevent.Event, p TransactionCFCategoryUpdatedPayload) (enums.AggregateType, string) {
			return enums.AggregateTransaction.Enum(), p.TxnUUID
		}, VersionValidate, nil),

		// A ledger account is in its account's aggregate: ledger_account.updated
		// names it by the integer ledger_id, so no aggregate of its own can be
		// derived (BFS R2 D7).
		newSpecWithLegacy(enums.AggregateAccount.Enum(), func(p AccountCreatePayload) string { return p.AccountId }, VersionNew),
		newSpecWithLegacy(enums.AggregateAccount.Enum(), func(p AccountUpdatedPayload) string { return p.AccountId }, VersionValidate),
		newSpecWithLegacy(enums.AggregateAccount.Enum(), func(p LedgerAccountCreatePayload) string { return p.AccountId }, VersionValidate),
		newSpecWithLegacy(enums.AggregateAccount.Enum(), func(p LedgerAccountUpdatedPayload) string { return p.AccountId }, VersionValidate),
		// Only the seed writes the bulk import (bootstrap/event_seed.go).
		newSpec(func(_ kevent.Event, _ AccountBulkImportedPayload) (enums.AggregateType, string) {
			return enums.AggregateAccount.Enum(), MerchantAggregateID
		}, VersionValidate, nil),

		// The config events are written by their dedicated APIs and the seed;
		// each merchant has one SYS_CONFIG and one ACCOUNT_CONFIG aggregate.
		newSpec(func(_ kevent.Event, _ SysAccountUpdatedPayload) (enums.AggregateType, string) {
			return enums.AggregateSysConfig.Enum(), MerchantAggregateID
		}, VersionValidate, nil),
		newSpec(func(_ kevent.Event, _ LedgerAccountTypeConfigUpdatedPayload) (enums.AggregateType, string) {
			return enums.AggregateAccountConfig.Enum(), MerchantAggregateID
		}, VersionValidate, nil),
		newSpec(func(_ kevent.Event, _ AssetTypeAccountConfigUpdatedPayload) (enums.AggregateType, string) {
			return enums.AggregateAccountConfig.Enum(), MerchantAggregateID
		}, VersionValidate, nil),

		// A category's aggregate is its creation event's UUID, which is also its
		// category_uuid (BFS R2 D6). Updates and deletes send the version the
		// screen read (ADR-027 rule 6). Only the dedicated APIs write them.
		newSpec(func(evt kevent.Event, _ PrepaidCategoryCreatedPayload) (enums.AggregateType, string) {
			return enums.AggregatePrepaidCategory.Enum(), evt.Uuid
		}, VersionNew, nil),
		newSpec(func(_ kevent.Event, p PrepaidCategoryUpdatedPayload) (enums.AggregateType, string) {
			return enums.AggregatePrepaidCategory.Enum(), p.CategoryUUID
		}, VersionValidate, nil),
		newSpec(func(_ kevent.Event, p PrepaidCategoryDeletedPayload) (enums.AggregateType, string) {
			return enums.AggregatePrepaidCategory.Enum(), p.CategoryUUID
		}, VersionValidate, nil),
		newSpec(func(evt kevent.Event, _ FixedAssetCategoryCreatedPayload) (enums.AggregateType, string) {
			return enums.AggregateFixedAssetCategory.Enum(), evt.Uuid
		}, VersionNew, nil),
		newSpec(func(_ kevent.Event, p FixedAssetCategoryUpdatedPayload) (enums.AggregateType, string) {
			return enums.AggregateFixedAssetCategory.Enum(), p.CategoryUUID
		}, VersionValidate, nil),
		newSpec(func(_ kevent.Event, p FixedAssetCategoryDeletedPayload) (enums.AggregateType, string) {
			return enums.AggregateFixedAssetCategory.Enum(), p.CategoryUUID
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
// from T, so they cannot disagree with it. aggregate receives the event and
// its payload already asserted to T.
func newSpec[T kevent.Payload](aggregate func(kevent.Event, T) (enums.AggregateType, string), version VersionPolicy, legacy *LegacyAggregate) Spec {
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
		Aggregate: func(evt kevent.Event) (enums.AggregateType, string, error) {
			p, ok := evt.Payload.(T)
			if !ok {
				return enums.AggregateType{}, "", fmt.Errorf("payload %T is not %T", evt.Payload, zero)
			}
			aggType, aggID := aggregate(evt, p)
			return aggType, aggID, nil
		},
		Version: version,
		Legacy:  legacy,
	}
}

// newSpecWithLegacy builds a Spec whose legacy aggregate is the derived one:
// POST /api/event/append must send (aggType, id(payload)). Both use the same id,
// so they cannot disagree.
func newSpecWithLegacy[T kevent.Payload](aggType enums.AggregateType, id func(T) string, version VersionPolicy) Spec {
	var zero T
	return newSpec(func(_ kevent.Event, p T) (enums.AggregateType, string) {
		return aggType, id(p)
	}, version, &LegacyAggregate{
		AggregateType: aggType,
		AggregateID: func(payload any) (string, error) {
			p, ok := payload.(T)
			if !ok {
				return "", fmt.Errorf("payload %T is not %T", payload, zero)
			}
			return id(p), nil
		},
	})
}
