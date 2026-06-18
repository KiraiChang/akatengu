package safety

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
)

// ─── safetyCall：表達有狀態 middleware 的單次呼叫 ──────────────────────────────

type safetyCall struct {
	evt      func() event.Event
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
		given: "依序處理三個唯一 UUID 的 event",
		when:  "CycleDetector 依序被呼叫三次",
		then:  "不回傳錯誤",
		calls: []safetyCall{
			{evt: func() event.Event { return mkEvt("a", typeA, "") }},
			{evt: func() event.Event { return mkEvt("b", typeA, "") }},
			{evt: func() event.Event { return mkEvt("c", typeA, "") }},
		},
	},
	{
		given: "相同 UUID 在第二次呼叫出現",
		when:  "CycleDetector 依序被呼叫兩次",
		then:  "第二次呼叫回傳 ErrEventLoopDetected",
		calls: []safetyCall{
			{evt: func() event.Event { return mkEvt("a", typeA, "") }},
			{
				evt:      func() event.Event { return mkEvt("a", typeA, "") },
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
		when:  "DeterministicLoopDetector 依序被呼叫兩次",
		then:  "不回傳錯誤",
		calls: []safetyCall{
			{evt: func() event.Event { return mkEvt("a", typeA, "") }},
			{evt: func() event.Event { return mkEvt("b", typeB, "") }},
		},
	},
	{
		given: "causation chain 上的 EventType 皆不同",
		when:  "DeterministicLoopDetector 依序被呼叫兩次",
		then:  "不回傳錯誤",
		calls: []safetyCall{
			{evt: func() event.Event { return mkEvt("a", typeA, "") }},
			{evt: func() event.Event { return mkEvt("b", typeB, "a") }}, // typeB caused by typeA
		},
	},
	{
		given: "直接 causation chain 出現相同 EventType（A → A）",
		when:  "DeterministicLoopDetector 依序被呼叫兩次",
		then:  "第二次呼叫回傳 ErrEventLoopDetected",
		calls: []safetyCall{
			{evt: func() event.Event { return mkEvt("a", typeA, "") }},
			{
				evt:      func() event.Event { return mkEvt("b", typeA, "a") },
				wantCode: kerrors.ErrEventLoopDetected,
			},
		},
	},
	{
		given: "三層間接 causation chain 出現相同 EventType（A → B → A）",
		when:  "DeterministicLoopDetector 依序被呼叫三次",
		then:  "第三次呼叫回傳 ErrEventLoopDetected",
		calls: []safetyCall{
			{evt: func() event.Event { return mkEvt("a", typeA, "") }},
			{evt: func() event.Event { return mkEvt("b", typeB, "a") }},
			{
				evt:      func() event.Event { return mkEvt("c", typeA, "b") },
				wantCode: kerrors.ErrEventLoopDetected,
			},
		},
	},
}

// ─── BackpressureScheduler ───────────────────────────────────────────────────

type backpressureScenario struct {
	given, when, then string
	maxPending        int
	evt               func() event.Event
	terminal          func() EventProcessor
	wantCode          kerrors.EventErrorCode
}

var backpressureScenarios = []backpressureScenario{
	{
		given:      "root event，handler 回傳空 children",
		when:       "BackpressureScheduler 被呼叫",
		then:       "不回傳錯誤",
		maxPending: 1,
		evt:        func() event.Event { return mkEvt("a", typeA, "") },
		terminal:   func() EventProcessor { return noop },
	},
	{
		given:      "root event，handler 回傳 2 children（在上限內）",
		when:       "BackpressureScheduler 被呼叫",
		then:       "不回傳錯誤（pending = 0 + 2 = 2 ≤ 3）",
		maxPending: 3,
		evt:        func() event.Event { return mkEvt("a", typeA, "") },
		terminal: func() EventProcessor {
			return withChildren([]event.Event{mkEvt("x", typeB, ""), mkEvt("y", typeB, "")})
		},
	},
	{
		given:      "root event，handler 回傳 3 children（超過上限）",
		when:       "BackpressureScheduler 被呼叫",
		then:       "回傳 ErrQueueOverflow（pending = 0 + 3 = 3 > 2）",
		maxPending: 2,
		evt:        func() event.Event { return mkEvt("a", typeA, "") },
		terminal: func() EventProcessor {
			return withChildren([]event.Event{mkEvt("x", typeB, ""), mkEvt("y", typeB, ""), mkEvt("z", typeB, "")})
		},
		wantCode: kerrors.ErrQueueOverflow,
	},
	{
		given:      "maxPending=0，root event 進入即觸發限制",
		when:       "BackpressureScheduler 被呼叫",
		then:       "回傳 ErrBackpressureActive（pending=1 > 0）",
		maxPending: 0,
		evt:        func() event.Event { return mkEvt("a", typeA, "") },
		terminal:   func() EventProcessor { return noop },
		wantCode:   kerrors.ErrBackpressureActive,
	},
}
