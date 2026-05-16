package query_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"

	"akatengu/internal/testutil"
)

type closureRow struct {
	AncestorId   string `db:"ancestor_id"`
	DescendantId string `db:"descendant_id"`
	Depth        int    `db:"depth"`
}

func queryClosure(t *testing.T, db *sqlx.DB, descendantId string) []closureRow {
	t.Helper()
	var rows []closureRow
	if err := db.SelectContext(context.Background(), &rows,
		`SELECT ancestor_id, descendant_id, depth FROM account_closure WHERE descendant_id = ? ORDER BY depth`,
		descendantId); err != nil {
		t.Fatalf("queryClosure %s: %v", descendantId, err)
	}
	return rows
}

func assertClosureRow(t *testing.T, rows []closureRow, ancestor, descendant string, depth int) {
	t.Helper()
	for _, r := range rows {
		if r.AncestorId == ancestor && r.DescendantId == descendant && r.Depth == depth {
			return
		}
	}
	t.Errorf("closure row not found: ancestor=%s descendant=%s depth=%d (got %v)", ancestor, descendant, depth, rows)
}

func closureInsertAccount(t *testing.T, db *sqlx.DB, accountId string, parentId *string, isSummary int) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO accounts (account_id, parent_id, name, type, normal_balance, currency, is_summary, is_active, version)
		 VALUES (?, ?, '測試科目', 'ASSET', 'DEBIT', 'TWD', ?, 1, 1)`,
		accountId, parentId, isSummary)
	if err != nil {
		t.Fatalf("closureInsertAccount %s: %v", accountId, err)
	}
}

// TestAccountClosure_Migration_SeededAccounts 驗證 seed 帳戶經 trigger 正確填充 closure 表。
func TestAccountClosure_Migration_SeededAccounts(t *testing.T) {
	db := testutil.NewTestDB(t)

	// 1101-01（葉）：self(0)、1101(1)、110(2)
	rows := queryClosure(t, db, "1101-01")
	assertClosureRow(t, rows, "1101-01", "1101-01", 0)
	assertClosureRow(t, rows, "1101", "1101-01", 1)
	assertClosureRow(t, rows, "110", "1101-01", 2)

	// 1101（直接父科目）：self(0)、110(1)
	rows = queryClosure(t, db, "1101")
	assertClosureRow(t, rows, "1101", "1101", 0)
	assertClosureRow(t, rows, "110", "1101", 1)

	// 110（祖父科目）：只有 self
	rows = queryClosure(t, db, "110")
	assertClosureRow(t, rows, "110", "110", 0)
	if len(rows) != 1 {
		t.Errorf("110 should have exactly 1 closure row, got %d", len(rows))
	}
}

// TestAccountClosure_Trigger_RootAccount 根科目（無父科目）INSERT 後 closure 只有 self。
func TestAccountClosure_Trigger_RootAccount(t *testing.T) {
	db := testutil.NewTestDB(t)

	closureInsertAccount(t, db, "9000", nil, 1)

	rows := queryClosure(t, db, "9000")
	if len(rows) != 1 {
		t.Fatalf("root account should have 1 closure row, got %d: %v", len(rows), rows)
	}
	assertClosureRow(t, rows, "9000", "9000", 0)
}

// TestAccountClosure_Trigger_DirectChild 子科目 INSERT 後有 self(0) 及 parent→child(1)。
func TestAccountClosure_Trigger_DirectChild(t *testing.T) {
	db := testutil.NewTestDB(t)

	parent := "9000"
	closureInsertAccount(t, db, parent, nil, 1)
	closureInsertAccount(t, db, "9001", &parent, 0)

	rows := queryClosure(t, db, "9001")
	if len(rows) != 2 {
		t.Fatalf("child account should have 2 closure rows, got %d: %v", len(rows), rows)
	}
	assertClosureRow(t, rows, "9001", "9001", 0)
	assertClosureRow(t, rows, "9000", "9001", 1)
}

// TestAccountClosure_Trigger_Grandchild 三層後孫科目有 self(0)、parent→gc(1)、grandparent→gc(2)。
func TestAccountClosure_Trigger_Grandchild(t *testing.T) {
	db := testutil.NewTestDB(t)

	p0 := (*string)(nil)
	p1 := func(s string) *string { return &s }

	closureInsertAccount(t, db, "9000", p0, 1)
	closureInsertAccount(t, db, "9001", p1("9000"), 1)
	closureInsertAccount(t, db, "9002", p1("9001"), 0)

	rows := queryClosure(t, db, "9002")
	if len(rows) != 3 {
		t.Fatalf("grandchild should have 3 closure rows, got %d: %v", len(rows), rows)
	}
	assertClosureRow(t, rows, "9002", "9002", 0)
	assertClosureRow(t, rows, "9001", "9002", 1)
	assertClosureRow(t, rows, "9000", "9002", 2)
}
