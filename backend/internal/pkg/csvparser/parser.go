package csvparser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode"

	"github.com/shopspring/decimal"
)

// Template holds the column mapping configuration for a bank CSV file.
type Template struct {
	SkipRows          int
	DateColumn        int
	DateFormat        string
	DescriptionColumn int
	DebitColumn       *int   // optional: withdrawal column
	CreditColumn      *int   // optional: deposit column
	AmountColumn      *int   // optional: single amount column (negative = debit)
	BalanceColumn     *int   // optional
	ReferenceColumn   *int   // optional
}

// Row is a parsed bank transaction row.
type Row struct {
	TxnDate     string
	Description string
	Debit       decimal.Decimal
	Credit      decimal.Decimal
	Balance     *decimal.Decimal
	ReferenceNo *string
}

// Parse reads a CSV from r using the given template and returns parsed rows.
func Parse(r io.Reader, tmpl Template) ([]Row, error) {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	var rows []Row
	lineNo := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("csv read error at line %d: %w", lineNo+1, err)
		}
		lineNo++
		if lineNo <= tmpl.SkipRows {
			continue
		}

		row, err := parseRecord(record, tmpl, lineNo)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
		if row == nil {
			continue
		}
		rows = append(rows, *row)
	}
	return rows, nil
}

func parseRecord(record []string, tmpl Template, lineNo int) (*Row, error) {
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

	// Date
	rawDate := strings.TrimSpace(record[tmpl.DateColumn])
	if rawDate == "" {
		return nil, nil
	}
	t, err := time.Parse(tmpl.DateFormat, rawDate)
	if err != nil {
		return nil, fmt.Errorf("date parse %q with format %q: %w", rawDate, tmpl.DateFormat, err)
	}
	txnDate := t.Format("2006-01-02")

	// Description
	description := strings.TrimSpace(record[tmpl.DescriptionColumn])

	// Debit / Credit
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

	// Both zero → skip empty rows
	if debit.IsZero() && credit.IsZero() {
		return nil, nil
	}

	// Balance (optional)
	var balance *decimal.Decimal
	if tmpl.BalanceColumn != nil {
		raw := strings.TrimSpace(record[*tmpl.BalanceColumn])
		if raw != "" {
			v := parseAmount(raw)
			balance = &v
		}
	}

	// Reference (optional)
	var referenceNo *string
	if tmpl.ReferenceColumn != nil {
		raw := strings.TrimSpace(record[*tmpl.ReferenceColumn])
		if raw != "" {
			referenceNo = &raw
		}
	}

	return &Row{
		TxnDate:     txnDate,
		Description: description,
		Debit:       debit,
		Credit:      credit,
		Balance:     balance,
		ReferenceNo: referenceNo,
	}, nil
}

// parseAmount converts a string like "1,234.56" or "(1,234.56)" to decimal.
func parseAmount(s string) decimal.Decimal {
	s = strings.TrimSpace(s)
	negative := false
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		negative = true
		s = s[1 : len(s)-1]
	}
	// Remove currency symbols and whitespace, keep digits, dot, minus
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
