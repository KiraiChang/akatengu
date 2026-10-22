package event_test

import (
	domainevent "akatengu/internal/domain/event"
	"akatengu/internal/enums"
	"akatengu/internal/model/payload"
	"akatengu/internal/shared/utils/test"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/shopspring/decimal"
)

// Every field is set, so a JSON tag that differs between the copies shows up.
const (
	contractAccountID = "1101-90"
	contractParentID  = "1101"
	contractNote      = "契約測試備註"
	contractAccountNo = "1234"
	contractDay       = "15"
)

func contractCreditLimit() decimal.NullDecimal {
	return decimal.NewNullDecimal(decimal.RequireFromString("50000.5"))
}

func contractAccountCreate() (domainevent.AccountCreatePayload, payload.AccountCreatePayload) {
	parentID, note := contractParentID, contractNote
	cf := enums.CashFlowCategoryOperating.Enum()
	return domainevent.AccountCreatePayload{
			AccountId: contractAccountID, ParentId: &parentID, Name: "契約測試科目",
			Type: enums.AccountAsset.Enum(), NormalBalance: enums.BalanceDebit.Enum(), Currency: "TWD",
			IsSummary: true, IsActive: true, Note: &note, CashFlowCategory: &cf,
		}, payload.AccountCreatePayload{
			AccountId: contractAccountID, ParentId: &parentID, Name: "契約測試科目",
			Type: enums.AccountAsset.Enum(), NormalBalance: enums.BalanceDebit.Enum(), Currency: "TWD",
			IsSummary: true, IsActive: true, Note: &note, CashFlowCategory: &cf,
		}
}

func contractLedgerCreate() (domainevent.LedgerAccountCreatePayload, payload.LedgerAccountCreatePayload) {
	accountNo, billingDay, dueDay, note := contractAccountNo, contractDay, contractDay, contractNote
	return domainevent.LedgerAccountCreatePayload{
			AccountId: contractAccountID, Institution: "契約測試銀行", Name: "契約測試帳戶", AccountNo: &accountNo,
			Currency: "TWD", CreditLimit: contractCreditLimit(), BillingDay: &billingDay, DueDay: &dueDay,
			IsActive: true, Note: &note, Type: enums.LedgerAccountTypeCreditCard.Enum(),
		}, payload.LedgerAccountCreatePayload{
			AccountId: contractAccountID, Institution: "契約測試銀行", Name: "契約測試帳戶", AccountNo: &accountNo,
			Currency: "TWD", CreditLimit: contractCreditLimit(), BillingDay: &billingDay, DueDay: &dueDay,
			IsActive: true, Note: &note, Type: enums.LedgerAccountTypeCreditCard.Enum(),
		}
}

// TestAccountPayloads_LegacyJSON is the contract of BFS R2 D12: the legacy
// pipeline and replay read the payloads BFS writes, and the BFS projectors read
// the payloads the legacy path wrote.
func TestAccountPayloads_LegacyJSON(t *testing.T) {
	test.InitEnums()

	newCreate, oldCreate := contractAccountCreate()
	newLedger, oldLedger := contractLedgerCreate()

	t.Run("account.created", func(t *testing.T) {
		expectSameJSON(t, newCreate, oldCreate)
	})
	t.Run("account.updated", func(t *testing.T) {
		expectSameJSON(t, domainevent.AccountUpdatedPayload{
			AccountId: newCreate.AccountId, ParentId: newCreate.ParentId, Name: newCreate.Name,
			Type: newCreate.Type, NormalBalance: newCreate.NormalBalance, Currency: newCreate.Currency,
			IsSummary: newCreate.IsSummary, IsActive: newCreate.IsActive, Note: newCreate.Note,
			Version: 3, CashFlowCategory: newCreate.CashFlowCategory,
		}, payload.AccountUpdatedPayload{
			AccountId: oldCreate.AccountId, ParentId: oldCreate.ParentId, Name: oldCreate.Name,
			Type: oldCreate.Type, NormalBalance: oldCreate.NormalBalance, Currency: oldCreate.Currency,
			IsSummary: oldCreate.IsSummary, IsActive: oldCreate.IsActive, Note: oldCreate.Note,
			Version: 3, CashFlowCategory: oldCreate.CashFlowCategory,
		})
	})
	t.Run("account.bulk_imported", func(t *testing.T) {
		expectSameJSON(t,
			domainevent.AccountBulkImportedPayload{Accounts: []domainevent.AccountCreatePayload{newCreate}},
			payload.AccountBulkImportedPayload{Accounts: []payload.AccountCreatePayload{oldCreate}})
	})
	t.Run("ledger_account.created", func(t *testing.T) {
		expectSameJSON(t, newLedger, oldLedger)
	})
	t.Run("ledger_account.updated", func(t *testing.T) {
		expectSameJSON(t, domainevent.LedgerAccountUpdatedPayload{
			LedgerId: 7, AccountId: newLedger.AccountId, Institution: newLedger.Institution, Name: newLedger.Name,
			AccountNo: newLedger.AccountNo, Currency: newLedger.Currency, CreditLimit: newLedger.CreditLimit,
			BillingDay: newLedger.BillingDay, DueDay: newLedger.DueDay, IsActive: newLedger.IsActive,
			Note: newLedger.Note, Version: 2, Type: newLedger.Type,
		}, payload.LedgerAccountUpdatedPayload{
			LedgerId: 7, AccountId: oldLedger.AccountId, Institution: oldLedger.Institution, Name: oldLedger.Name,
			AccountNo: oldLedger.AccountNo, Currency: oldLedger.Currency, CreditLimit: oldLedger.CreditLimit,
			BillingDay: oldLedger.BillingDay, DueDay: oldLedger.DueDay, IsActive: oldLedger.IsActive,
			Note: oldLedger.Note, Version: 2, Type: oldLedger.Type,
		})
	})
}

// expectSameJSON checks that the BFS payload n and the legacy payload o of the
// same data serialize alike, and that each type decodes the other's JSON.
func expectSameJSON[N, O any](t *testing.T, n N, o O) {
	t.Helper()
	newJSON, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("marshal %T: %v", n, err)
	}
	oldJSON, err := json.Marshal(o)
	if err != nil {
		t.Fatalf("marshal %T: %v", o, err)
	}
	if string(newJSON) != string(oldJSON) {
		t.Fatalf("JSON differs:\n new %s\n old %s", newJSON, oldJSON)
	}

	var gotOld O
	if err := json.Unmarshal(newJSON, &gotOld); err != nil {
		t.Fatalf("legacy type decodes BFS JSON: %v", err)
	}
	if !reflect.DeepEqual(gotOld, o) {
		t.Errorf("legacy type decoded %#v, want %#v", gotOld, o)
	}
	var gotNew N
	if err := json.Unmarshal(oldJSON, &gotNew); err != nil {
		t.Fatalf("BFS type decodes legacy JSON: %v", err)
	}
	if !reflect.DeepEqual(gotNew, n) {
		t.Errorf("BFS type decoded %#v, want %#v", gotNew, n)
	}
}
