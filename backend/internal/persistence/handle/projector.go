package handle

import (
	"akatengu/internal/kernel/errors"
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/event"
)

type handle func(ctx context.Context, db *sqlx.DB, evt event.Event, state any) error

// Projector updates a single read model in response to an event.
// EventTypes returns the set of event types this projector handles;
// Apply is only called when evt.EventType is in that set.
type Projector interface {
	Name() string
	EventTypes() []event_types.EventType
	Apply(ctx context.Context, db *sqlx.DB, evt event.Event, state any) error
}

// Registry holds all registered projectors and dispatches events to matching ones.
type Registry struct {
	db         *sqlx.DB
	projectors []Projector
}

func NewRegistry(db *sqlx.DB) *Registry {
	return &Registry{db: db}
}

func (r *Registry) Register(p Projector) {
	r.projectors = append(r.projectors, p)
}

// ApplyAll dispatches evt to every projector whose EventTypes includes evt.EventType.
func (r *Registry) ApplyAll(ctx context.Context, evt event.Event, state any) error {
	for _, p := range r.projectors {
		if evt.EventType.InValues(p.EventTypes()...) {
			if err := p.Apply(ctx, r.db, evt, state); err != nil {
				return errors.NewRuntimeError(errors.ErrProjectorError, evt, fmt.Errorf("projector %s: %w", p.Name(), err))
			}
		}
	}
	return nil
}
