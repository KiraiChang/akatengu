package xlsxparser

import (
	"akatengu/internal/pkg/csvparser"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

// Parse reads an Excel (.xlsx) file from r using the given template and returns parsed rows.
// password may be empty for unencrypted files.
func Parse(r io.Reader, tmpl csvparser.Template, password string) ([]csvparser.Row, error) {
	opts := excelize.Options{}
	if password != "" {
		opts.Password = password
	}

	f, err := excelize.OpenReader(r, opts)
	if err != nil {
		if isPasswordError(err) {
			return nil, fmt.Errorf("incorrect password or file is encrypted")
		}
		return nil, fmt.Errorf("open excel file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("excel file has no sheets")
	}

	allRows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("read sheet %q: %w", sheets[0], err)
	}

	var rows []csvparser.Row
	for i, record := range allRows {
		lineNo := i + 1
		if lineNo <= tmpl.SkipRows {
			continue
		}
		row, err := parseRecord(record, tmpl, lineNo)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", lineNo, err)
		}
		if row == nil {
			continue
		}
		rows = append(rows, *row)
	}
	return rows, nil
}

func isPasswordError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "password") ||
		strings.Contains(msg, "incorrect") ||
		strings.Contains(msg, "decrypt")
}

func parseRecord(record []string, tmpl csvparser.Template, lineNo int) (*csvparser.Row, error) {
	maxIdx := max(tmpl.DateColumn, tmpl.DescriptionColumn)
	if tmpl.DebitColumn != nil && *tmpl.DebitColumn > maxIdx {
		maxIdx = *tmpl.DebitColumn
	}
	if tmpl.CreditColumn != nil && *tmpl.CreditColumn > maxIdx {
		maxIdx = *tmpl.CreditColumn
	}
	if tmpl.AmountColumn != nil && *tmpl.AmountColumn > maxIdx {
		maxIdx = *tmpl.AmountColumn
	}
	if tmpl.BalanceColumn != nil && *tmpl.BalanceColumn > maxIdx {
		maxIdx = *tmpl.BalanceColumn
	}
	if tmpl.ReferenceColumn != nil && *tmpl.ReferenceColumn > maxIdx {
		maxIdx = *tmpl.ReferenceColumn
	}
	if len(record) <= maxIdx {
		return nil, nil
	}

	rawDate := strings.TrimSpace(record[tmpl.DateColumn])
	if rawDate == "" {
		return nil, nil
	}
	t, err := time.Parse(tmpl.DateFormat, rawDate)
	if err != nil {
		return nil, fmt.Errorf("date parse %q with format %q: %w", rawDate, tmpl.DateFormat, err)
	}
	txnDate := t.Format("2006-01-02")

	description := strings.TrimSpace(record[tmpl.DescriptionColumn])

	var debit, credit decimal.Decimal
	switch {
	case tmpl.DebitColumn != nil && tmpl.CreditColumn != nil:
		debit = parseAmount(record[*tmpl.DebitColumn])
		credit = parseAmount(record[*tmpl.CreditColumn])
	case tmpl.AmountColumn != nil:
		amt := parseAmount(record[*tmpl.AmountColumn])
		if amt.IsNegative() {
			debit = amt.Neg()
		} else {
			credit = amt
		}
	default:
		return nil, fmt.Errorf("template must have either (debit_column + credit_column) or amount_column")
	}

	if debit.IsZero() && credit.IsZero() {
		return nil, nil
	}

	var balance *decimal.Decimal
	if tmpl.BalanceColumn != nil {
		raw := strings.TrimSpace(safeGet(record, *tmpl.BalanceColumn))
		if raw != "" {
			v := parseAmount(raw)
			balance = &v
		}
	}

	var referenceNo *string
	if tmpl.ReferenceColumn != nil {
		raw := strings.TrimSpace(safeGet(record, *tmpl.ReferenceColumn))
		if raw != "" {
			referenceNo = &raw
		}
	}

	return &csvparser.Row{
		TxnDate:     txnDate,
		Description: description,
		Debit:       debit,
		Credit:      credit,
		Balance:     balance,
		ReferenceNo: referenceNo,
	}, nil
}

// safeGet returns record[i] or "" if i is out of bounds.
// excelize trims trailing empty cells, so optional columns may be absent.
func safeGet(record []string, i int) string {
	if i < len(record) {
		return record[i]
	}
	return ""
}

func parseAmount(s string) decimal.Decimal {
	s = strings.TrimSpace(s)
	negative := false
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		negative = true
		s = s[1 : len(s)-1]
	}
	s = strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) || r == '.' || r == '-' {
			return r
		}
		return -1
	}, s)
	if s == "" || s == "-" {
		return decimal.Zero
	}
	v, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	if negative {
		return v.Neg()
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
