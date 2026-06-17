package safety

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
)

// ─── safetyCall：表達有狀態 middleware 的單次呼叫 ──────────────────────────────

type safetyCall struct {
	batch    func() []event.Event
	wantCode kerrors.EventErrorCode // "" 表示不預期錯誤
}

// ─── DepthGuard ───────────────────────────────────────────────────────────────

type depthGuardScenario struct {
	given, when, then string
	maxDepth, depth   int
	wantCode          kerrors.EventErrorCode
}

var depthGuardScenarios = []depthGuardScenario{
	{
		given:    "depth 低於上限",
		when:     "DepthGuard 被呼叫",
		then:     "不回傳錯誤",
		maxDepth: 2, depth: 1,
	},
	{
		given:    "depth 等於上限",
		when:     "DepthGuard 被呼叫",
		then:     "不回傳錯誤（邊界值允許）",
		maxDepth: 2, depth: 2,
	},
	{
		given:    "depth 超過上限",
		when:     "DepthGuard 被呼叫",
		then:     "回傳 ErrBFSDepthExceeded",
		maxDepth: 2, depth: 3,
		wantCode: kerrors.ErrBFSDepthExceeded,
	},
}

// ─── CycleDetector ───────────────────────────────────────────────────────────

type cycleDetectorScenario struct {
	given, when, then string
	calls             []safetyCall
}

var cycleDetectorScenarios = []cycleDetectorScenario{
	{
		given: "兩個 level 的 event 各有唯一 UUID",
		when:  "CycleDetector 依序被呼叫兩次",
		then:  "不回傳錯誤",
		calls: []safetyCall{
			{batch: func() []event.Event { return []event.Event{mkEvt("a", typeA, ""), mkEvt("b", typeA, "")} }},
			{batch: func() []event.Event { return []event.Event{mkEvt("c", typeA, "")} }},
		},
	},
	{
		given: "同一 batch 中有兩個相同 UUID",
		when:  "CycleDetector 被呼叫",
		then:  "回傳 ErrEventLoopDetected",
		calls: []safetyCall{
			{
				batch:    func() []event.Event { return []event.Event{mkEvt("a", typeA, ""), mkEvt("a", typeA, "")} },
				wantCode: kerrors.ErrEventLoopDetected,
			},
		},
	},
	{
		given: "相同 UUID 跨兩個 level 出現",
		when:  "CycleDetector 依序被呼叫兩次",
		then:  "第二次呼叫回傳 ErrEventLoopDetected",
		calls: []safetyCall{
			{batch: func() []event.Event { return []event.Event{mkEvt("a", typeA, "")} }},
			{
				batch:    func() []event.Event { return []event.Event{mkEvt("a", typeA, "")} },
				wantCode: kerrors.ErrEventLoopDetected,
			},
		},
	},
}

// ─── DeterministicLoopDetector ───────────────────────────────────────────────

type deterministicLoopScenario struct {
	given, when, then string
	calls             []safetyCall
}

var deterministicLoopScenarios = []deterministicLoopScenario{
	{
		given: "event 皆無 causation chain",
		when:  "DeterministicLoopDetector 被呼叫",
		then:  "不回傳錯誤",
		calls: []safetyCall{
			{batch: func() []event.Event { return []event.Event{mkEvt("a", typeA, ""), mkEvt("b", typeB, "")} }},
		},
	},
	{
		given: "causation chain 上的 EventType 皆不同",
		when:  "DeterministicLoopDetector 依序被呼叫兩次",
		then:  "不回傳錯誤",
		calls: []safetyCall{
			{batch: func() []event.Event { return []event.Event{mkEvt("a", typeA, "")} }},
			{batch: func() []event.Event { return []event.Event{mkEvt("b", typeB, "a")} }}, // typeB caused by typeA
		},
	},
	{
		given: "直接 causation chain 出現相同 EventType（A → A）",
		when:  "DeterministicLoopDetector 依序被呼叫兩次",
		then:  "第二次呼叫回傳 ErrEventLoopDetected",
		calls: []safetyCall{
			{batch: func() []event.Event { return []event.Event{mkEvt("a", typeA, "")} }},
			{
				batch:    func() []event.Event { return []event.Event{mkEvt("b", typeA, "a")} },
				wantCode: kerrors.ErrEventLoopDetected,
			},
		},
	},
	{
		given: "三層間接 causation chain 出現相同 EventType（A → B → A）",
		when:  "DeterministicLoopDetector 依序被呼叫三次",
		then:  "第三次呼叫回傳 ErrEventLoopDetected",
		calls: []safetyCall{
			{batch: func() []event.Event { return []event.Event{mkEvt("a", typeA, "")} }},
			{batch: func() []event.Event { return []event.Event{mkEvt("b", typeB, "a")} }},
			{
				batch:    func() []event.Event { return []event.Event{mkEvt("c", typeA, "b")} },
				wantCode: kerrors.ErrEventLoopDetected,
			},
		},
	},
}

// ─── BackpressureScheduler ───────────────────────────────────────────────────

type backpressureScenario struct {
	given, when, then string
	maxPending        int
	batch             func() []event.Event
	terminal          func() BatchProcessor
	wantCode          kerrors.EventErrorCode
}

var backpressureScenarios = []backpressureScenario{
	{
		given:      "初始 batch 未超過上限",
		when:       "BackpressureScheduler 被呼叫",
		then:       "不回傳錯誤",
		maxPending: 3,
		batch:      func() []event.Event { return []event.Event{mkEvt("a", typeA, ""), mkEvt("b", typeA, "")} },
		terminal:   func() BatchProcessor { return noop },
	},
	{
		given:      "初始 batch 超過上限",
		when:       "BackpressureScheduler 被呼叫",
		then:       "回傳 ErrBackpressureActive",
		maxPending: 1,
		batch:      func() []event.Event { return []event.Event{mkEvt("a", typeA, ""), mkEvt("b", typeA, "")} },
		terminal:   func() BatchProcessor { return noop },
		wantCode:   kerrors.ErrBackpressureActive,
	},
	{
		given:      "handler 產生的 children 在上限內",
		when:       "BackpressureScheduler 被呼叫",
		then:       "不回傳錯誤",
		maxPending: 3,
		batch:      func() []event.Event { return []event.Event{mkEvt("a", typeA, "")} },
		terminal: func() BatchProcessor {
			return withChildren([]event.Event{mkEvt("x", typeB, ""), mkEvt("y", typeB, "")})
		},
	},
	{
		given:      "handler 產生的 children 超過上限",
		when:       "BackpressureScheduler 被呼叫",
		then:       "回傳 ErrQueueOverflow",
		maxPending: 2,
		batch:      func() []event.Event { return []event.Event{mkEvt("a", typeA, "")} },
		terminal: func() BatchProcessor {
			return withChildren([]event.Event{mkEvt("x", typeB, ""), mkEvt("y", typeB, ""), mkEvt("z", typeB, "")})
		},
		wantCode: kerrors.ErrQueueOverflow,
	},
}
