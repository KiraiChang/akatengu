package safety

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func extractCode(err error) (kerrors.EventErrorCode, bool) {
	var ee kerrors.EventError
	return ee.Code, errors.As(err, &ee)
}

var _ = Describe("Safety Middlewares", func() {

	// ─── DepthGuard ──────────────────────────────────────────────────────────

	Describe("DepthGuard", func() {
		for _, s := range depthGuardScenarios {
			s := s
			It(fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then), func() {
				_, err := DepthGuard(s.maxDepth)(noop)(ctx, s.depth, event.Event{})
				if s.wantCode == "" {
					Expect(err).To(BeNil())
				} else {
					code, ok := extractCode(err)
					Expect(ok).To(BeTrue(), "expected EventError, got %T", err)
					Expect(code).To(Equal(s.wantCode))
				}
			})
		}
	})

	// ─── CycleDetector ───────────────────────────────────────────────────────

	Describe("CycleDetector", func() {
		for _, s := range cycleDetectorScenarios {
			s := s
			It(fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then), func() {
				p := CycleDetector()(noop)
				for i, c := range s.calls {
					_, err := p(ctx, i, c.evt())
					if c.wantCode == "" {
						Expect(err).To(BeNil())
					} else {
						code, ok := extractCode(err)
						Expect(ok).To(BeTrue(), "expected EventError, got %T", err)
						Expect(code).To(Equal(c.wantCode))
						return
					}
				}
			})
		}
	})

	// ─── DeterministicLoopDetector ───────────────────────────────────────────

	Describe("DeterministicLoopDetector", func() {
		for _, s := range deterministicLoopScenarios {
			s := s
			It(fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then), func() {
				p := DeterministicLoopDetector()(noop)
				for i, c := range s.calls {
					_, err := p(ctx, i, c.evt())
					if c.wantCode == "" {
						Expect(err).To(BeNil())
					} else {
						code, ok := extractCode(err)
						Expect(ok).To(BeTrue(), "expected EventError, got %T", err)
						Expect(code).To(Equal(c.wantCode))
						return
					}
				}
			})
		}
	})

	// ─── BackpressureScheduler ───────────────────────────────────────────────

	Describe("BackpressureScheduler", func() {
		for _, s := range backpressureScenarios {
			s := s
			It(fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then), func() {
				_, err := BackpressureScheduler(s.maxPending)(s.terminal())(ctx, 0, s.evt())
				if s.wantCode == "" {
					Expect(err).To(BeNil())
				} else {
					code, ok := extractCode(err)
					Expect(ok).To(BeTrue(), "expected EventError, got %T", err)
					Expect(code).To(Equal(s.wantCode))
				}
			})
		}
	})
})
