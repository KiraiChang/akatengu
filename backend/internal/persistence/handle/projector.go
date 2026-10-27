package handle

import (
	"akatengu/internal/kernel/errors"
	"akatengu/internal/persistence/repos"
	"context"
	"fmt"

	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/event"
)

type handle func(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error

// Projector updates a single read model in response to an event.
// EventTypes returns the set of event types this projector handles;
// Apply is only called when evt.EventType is in that set.
// Apply depends only on the event and its own read model, so the same
// projector can rebuild the read model from event_store.
// Reset deletes the rows of its read model that belong to the merchant in ctx,
// so a rebuild starts from nothing (BFS R3 D1); every table Apply writes must
// be cleared.
type Projector interface {
	Name() string
	EventTypes() []event_types.EventType
	Apply(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error
	Reset(ctx context.Context, tx *repos.DbTransaction) error
}

// Registry holds all registered projectors and dispatches events to matching ones.
type Registry struct {
	projectors []Projector
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(p Projector) {
	r.projectors = append(r.projectors, p)
}

// ApplyAll dispatches evt to every projector whose EventTypes includes evt.EventType.
func (r *Registry) ApplyAll(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	for _, p := range r.projectors {
		if evt.EventType.InValues(p.EventTypes()...) {
			if err := p.Apply(ctx, tx, evt); err != nil {
				return errors.NewRuntimeError(errors.ErrProjectorError, evt, fmt.Errorf("projector %s: %w", p.Name(), err))
			}
			if tx.Check != nil {
				if err := tx.Check.Upsert(ctx, p.Name(), evt.Id); err != nil {
					return errors.NewRuntimeError(errors.ErrCheckPointError, evt, fmt.Errorf("checkpoint %s: %w", p.Name(), err))
				}
			}
		}
	}
	return nil
}

// ResetAll resets every projector and deletes its checkpoint, for the
// merchant in ctx (BFS R3 D1, D6). Tables of different projectors may refer
// to each other; a rebuild defers the foreign key checks to its commit (D4).
func (r *Registry) ResetAll(ctx context.Context, tx *repos.DbTransaction) error {
	for _, p := range r.projectors {
		if err := p.Reset(ctx, tx); err != nil {
			return fmt.Errorf("reset projector %s: %w", p.Name(), err)
		}
		if tx.Check != nil {
			if err := tx.Check.Delete(ctx, p.Name()); err != nil {
				return fmt.Errorf("delete checkpoint %s: %w", p.Name(), err)
			}
		}
	}
	return nil
}
