package engine

import "akatengu/internal/kernel/event"

type queueItem struct {
	evt   event.Event
	depth int
}

type Queue struct {
	items []queueItem
}

func (q *Queue) Push(evt event.Event, depth int) {
	q.items = append(q.items, queueItem{evt: evt, depth: depth})
}

func (q *Queue) Pop() (event.Event, int) {
	item := q.items[0]
	q.items = q.items[1:]
	return item.evt, item.depth
}

func (q *Queue) Empty() bool {
	return len(q.items) == 0
}
