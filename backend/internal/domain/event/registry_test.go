package event_test

import (
	domainevent "akatengu/internal/domain/event"
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	kerrors "akatengu/internal/kernel/errors"
	kevent "akatengu/internal/kernel/event"
	"akatengu/internal/shared/utils/test"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

// cfTestPayload is an explicit cash flow category update used by the tests.
func cfTestPayload() domainevent.TransactionCFCategoryUpdatedPayload {
	return domainevent.TransactionCFCategoryUpdatedPayload{
		TxnUUID: "0192f3f0-0000-7000-8000-000000000001",
		Entries: []domainevent.TransactionCFEntryItem{
			{EntryUUID: "0192f3f0-0000-7000-8000-000000000011", CFCategory: enums.CashFlowCategoryInvesting.Enum()},
		},
	}
}

func TestLookup_CFCategoryUpdated(t *testing.T) {
	test.InitEnums()
	cfType := event_types.EventTransactionCFCategoryUpdated.Enum()

	spec, ok := domainevent.Lookup(cfType)
	if !ok {
		t.Fatalf("%s is not registered", cfType)
	}
	if spec.EventType != cfType {
		t.Errorf("EventType = %v, want %v", spec.EventType, cfType)
	}
	if spec.Version != domainevent.VersionValidate {
		t.Errorf("Version = %q, want %q", spec.Version, domainevent.VersionValidate)
	}
	if spec.Legacy != nil {
		t.Errorf("Legacy = %+v, want nil: the event is only written through its dedicated API", spec.Legacy)
	}

	aggType, aggID, err := spec.Aggregate(cfTestPayload())
	if err != nil {
		t.Fatalf("Aggregate: %v", err)
	}
	if aggType != enums.AggregateTransaction.Enum() || aggID != cfTestPayload().TxnUUID {
		t.Errorf("aggregate = (%v, %s), want (TRANSACTION, %s)", aggType, aggID, cfTestPayload().TxnUUID)
	}
}

func TestDecode(t *testing.T) {
	test.InitEnums()
	cfType := event_types.EventTransactionCFCategoryUpdated.Enum()
	cfJSON, err := json.Marshal(cfTestPayload())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	cases := []struct {
		name      string
		eventType event_types.EventType
		raw       string
		want      any                    // decoded payload; nil when an error is expected
		wantCode  kerrors.EventErrorCode // "" = no error
	}{
		{
			name:      "現金流分類修改的 payload 解碼成值型別",
			eventType: cfType,
			raw:       string(cfJSON),
			want:      cfTestPayload(),
		},
		{
			name:      "未登錄的事件類型",
			eventType: event_types.EventTransactionCreated.Enum(),
			raw:       `{}`,
			wantCode:  kerrors.ErrUnknownEventType,
		},
		{
			name:      "無法解碼的 JSON",
			eventType: cfType,
			raw:       `{"txn_uuid":`,
			wantCode:  kerrors.ErrDecodePayload,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := domainevent.Decode(tc.eventType, json.RawMessage(tc.raw))
			if tc.wantCode != "" {
				var ee kerrors.EventError
				if !errors.As(err, &ee) || ee.Code != tc.wantCode {
					t.Fatalf("err = %v, want code %s", err, tc.wantCode)
				}
				return
			}
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("payload = %#v (%T), want %#v (%T)", got, got, tc.want, tc.want)
			}
		})
	}
}

func TestSpecs_Consistent(t *testing.T) {
	test.InitEnums()

	seen := make(map[event_types.EventType]bool)
	for _, spec := range domainevent.Specs() {
		if seen[spec.EventType] {
			t.Errorf("%v is registered twice", spec.EventType)
		}
		seen[spec.EventType] = true

		decoded, err := spec.Decode(json.RawMessage(`{}`))
		if err != nil {
			t.Errorf("%v: decode {}: %v", spec.EventType, err)
			continue
		}
		p, ok := decoded.(kevent.Payload)
		if !ok {
			t.Errorf("%v: decoded %T is not a kernel event payload", spec.EventType, decoded)
			continue
		}
		if p.EventType() != spec.EventType {
			t.Errorf("%v: decoded payload reports %v", spec.EventType, p.EventType())
		}
	}
	if len(seen) == 0 {
		t.Fatal("no event type is registered")
	}
}
