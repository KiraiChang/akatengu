package engine

import (
	kerrors "akatengu/internal/kernel/errors"
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Executor ExecuteBatch", func() {
	ctx := context.Background()

	for _, s := range executorScenarios {
		s := s
		It(fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then), func() {
			children, err := s.setup().ExecuteBatch(ctx, s.batch())

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
