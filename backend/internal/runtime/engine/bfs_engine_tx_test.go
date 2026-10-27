package engine

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/event"
	"akatengu/internal/kernel/result"
	"akatengu/internal/persistence/handle"
	"akatengu/internal/persistence/repos"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/repos/query"
	"akatengu/internal/runtime/mediator"
	"akatengu/internal/shared/utils/test"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// ─── Recorder and handlers ───────────────────────────────────────────────────

// txRecorder lets the handlers of a scenario report what they saw.
type txRecorder struct {
	cancel     context.CancelFunc
	childCalls int      // typeB handler calls
	rootFound  int      // root events found by queryRoot
	projected  []string // Note of each payload the projector applied
}

var errTxChild = errors.New("r1 child handler failed")

func childOf(parent event.Event) event.Event {
	return event.NewNextEvent(parent, testPayload{typ: typeB})
}

// emitSame returns n typeB children in the root's aggregate.
func emitSame(n int) func(*txRecorder) mediator.HandlerFunc {
	return func(_ *txRecorder) mediator.HandlerFunc {
		return func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
			children := make([]event.Event, n)
			for i := range children {
				children[i] = childOf(evt)
			}
			return result.EventResult{Events: children}, nil
		}
	}
}

// emitDerived returns one typeB child per purpose, each in a derived aggregate.
func emitDerived(purposes ...string) func(*txRecorder) mediator.HandlerFunc {
	return func(_ *txRecorder) mediator.HandlerFunc {
		return func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
			children := make([]event.Event, len(purposes))
			for i, p := range purposes {
				children[i] = event.NewDerivedEvent(evt, enums.AggregateTransaction.Enum(), p, testPayload{typ: typeB})
			}
			return result.EventResult{Events: children}, nil
		}
	}
}

func cancelThenEmit(rec *txRecorder) mediator.HandlerFunc {
	return func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
		rec.cancel()
		return result.EventResult{Events: []event.Event{childOf(evt)}}, nil
	}
}

// emitWithOtherEnvelope sets envelope fields the engine must overwrite.
func emitWithOtherEnvelope(_ *txRecorder) mediator.HandlerFunc {
	return func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
		child := childOf(evt)
		child.MerchantID = 2
		child.UpdatedBy = "r1-other"
		child.Metadata.Tracing.CorrelationID = "r1-other-correlation"
		child.Metadata.Tracing.RequestID = "r1-other-request"
		child.Metadata.Tracing.CausationID = "r1-other-causation"
		return result.EventResult{Events: []event.Event{child}}, nil
	}
}

// emitOtherAggregate moves a child to another aggregate without NewDerivedEvent.
func emitOtherAggregate(_ *txRecorder) mediator.HandlerFunc {
	return func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
		child := childOf(evt)
		child.AggregateUuid = "0192f3d0-0000-7000-8000-0000000000b1"
		return result.EventResult{Events: []event.Event{child}}, nil
	}
}

func recordChild(rec *txRecorder) mediator.HandlerFunc {
	return func(_ context.Context, _ *query.Repo, _ event.Event) (result.EventResult, error) {
		rec.childCalls++
		return result.EventResult{}, nil
	}
}

// emitGrandchild makes a direct child of the root emit one grandchild in its
// own aggregate; the grandchild emits nothing.
func emitGrandchild(rec *txRecorder) mediator.HandlerFunc {
	return func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
		rec.childCalls++
		if evt.Metadata.Tracing.CausationID != txRootUUID {
			return result.EventResult{}, nil
		}
		return result.EventResult{Events: []event.Event{childOf(evt)}}, nil
	}
}

func failChild(rec *txRecorder) mediator.HandlerFunc {
	return func(_ context.Context, _ *query.Repo, _ event.Event) (result.EventResult, error) {
		rec.childCalls++
		return result.EventResult{}, errTxChild
	}
}

// queryRoot reads the root aggregate with the Repo the engine passes in, before
// the Run commits.
func queryRoot(rec *txRecorder) mediator.HandlerFunc {
	return func(ctx context.Context, repo *query.Repo, _ event.Event) (result.EventResult, error) {
		rec.childCalls++
		events, err := repo.Event.GetByAggregate(ctx, enums.AggregateTransaction.Enum(), txRootAggID)
		if err != nil {
			return result.EventResult{}, err
		}
		for _, e := range events {
			if e.EventUuid == txRootUUID {
				rec.rootFound++
			}
		}
		return result.EventResult{}, nil
	}
}

// enrichRoot returns the root's payload with Note set, and one typeB child.
func enrichRoot(_ *txRecorder) mediator.HandlerFunc {
	return func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
		p := evt.Payload.(testPayload)
		p.Note = txEnrichedNote
		return result.EventResult{Payload: p, Events: []event.Event{childOf(evt)}}, nil
	}
}

// otherTestPayload is not the type of the root's payload.
type otherTestPayload struct{ Note string }

func enrichWithOtherType(_ *txRecorder) mediator.HandlerFunc {
	return func(_ context.Context, _ *query.Repo, _ event.Event) (result.EventResult, error) {
		return result.EventResult{Payload: otherTestPayload{Note: txEnrichedNote}}, nil
	}
}

// expectEnrichedRoot checks that event_store and the projector got the root's
// enriched payload, and the child's own payload.
func expectEnrichedRoot(rec *txRecorder, db *sqlx.DB) {
	var payloads []string
	Expect(db.Select(&payloads, `SELECT payload FROM event_store ORDER BY event_id`)).To(Succeed())
	Expect(payloads).To(Equal([]string{`{"note":"` + txEnrichedNote + `"}`, `{}`}))
	Expect(rec.projected).To(Equal([]string{txEnrichedNote, ""}))
}

func expectChildCalls(n int) func(*txRecorder, *sqlx.DB) {
	return func(rec *txRecorder, _ *sqlx.DB) {
		Expect(rec.childCalls).To(Equal(n))
	}
}

func expectRootFound(rec *txRecorder, _ *sqlx.DB) {
	Expect(rec.rootFound).To(Equal(1))
}

func expectChildEnvelope(_ *txRecorder, db *sqlx.DB) {
	var row struct {
		MerchantID int64  `db:"merchant_id"`
		UpdatedBy  string `db:"updated_by"`
		Metadata   string `db:"metadata"`
	}
	err := db.Get(&row, `SELECT merchant_id, COALESCE(updated_by, '') AS updated_by, COALESCE(metadata, '') AS metadata
		FROM event_store WHERE event_type = ?`, txTypeB)
	Expect(err).NotTo(HaveOccurred())
	Expect(row.MerchantID).To(Equal(txMerchantID))
	Expect(row.UpdatedBy).To(Equal(txUserName))

	var meta event.Metadata
	Expect(json.Unmarshal([]byte(row.Metadata), &meta)).To(Succeed())
	Expect(meta.Tracing.CorrelationID).To(Equal(txCorrelationID))
	Expect(meta.Tracing.RequestID).To(Equal(txRequestID))
	Expect(meta.Tracing.CausationID).To(Equal(txRootUUID))
}

// ─── Projector and fixture ───────────────────────────────────────────────────

// txProjector handles typeA and typeB; its checkpoint shows whether the
// projector writes of a Run were committed.
type txProjector struct {
	failOnB bool
	rec     *txRecorder
}

func (p *txProjector) Name() string { return "r1TxProjector" }

func (p *txProjector) EventTypes() []event_types.EventType {
	return []event_types.EventType{typeA, typeB}
}

func (p *txProjector) Apply(_ context.Context, _ *repos.DbTransaction, evt event.Event) error {
	if p.failOnB && evt.EventType == typeB {
		return errors.New("r1 projector failed")
	}
	if tp, ok := evt.Payload.(testPayload); ok && p.rec != nil {
		p.rec.projected = append(p.rec.projected, tp.Note)
	}
	return nil
}

func (p *txProjector) Reset(_ context.Context, _ *repos.DbTransaction) error { return nil }

func newTxTestDB() *sqlx.DB {
	db, closeDB := test.NewTestDBGinkgo()
	DeferCleanup(closeDB)
	// Every new connection to ":memory:" is an empty database; keep a single one.
	db.SetMaxOpenConns(1)
	return db
}

func txRootEvent(expectedVersion int64, mode event.RootVersion) event.Event {
	evt := event.NewEvent[testPayload](txRootAggID, testPayload{typ: typeA}, event.Metadata{
		Tracing: event.TracingData{CorrelationID: txCorrelationID, RequestID: txRequestID},
	})
	evt.Uuid = txRootUUID
	evt.AggregateType = enums.AggregateTransaction.Enum()
	evt.MerchantID = txMerchantID
	evt.UpdatedBy = txUserName
	evt.ExpectedVersion = expectedVersion
	evt.RootVersion = mode
	return evt
}

// ─── Spec runner ─────────────────────────────────────────────────────────────

var _ = Describe("BFSEngine Run with a DB: event graph transaction and child events", func() {
	for _, s := range txScenarios {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)

		It(label, func() {
			db := newTxTestDB()
			for _, v := range s.seedVersions {
				_, err := db.Exec(`INSERT INTO aggregate_versions(aggregate_type, aggregate_id, merchant_id, current_version)
					VALUES(?, ?, ?, ?)`, v.AggregateType, v.AggregateID, txMerchantID, v.Version)
				Expect(err).NotTo(HaveOccurred())
			}

			base := context.WithValue(context.Background(), ctxkey.MerchantID, txMerchantID)
			var ctx context.Context
			var cancel context.CancelFunc
			if s.timeout > 0 {
				ctx, cancel = context.WithTimeout(base, s.timeout)
			} else {
				ctx, cancel = context.WithCancel(base)
			}
			defer cancel()
			rec := &txRecorder{cancel: cancel}

			child := recordChild
			if s.childHandler != nil {
				child = s.childHandler
			}
			med := mediator.NewMediator()
			med.Register(typeA, s.rootHandler(rec))
			med.Register(typeB, child(rec))

			projectors := handle.NewRegistry()
			projectors.Register(&txProjector{failOnB: s.failProjectorOnB, rec: rec})

			middlewares := s.middlewares
			if middlewares == nil {
				middlewares = defaultMiddlewares()
			}
			eng := NewBFSEngine(NewExecutor(med)).WithMiddleware(middlewares...).WithProjector(projectors)
			if !s.noUnitOfWork {
				eng = eng.WithUnitOfWork(repos.NewUnitOfWork(db))
			}

			_, err := eng.Run(ctx, txRootEvent(s.rootVersion, s.rootMode))

			switch {
			case s.wantCode != "":
				expectEventErrorCode(err, s.wantCode)
			case s.wantErrIs != nil:
				Expect(errors.Is(err, s.wantErrIs)).To(BeTrue(), "err = %v", err)
			case s.wantErr:
				Expect(err).To(HaveOccurred())
			default:
				Expect(err).NotTo(HaveOccurred())
			}

			var events []txEventRow
			Expect(db.Select(&events, `SELECT aggregate_type, aggregate_id, aggregate_version, event_type
				FROM event_store ORDER BY event_id`)).To(Succeed())
			if s.wantEvents == nil {
				Expect(events).To(BeEmpty())
			} else {
				Expect(events).To(Equal(s.wantEvents))
			}

			var versions []txVersionRow
			Expect(db.Select(&versions, `SELECT aggregate_type, aggregate_id, current_version FROM aggregate_versions`)).To(Succeed())
			if s.wantVersions == nil {
				Expect(versions).To(BeEmpty())
			} else {
				Expect(versions).To(ConsistOf(s.wantVersions))
			}

			var checkpoints int
			Expect(db.Get(&checkpoints, `SELECT COUNT(*) FROM projection_checkpoints`)).To(Succeed())
			Expect(checkpoints).To(Equal(s.wantCheckpoints))

			if s.verify != nil {
				s.verify(rec, db)
			}
		})
	}
})
