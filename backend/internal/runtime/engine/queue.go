package engine

import "akatengu/internal/kernel/event"

type Queue struct {
	batches [][]event.Event
}

func (q *Queue) PushBatch(events []event.Event) {
	if len(events) == 0 {
		return
	}
	q.batches = append(q.batches, events)
}

func (q *Queue) PopCurrentLevel() []event.Event {
	if len(q.batches) == 0 {
		return nil
	}
	batch := q.batches[0]
	q.batches = q.batches[1:]
	return batch
}

func (q *Queue) Empty() bool {
	return len(q.batches) == 0
}
