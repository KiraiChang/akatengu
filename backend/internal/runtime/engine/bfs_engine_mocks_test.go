package engine

import (
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/event"
	"akatengu/internal/persistence/repos"
	"context"
)

type mockStore struct {
	appendCount int
	err         error
}

func (m *mockStore) Append(_ context.Context, _ event.Event) (int64, int64, error) {
	m.appendCount++
	return 0, 0, m.err
}

type mockUoW struct {
	store *mockStore
	err   error
}

func (m *mockUoW) Do(ctx context.Context, fn func(*repos.DbTransaction) error) error {
	if m.err != nil {
		return m.err
	}
	return fn(&repos.DbTransaction{Store: m.store})
}

type mockProjector struct {
	types      []event_types.EventType
	applyCount int
	applyErr   error
}

func (m *mockProjector) Name() string                        { return "mock" }
func (m *mockProjector) EventTypes() []event_types.EventType { return m.types }
func (m *mockProjector) Apply(_ context.Context, _ *repos.DbTransaction, _ event.Event, _ any) error {
	m.applyCount++
	return m.applyErr
}
