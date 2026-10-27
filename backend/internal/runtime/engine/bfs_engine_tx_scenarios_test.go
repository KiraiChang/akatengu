package engine

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"akatengu/internal/pkg/uuidx"
	"akatengu/internal/runtime/mediator"
	"akatengu/internal/runtime/safety"
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

// Fixture identifiers for the engine scenarios that run against a DB. Event
// types are written as strings because typeA/typeB are only set in BeforeSuite.
const (
	txRootUUID      = "0192f3d0-0000-7000-8000-000000000001"
	txRootAggID     = "0192f3d0-0000-7000-8000-0000000000a1"
	txMerchantID    = int64(1)
	txUserName      = "r1-tester"
	txCorrelationID = "0192f3d0-0000-7000-8000-0000000000c1"
	txRequestID     = "r1-request"
	txPurpose       = "r1:derived"
	txAggType       = "TRANSACTION"
	txTypeA         = "transaction.created"
	txTypeB         = "transaction.corrected"
	txEnrichedNote  = "r3-enriched"
)

// txDerivedAggID is the aggregate NewDerivedEvent creates from the root event.
var txDerivedAggID = uuidx.NewFromEvent(txRootUUID, txPurpose)

// txEventRow is a row of event_store, in event_id order.
type txEventRow struct {
	AggregateType string `db:"aggregate_type"`
	AggregateID   string `db:"aggregate_id"`
	Version       int64  `db:"aggregate_version"`
	EventType     string `db:"event_type"`
}

// txVersionRow is a row of aggregate_versions.
type txVersionRow struct {
	AggregateType string `db:"aggregate_type"`
	AggregateID   string `db:"aggregate_id"`
	Version       int64  `db:"current_version"`
}

type txScenario struct {
	given, when, then string
	middlewares       []safety.Middleware // nil = defaultMiddlewares()
	noUnitOfWork      bool                // run without a DB transaction
	seedVersions      []txVersionRow      // aggregate_versions rows written before Run
	rootVersion       int64               // ExpectedVersion of the root event
	rootMode          event.RootVersion   // version mode of the root event; zero = RootVersionExpected
	timeout           time.Duration       // deadline of the Run context; 0 = none
	rootHandler       func(rec *txRecorder) mediator.HandlerFunc
	childHandler      func(rec *txRecorder) mediator.HandlerFunc // nil = recordChild
	failProjectorOnB  bool                                       // the projector fails on typeB events

	wantCode        kerrors.EventErrorCode // EventError code of the Run error
	wantErrIs       error                  // errors.Is target of the Run error
	wantErr         bool                   // any Run error, when its kind is not fixed
	wantEvents      []txEventRow           // event_store after Run; nil = empty
	wantVersions    []txVersionRow         // aggregate_versions after Run, any order; nil = empty
	wantCheckpoints int                    // rows in projection_checkpoints after Run
	verify          func(rec *txRecorder, db *sqlx.DB)
}

// txRootV1 is the root event written as version 1 of the root aggregate.
var txRootV1 = txEventRow{txAggType, txRootAggID, 1, txTypeA}

// txRootAt is the root event written as version v of the root aggregate.
func txRootAt(v int64) txEventRow { return txEventRow{txAggType, txRootAggID, v, txTypeA} }

// txSameEvent is a typeB event at version v of the root aggregate.
func txSameEvent(v int64) txEventRow { return txEventRow{txAggType, txRootAggID, v, txTypeB} }

// txDerivedEvent is a typeB event at version v of the derived aggregate.
func txDerivedEvent(v int64) txEventRow { return txEventRow{txAggType, txDerivedAggID, v, txTypeB} }

func txRootVersion(v int64) txVersionRow { return txVersionRow{txAggType, txRootAggID, v} }

func txDerivedVersion(v int64) txVersionRow { return txVersionRow{txAggType, txDerivedAggID, v} }

var txScenarios = []txScenario{
	{
		given:           "單一根事件，handler 與 projector 成功",
		when:            "Run",
		then:            "寫入版本 1 的事件與 1 筆 checkpoint",
		rootHandler:     emitSame(0),
		wantEvents:      []txEventRow{txRootV1},
		wantVersions:    []txVersionRow{txRootVersion(1)},
		wantCheckpoints: 1,
	},
	{
		given:           "根事件產生一個同 aggregate 的子事件",
		when:            "Run",
		then:            "同一個 aggregate 有版本 1、2 的事件",
		rootHandler:     emitSame(1),
		wantEvents:      []txEventRow{txRootV1, txSameEvent(2)},
		wantVersions:    []txVersionRow{txRootVersion(2)},
		wantCheckpoints: 1,
	},
	{
		given:           "根事件產生兩個同 aggregate 的子事件（兄弟）",
		when:            "Run",
		then:            "版本依序為 1、2、3",
		rootHandler:     emitSame(2),
		wantEvents:      []txEventRow{txRootV1, txSameEvent(2), txSameEvent(3)},
		wantVersions:    []txVersionRow{txRootVersion(3)},
		wantCheckpoints: 1,
	},
	{
		given:           "根事件 → 子事件 → 孫事件，都在同一個 aggregate",
		when:            "Run",
		then:            "版本依序為 1、2、3",
		rootHandler:     emitSame(1),
		childHandler:    emitGrandchild,
		wantEvents:      []txEventRow{txRootV1, txSameEvent(2), txSameEvent(3)},
		wantVersions:    []txVersionRow{txRootVersion(3)},
		wantCheckpoints: 1,
	},
	{
		given:           "根事件以 NewDerivedEvent 產生衍生 aggregate 的子事件",
		when:            "Run",
		then:            "子事件建立衍生的 aggregate，版本 1；根事件的 aggregate 版本 1",
		rootHandler:     emitDerived(txPurpose),
		wantEvents:      []txEventRow{txRootV1, txDerivedEvent(1)},
		wantVersions:    []txVersionRow{txRootVersion(1), txDerivedVersion(1)},
		wantCheckpoints: 1,
	},
	{
		given:           "衍生 aggregate 的子事件再產生同 aggregate 的孫事件",
		when:            "Run",
		then:            "衍生 aggregate 有版本 1、2 的事件；根事件的 aggregate 版本 1",
		rootHandler:     emitDerived(txPurpose),
		childHandler:    emitGrandchild,
		wantEvents:      []txEventRow{txRootV1, txDerivedEvent(1), txDerivedEvent(2)},
		wantVersions:    []txVersionRow{txRootVersion(1), txDerivedVersion(2)},
		wantCheckpoints: 1,
	},
	{
		given:        "衍生的 aggregate 已有版本 0 的紀錄",
		when:         "根事件產生衍生 aggregate 的子事件",
		then:         "回傳 ErrVersionConflict，全部回復，版本紀錄仍為 0",
		seedVersions: []txVersionRow{txDerivedVersion(0)},
		rootHandler:  emitDerived(txPurpose),
		wantCode:     kerrors.ErrVersionConflict,
		wantVersions: []txVersionRow{txDerivedVersion(0)},
	},
	{
		given:        "衍生的 aggregate 已有版本 3 的紀錄",
		when:         "根事件產生衍生 aggregate 的子事件",
		then:         "回傳 ErrVersionConflict，全部回復，版本紀錄仍為 3",
		seedVersions: []txVersionRow{txDerivedVersion(3)},
		rootHandler:  emitDerived(txPurpose),
		wantCode:     kerrors.ErrVersionConflict,
		wantVersions: []txVersionRow{txDerivedVersion(3)},
	},
	{
		given:       "同一次 Run 兩個子事件衍生出相同的 aggregate",
		when:        "Run",
		then:        "第二個回傳 ErrVersionConflict，全部回復",
		rootHandler: emitDerived(txPurpose, txPurpose),
		wantCode:    kerrors.ErrVersionConflict,
	},
	{
		given:        "子事件的 handler 回傳錯誤",
		when:         "Run",
		then:         "Run 回傳該錯誤，根事件與它的 checkpoint 也回復",
		rootHandler:  emitSame(1),
		childHandler: failChild,
		wantErrIs:    errTxChild,
	},
	{
		given:            "子事件的 projector 回傳錯誤",
		when:             "Run",
		then:             "回傳 ErrProjectorError，全部回復",
		rootHandler:      emitSame(1),
		failProjectorOnB: true,
		wantCode:         kerrors.ErrProjectorError,
	},
	{
		given:       "DepthGuard(0)",
		when:        "根事件產生子事件",
		then:        "子事件被擋下，回傳 ErrBFSDepthExceeded，全部回復",
		middlewares: []safety.Middleware{safety.DepthGuard(0)},
		rootHandler: emitSame(1),
		wantCode:    kerrors.ErrBFSDepthExceeded,
	},
	{
		given:       "BackpressureScheduler(1)",
		when:        "根事件寫入並套用 projector 後，才因產生兩個子事件被判定超量",
		then:        "回傳 ErrQueueOverflow，全部回復",
		middlewares: []safety.Middleware{safety.BackpressureScheduler(1), safety.DepthGuard(10), safety.CycleDetector()},
		rootHandler: emitSame(2),
		wantCode:    kerrors.ErrQueueOverflow,
	},
	{
		given:       "根事件的 handler 取消 ctx 後回傳子事件",
		when:        "Run",
		then:        "Run 回傳錯誤，子事件的 handler 沒有執行，全部回復",
		rootHandler: cancelThenEmit,
		wantErr:     true,
		verify:      expectChildCalls(0),
	},
	{
		given:        "沒有設定 UnitOfWork，根事件的 handler 取消 ctx 後回傳子事件",
		when:         "Run",
		then:         "Run 回傳 context.Canceled，子事件的 handler 沒有執行",
		noUnitOfWork: true,
		rootHandler:  cancelThenEmit,
		wantErrIs:    context.Canceled,
		verify:       expectChildCalls(0),
	},
	{
		given:           "子事件的 handler 以收到的 Repo 查詢根事件，ctx 1 秒逾時",
		when:            "Run",
		then:            "查得到尚未 commit 的根事件，Run 成功",
		timeout:         time.Second,
		rootHandler:     emitSame(1),
		childHandler:    queryRoot,
		wantEvents:      []txEventRow{txRootV1, txSameEvent(2)},
		wantVersions:    []txVersionRow{txRootVersion(2)},
		wantCheckpoints: 1,
		verify:          expectRootFound,
	},
	{
		given:           "根事件的 handler 在子事件填入其他 MerchantID、UpdatedBy、CorrelationID、RequestID",
		when:            "Run",
		then:            "子事件的封套欄位仍與父事件相同，CausationID 為父事件 UUID",
		rootHandler:     emitWithOtherEnvelope,
		wantEvents:      []txEventRow{txRootV1, txSameEvent(2)},
		wantVersions:    []txVersionRow{txRootVersion(2)},
		wantCheckpoints: 1,
		verify:          expectChildEnvelope,
	},
	{
		given:       "子事件直接填入與父事件不同的 aggregate，沒有使用 NewDerivedEvent",
		when:        "Run",
		then:        "回傳 ErrChildEnvelopeInvalid，全部回復",
		rootHandler: emitOtherAggregate,
		wantCode:    kerrors.ErrChildEnvelopeInvalid,
	},
	{
		given:        "根事件的 aggregate 已在版本 1",
		when:         "根事件以過期的 ExpectedVersion 0 送出",
		then:         "Run 失敗，event_store 不變（既有行為：新增版本紀錄時撞上主鍵）",
		seedVersions: []txVersionRow{txRootVersion(1)},
		rootHandler:  emitSame(0),
		wantCode:     kerrors.ErrStoreWriteFailed,
		wantVersions: []txVersionRow{txRootVersion(1)},
	},
	{
		given:           "根事件為 RootVersionNew，aggregate 不存在",
		when:            "Run",
		then:            "建立 aggregate，寫入版本 1",
		rootMode:        event.RootVersionNew,
		rootHandler:     emitSame(0),
		wantEvents:      []txEventRow{txRootV1},
		wantVersions:    []txVersionRow{txRootVersion(1)},
		wantCheckpoints: 1,
	},
	{
		given:        "根事件為 RootVersionNew，aggregate 已有版本 0 的紀錄",
		when:         "Run",
		then:         "回傳 ErrVersionConflict，不寫入事件，版本紀錄仍為 0",
		seedVersions: []txVersionRow{txRootVersion(0)},
		rootMode:     event.RootVersionNew,
		rootHandler:  emitSame(0),
		wantCode:     kerrors.ErrVersionConflict,
		wantVersions: []txVersionRow{txRootVersion(0)},
	},
	{
		given:        "根事件為 RootVersionNew，aggregate 已在版本 2",
		when:         "Run",
		then:         "回傳 ErrVersionConflict，不寫入事件，版本紀錄仍為 2",
		seedVersions: []txVersionRow{txRootVersion(2)},
		rootMode:     event.RootVersionNew,
		rootHandler:  emitSame(0),
		wantCode:     kerrors.ErrVersionConflict,
		wantVersions: []txVersionRow{txRootVersion(2)},
	},
	{
		given:           "根事件為 RootVersionCurrent，aggregate 不存在",
		when:            "根事件以 ExpectedVersion 5 送出",
		then:            "忽略 ExpectedVersion，建立 aggregate，寫入版本 1",
		rootMode:        event.RootVersionCurrent,
		rootVersion:     5,
		rootHandler:     emitSame(0),
		wantEvents:      []txEventRow{txRootV1},
		wantVersions:    []txVersionRow{txRootVersion(1)},
		wantCheckpoints: 1,
	},
	{
		given:           "根事件為 RootVersionCurrent，aggregate 已在版本 3",
		when:            "根事件以 ExpectedVersion 1 送出，並產生一個同 aggregate 的子事件",
		then:            "忽略 ExpectedVersion，根事件寫入版本 4，子事件接著寫入版本 5",
		seedVersions:    []txVersionRow{txRootVersion(3)},
		rootMode:        event.RootVersionCurrent,
		rootVersion:     1,
		rootHandler:     emitSame(1),
		wantEvents:      []txEventRow{txRootAt(4), txSameEvent(5)},
		wantVersions:    []txVersionRow{txRootVersion(5)},
		wantCheckpoints: 1,
	},
	{
		given:           "根事件的 handler 回傳補充後的 payload（同型別），並產生一個同 aggregate 的子事件",
		when:            "Run",
		then:            "event_store 與 projector 都取得根事件補充後的 payload；子事件的 payload 不受影響",
		rootHandler:     enrichRoot,
		wantEvents:      []txEventRow{txRootV1, txSameEvent(2)},
		wantVersions:    []txVersionRow{txRootVersion(2)},
		wantCheckpoints: 1,
		verify:          expectEnrichedRoot,
	},
	{
		given:       "根事件的 handler 回傳不同型別的 payload",
		when:        "Run",
		then:        "回傳 ErrResultPayloadInvalid，不寫入事件與版本紀錄",
		rootHandler: enrichWithOtherType,
		wantCode:    kerrors.ErrResultPayloadInvalid,
	},
}
