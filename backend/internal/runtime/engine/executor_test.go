package engine

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"akatengu/internal/kernel/result"
	"akatengu/internal/repos/query"
	"akatengu/internal/runtime/mediator"
	"context"
	"errors"
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Executor ExecuteBatch", func() {
	ctx := context.Background()

	for _, s := range executorScenarios {
		s := s
		It(fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then), func() {
			children, err := s.setup().ExecuteBatch(ctx, nil, s.batch())

			if s.wantCode != "" {
				var ee kerrors.EventError
				Expect(errors.As(err, &ee)).To(BeTrue(), "expected EventError, got %T: %v", err, err)
				Expect(ee.Code).To(Equal(s.wantCode))
			} else {
				Expect(err).To(BeNil())
				Expect(childUUIDs(children)).To(Equal(s.wantUUIDs))
			}
		})
	}
})

var _ = Describe("Executor 轉傳 Repo", func() {
	for _, s := range executorRepoScenarios {
		s := s
		It(fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then), func() {
			repo := &query.Repo{}
			var mu sync.Mutex
			var got []*query.Repo
			med := mediator.NewMediator()
			med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, r *query.Repo, _ event.Event) (result.EventResult, error) {
				mu.Lock()
				defer mu.Unlock()
				got = append(got, r)
				return result.EventResult{}, nil
			}))

			Expect(s.dispatch(NewExecutor(med), repo)).To(Succeed())

			Expect(got).To(HaveLen(s.wantCalls))
			for _, r := range got {
				Expect(r).To(BeIdenticalTo(repo))
			}
		})
	}
})
