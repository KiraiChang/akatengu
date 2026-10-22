// Package command writes events through the BFS engine for the handlers, the
// dedicated APIs and the seeds (BFS R2 D4). It may depend on domain/event, the
// engine, process, persistence and query; none of them may depend on it.
package command

import (
	"akatengu/internal/persistence/handle"
	"akatengu/internal/persistence/repos"
	"akatengu/internal/process"
	"akatengu/internal/runtime/engine"
	"akatengu/internal/runtime/mediator"
	"akatengu/internal/runtime/safety"

	"github.com/jmoiron/sqlx"
)

// NewEngine assembles the BFS engine shared by every write path: the mediator
// handlers, the projectors and the middlewares, with a unit of work on db.
func NewEngine(db *sqlx.DB) *engine.BFSEngine {
	m := mediator.NewMediator()
	process.Registry(m)

	projector := handle.NewRegistry()
	projector.Register(handle.NewCashFlowCategoryProjector(db))
	projector.Register(handle.NewAccountProjector(db))
	projector.Register(handle.NewConfigProjector(db))
	projector.Register(handle.NewPrepaidCategoryProjector(db))
	projector.Register(handle.NewFixedAssetCategoryProjector(db))
	projector.Register(handle.NewPeriodProjector(db))

	return engine.NewBFSEngine(engine.NewExecutor(m)).
		WithMiddleware(safety.BackpressureScheduler(3), safety.DepthGuard(10), safety.CycleDetector()).
		WithProjector(projector).
		WithUnitOfWork(repos.NewUnitOfWork(db))
}
