package projection

import (
	dbprojection "akatengu/internal/model/db/projection"
	"context"
)

// InsertPeriodClose runs the same sqlc query as the legacy
// projection_repo/period.go; the opening projector does not use the new
// closing_id.
func (r *sqlxProjection) InsertPeriodClose(ctx context.Context, p dbprojection.PeriodClosing) error {
	_, err := r.q.InsertPeriodClose(ctx, p.ToInsertPeriodCloseParams())
	return err
}
