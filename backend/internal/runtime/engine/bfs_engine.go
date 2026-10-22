package engine

import (
	"akatengu/internal/enums"
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"akatengu/internal/kernel/result"
	"akatengu/internal/model/db"
	"akatengu/internal/persistence/handle"
	"akatengu/internal/persistence/repos"
	"akatengu/internal/pkg/uuidx"
	"akatengu/internal/repos/query"
	"akatengu/internal/runtime/safety"
	"context"
	"encoding/json"
	"fmt"
)

type BFSEngine struct {
	executor    *Executor
	middlewares []safety.Middleware
	uow         repos.UnitOfWork
	projector   *handle.Registry
}

func NewBFSEngine(executor *Executor) *BFSEngine {
	return &BFSEngine{executor: executor}
}

func (e *BFSEngine) WithMiddleware(middlewares ...safety.Middleware) *BFSEngine {
	// Later calls wrap earlier ones, so their middlewares go first (outermost).
	e.middlewares = append(append([]safety.Middleware{}, middlewares...), e.middlewares...)
	return e
}

func (e *BFSEngine) WithUnitOfWork(uow repos.UnitOfWork) *BFSEngine {
	e.uow = uow
	return e
}

func (e *BFSEngine) WithProjector(r *handle.Registry) *BFSEngine {
	e.projector = r
	return e
}

// Run processes evt and every child event it leads to, in BFS order. With a
// UnitOfWork the whole event graph runs in one transaction (ADR-023 A1): if any
// handler, middleware, append or projector fails, nothing of this Run is
// committed. Without one, handlers run but nothing is written.
func (e *BFSEngine) Run(ctx context.Context, evt event.Event) (event.Event, error) {
	if e.uow == nil {
		return evt, e.run(ctx, evt, nil)
	}
	err := e.uow.Do(ctx, func(tx *repos.DbTransaction) error {
		return e.run(ctx, evt, tx)
	})
	return evt, err
}

func (e *BFSEngine) run(ctx context.Context, root event.Event, tx *repos.DbTransaction) error {
	// Stateful middlewares allocate their state when wrapping, so building the
	// chain once per Run keeps concurrent and successive Runs isolated. The
	// terminal holds this Run's transaction and aggregate versions the same way.
	process := safety.Chain(e.terminal(tx, aggregateVersions{}), e.middlewares...)
	q := &Queue{}
	q.Push(root, 0)

	for !q.Empty() {
		// The graph holds the only DB connection; stop as soon as the caller
		// gives up (ADR-023 A1 rule 4).
		if err := ctx.Err(); err != nil {
			return err
		}
		current, depth := q.Pop()
		r, err := process(ctx, depth, current)
		if err != nil {
			// r.Policy describes the error-handling strategy:
			//   Fatal=true     → unrecoverable, abort immediately.
			//   Retryable=true → transient, eligible for retry (retry logic TBD).
			// Currently all errors abort the run regardless of policy.
			return err
		}
		for _, child := range r.Events {
			next, err := inheritEnvelope(current, child)
			if err != nil {
				return err
			}
			q.Push(next, depth+1)
		}
	}
	return nil
}

// terminal dispatches an event to its handler and, with a transaction, writes
// it and applies the projectors. versions belongs to the current Run.
func (e *BFSEngine) terminal(tx *repos.DbTransaction, versions aggregateVersions) safety.EventProcessor {
	var repo *query.Repo
	if tx != nil {
		repo = tx.Query
	}
	return func(ctx context.Context, depth int, ev event.Event) (result.EventResult, error) {
		r, err := e.executor.Dispatch(ctx, repo, ev)
		if err != nil {
			// Pass through the result so callers can inspect r.Policy.
			return r, err
		}
		if tx == nil {
			return r, nil
		}

		eventId, newVersion, err := write(ctx, tx.Store, depth, ev, versions)
		if err != nil {
			return r, err
		}
		ev.Id = eventId
		versions.set(ev, newVersion)

		if e.projector != nil {
			if err := e.projector.ApplyAll(ctx, tx, ev); err != nil {
				return r, err
			}
		}

		// snapshot 決策
		if tx.Snap != nil {
			if newVersion%50 == 0 {
				state, _ := json.Marshal(ev)
				if err := tx.Snap.Upsert(ctx, db.Snapshot{
					SnapshotUuid:  uuidx.NewFromEvent(ev.Uuid, "snapshot"),
					AggregateType: ev.AggregateType,
					AggregateId:   ev.AggregateUuid,
					AtVersion:     newVersion,
					State:         state,
				}); err != nil {
					return r, fmt.Errorf("upsert snapshot: %w", err)
				}
			}
		}
		return r, nil
	}
}

// write appends ev with the version rule of ADR-023 A2: the root event follows
// its RootVersion, a child in its parent's aggregate continues from the
// version written earlier in this Run, and a derived aggregate is created.
func write(ctx context.Context, store repos.Store, depth int, ev event.Event, versions aggregateVersions) (int64, int64, error) {
	switch {
	case depth == 0:
		return writeRoot(ctx, store, ev)
	case ev.ChildAggregate == event.ChildAggregateDerived:
		return store.Create(ctx, ev)
	default:
		v, ok := versions.get(ev)
		if !ok {
			return 0, 0, kerrors.NewRuntimeError(kerrors.ErrChildEnvelopeInvalid, ev,
				fmt.Errorf("aggregate %s has not been written in this run", ev.AggregateUuid))
		}
		ev.ExpectedVersion = v
		return store.Append(ctx, ev)
	}
}

// writeRoot appends a root event with the version mode set from its event
// type's version policy (ADR-027 rule 4).
func writeRoot(ctx context.Context, store repos.Store, ev event.Event) (int64, int64, error) {
	switch ev.RootVersion {
	case event.RootVersionExpected:
		return store.Append(ctx, ev)
	case event.RootVersionNew:
		return store.Create(ctx, ev)
	case event.RootVersionCurrent:
		return store.AppendCurrent(ctx, ev)
	default:
		return 0, 0, kerrors.NewRuntimeError(kerrors.ErrRootEnvelopeInvalid, ev,
			fmt.Errorf("unknown root version mode %d", ev.RootVersion))
	}
}

// inheritEnvelope fills a child event's envelope from its parent (ADR-023 A2).
// The engine overwrites these fields, so a handler cannot set them otherwise.
func inheritEnvelope(parent, child event.Event) (event.Event, error) {
	switch child.ChildAggregate {
	case event.ChildAggregateParent:
		if (child.AggregateUuid != "" && child.AggregateUuid != parent.AggregateUuid) ||
			(!child.AggregateType.IsZero() && child.AggregateType != parent.AggregateType) {
			return child, kerrors.NewRuntimeError(kerrors.ErrChildEnvelopeInvalid, child,
				fmt.Errorf("aggregate %s/%s differs from the parent's %s/%s; use event.NewDerivedEvent",
					child.AggregateType, child.AggregateUuid, parent.AggregateType, parent.AggregateUuid))
		}
		child.AggregateType = parent.AggregateType
		child.AggregateUuid = parent.AggregateUuid
	case event.ChildAggregateDerived:
		if child.AggregateType.IsZero() || child.AggregateUuid == "" {
			return child, kerrors.NewRuntimeError(kerrors.ErrChildEnvelopeInvalid, child,
				fmt.Errorf("derived aggregate needs a type and an ID"))
		}
	default:
		return child, kerrors.NewRuntimeError(kerrors.ErrChildEnvelopeInvalid, child,
			fmt.Errorf("unknown child aggregate %d", child.ChildAggregate))
	}

	child.MerchantID = parent.MerchantID
	child.UpdatedBy = parent.UpdatedBy
	child.Metadata.Tracing.CorrelationID = parent.Metadata.Tracing.CorrelationID
	child.Metadata.Tracing.RequestID = parent.Metadata.Tracing.RequestID
	child.Metadata.Tracing.CausationID = parent.Uuid
	return child, nil
}

// aggregateVersions is the latest version of each aggregate written in a Run.
type aggregateVersions map[aggregateKey]int64

type aggregateKey struct {
	merchantID    int64
	aggregateType enums.AggregateType
	aggregateID   string
}

func keyOf(ev event.Event) aggregateKey {
	return aggregateKey{merchantID: ev.MerchantID, aggregateType: ev.AggregateType, aggregateID: ev.AggregateUuid}
}

func (v aggregateVersions) get(ev event.Event) (int64, bool) {
	version, ok := v[keyOf(ev)]
	return version, ok
}

func (v aggregateVersions) set(ev event.Event, version int64) {
	v[keyOf(ev)] = version
}
