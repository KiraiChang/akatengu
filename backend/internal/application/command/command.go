package command

import (
	domainevent "akatengu/internal/domain/event"
	kerrors "akatengu/internal/kernel/errors"
	kevent "akatengu/internal/kernel/event"
	dbmodel "akatengu/internal/model/db"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/repos/query"
	"akatengu/internal/runtime/engine"
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

var (
	// ErrLegacyNotAccepted rejects a registered event type that has no legacy
	// aggregate: POST /api/event/append does not accept it (ADR-027 rule 2).
	ErrLegacyNotAccepted = errors.New("event type is migrated to BFS and is not accepted by POST /api/event/append")
	// ErrLegacyAggregateMismatch rejects a request whose aggregate is not the
	// legacy aggregate of its event type (ADR-027 rule 3).
	ErrLegacyAggregateMismatch = errors.New("aggregate does not match the legacy aggregate of the event type")
)

// Build makes the root event of a write from a registered payload. The
// registry derives the aggregate and the version mode (ADR-027 rules 3, 4);
// expectedVersion is used only by event types whose policy is to validate the
// client's version. MerchantID, UpdatedBy and an empty params.RequestID come
// from ctx.
func Build(ctx context.Context, payload kevent.Payload, expectedVersion int64, params kevent.EventParams) (kevent.Event, error) {
	spec, ok := domainevent.Lookup(payload.EventType())
	if !ok {
		return kevent.Event{}, kerrors.NewContractError(kerrors.ErrUnknownEventType, kevent.Event{EventType: payload.EventType()}, nil)
	}
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return kevent.Event{}, err
	}
	if params.RequestID == "" {
		params.RequestID = ctxkey.GetRequestID(ctx)
	}

	// The aggregate may be derived from the event UUID, so it is set after the
	// event is made.
	evt := kevent.NewFullEvent[kevent.Payload]("", payload, params)
	aggType, aggID, err := spec.Aggregate(evt)
	if err != nil {
		return kevent.Event{}, fmt.Errorf("derive aggregate of %v: %w", evt.EventType, err)
	}
	evt.AggregateType = aggType
	evt.AggregateUuid = aggID
	if params.PartitionKey == "" {
		evt.Metadata.Execution.PartitionKey = aggID
	}
	evt.MerchantID = merchantID
	evt.UpdatedBy = ctxkey.GetUserName(ctx)

	switch spec.Version {
	case domainevent.VersionValidate:
		evt.RootVersion = kevent.RootVersionExpected
		evt.ExpectedVersion = expectedVersion
	case domainevent.VersionNew:
		evt.RootVersion = kevent.RootVersionNew
	case domainevent.VersionInTx:
		evt.RootVersion = kevent.RootVersionCurrent
	default:
		return kevent.Event{}, fmt.Errorf("unknown version policy %q of %v", spec.Version, evt.EventType)
	}
	return evt, nil
}

// Writer runs root events through the BFS engine and reads the stored event
// back, so a migrated write answers with the same record as the legacy path.
type Writer struct {
	engine *engine.BFSEngine
	events query.EventRepo
}

func NewWriter(db *sqlx.DB, eng *engine.BFSEngine) *Writer {
	return &Writer{engine: eng, events: query.NewEventRepo(db)}
}

// Append runs evt and returns its event_store record.
func (w *Writer) Append(ctx context.Context, evt kevent.Event) (*dbmodel.EventStore, error) {
	if _, err := w.engine.Run(ctx, evt); err != nil {
		return nil, err
	}
	return w.events.GetByUUID(ctx, evt.Uuid)
}

// DispatchLegacy writes a POST /api/event/append request through BFS when its
// event type is registered (ADR-027 rules 2, 3). handled is false only for an
// unregistered event type, which the caller hands to the legacy path; once
// handled, an error must not fall back to it.
func (w *Writer) DispatchLegacy(ctx context.Context, req cmd.AppendCmd, params kevent.EventParams) (result *dbmodel.EventStore, handled bool, err error) {
	spec, ok := domainevent.Lookup(req.EventType)
	if !ok {
		return nil, false, nil
	}
	if spec.Legacy == nil {
		return nil, true, fmt.Errorf("%w: %v", ErrLegacyNotAccepted, req.EventType)
	}

	decoded, err := domainevent.Decode(req.EventType, req.Payload)
	if err != nil {
		return nil, true, err
	}
	legacyID, err := spec.Legacy.AggregateID(decoded)
	if err != nil {
		return nil, true, fmt.Errorf("legacy aggregate of %v: %w", req.EventType, err)
	}
	if req.AggregateType != spec.Legacy.AggregateType || req.AggregateID != legacyID {
		return nil, true, fmt.Errorf("%w: %v got (%v, %q), want (%v, %q)", ErrLegacyAggregateMismatch,
			req.EventType, req.AggregateType, req.AggregateID, spec.Legacy.AggregateType, legacyID)
	}

	payload, ok := decoded.(kevent.Payload)
	if !ok {
		return nil, true, fmt.Errorf("decoded payload %T of %v is not an event payload", decoded, req.EventType)
	}
	evt, err := Build(ctx, payload, req.ExpectedVersion, params)
	if err != nil {
		return nil, true, err
	}
	result, err = w.Append(ctx, evt)
	return result, true, err
}
