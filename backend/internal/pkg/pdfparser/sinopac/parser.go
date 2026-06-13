package sinopac

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"akatengu/internal/pkg/pdfparser"

	"github.com/shopspring/decimal"
)

// Column rune boundaries determined from pdftotext -layout output of 永豐銀行 statements.
// Header: "  交易日  摘要  支出  存入  餘額  備註/資金用途"
//   支出 (Debit)  starts at rune[32]
//   存入 (Credit) starts at rune[52]
//   餘額 (Balance) starts at rune[69]
const (
	debitStart   = 32
	creditStart  = 52
	balanceStart = 69
)

var (
	dateRe   = regexp.MustCompile(`^\d{4}/\d{2}/\d{2}`)
	accountRe = regexp.MustCompile(`帳號:[\w*-]+\(`)
	numRe    = regexp.MustCompile(`\d+(?:,\d+)*`)
)

type Parser struct{}

func (p *Parser) Parse(r io.Reader, ledgers []pdfparser.LedgerHint, opts pdfparser.ParseOptions) ([]pdfparser.ParsedRow, error) {
	if len(ledgers) == 0 {
		return nil, fmt.Errorf("sinopac parser: at least one ledger hint is required")
	}

	tmpFile, err := os.CreateTemp("", "sinopac-*.pdf")
	if err != nil {
		return nil, fmt.Errorf("sinopac parser: create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, r); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("sinopac parser: write temp file: %w", err)
	}
	tmpFile.Close()

	args := []string{"-layout", "-enc", "UTF-8"}
	if opts.Password != nil && *opts.Password != "" {
		args = append(args, "-upw", *opts.Password)
	}
	args = append(args, tmpFile.Name(), "-")

	out, err := exec.Command("pdftotext", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("sinopac parser: pdftotext failed: %w", err)
	}

	return parseText(string(out), ledgers)
}

func parseText(text string, ledgers []pdfparser.LedgerHint) ([]pdfparser.ParsedRow, error) {
	var rows []pdfparser.ParsedRow
	sectionIdx := -1

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()

		if accountRe.MatchString(line) {
			sectionIdx++
			continue
		}

		if !dateRe.MatchString(line) {
			continue
		}

		ledgerUUID := ledgers[len(ledgers)-1].LedgerUUID
		if sectionIdx >= 0 && sectionIdx < len(ledgers) {
			ledgerUUID = ledgers[sectionIdx].LedgerUUID
		}

		row, err := parseLine(line, ledgerUUID)
		if err != nil {
			continue
		}
		rows = append(rows, *row)
	}

	return rows, nil
}

func parseLine(line, ledgerUUID string) (*pdfparser.ParsedRow, error) {
	runes := []rune(line)
	if len(runes) < balanceStart {
		return nil, fmt.Errorf("line too short")
	}

	date := strings.TrimSpace(string(runes[0:10]))
	date = strings.ReplaceAll(date, "/", "-")

	end := debitStart
	if end > len(runes) {
		end = len(runes)
	}
	desc := strings.TrimSpace(string(runes[10:end]))

	debit := extractAmount(runes, debitStart, creditStart)
	credit := extractAmount(runes, creditStart, balanceStart)

	var balance *decimal.Decimal
	var refNo *string

	if len(runes) > balanceStart {
		afterBal := string(runes[balanceStart:])
		m := numRe.FindStringIndex(afterBal)
		if m != nil {
			balStr := strings.ReplaceAll(afterBal[m[0]:m[1]], ",", "")
			b, err := decimal.NewFromString(balStr)
			if err == nil {
				balance = &b
			}
			rest := strings.TrimSpace(afterBal[m[1]:])
			if rest != "" {
				refNo = &rest
			}
		}
	}

	var debitAmt, creditAmt decimal.Decimal
	if debit != nil {
		debitAmt = *debit
	}
	if credit != nil {
		creditAmt = *credit
	}

	return &pdfparser.ParsedRow{
		LedgerUUID:  ledgerUUID,
		TxnDate:     date,
		Description: desc,
		Debit:       debitAmt,
		Credit:      creditAmt,
		Balance:     balance,
		ReferenceNo: refNo,
	}, nil
}

func extractAmount(runes []rune, start, end int) *decimal.Decimal {
	if start >= len(runes) {
		return nil
	}
	if end > len(runes) {
		end = len(runes)
	}
	segment := string(runes[start:end])
	m := numRe.FindString(segment)
	if m == "" {
		return nil
	}
	cleaned := strings.ReplaceAll(m, ",", "")
	d, err := decimal.NewFromString(cleaned)
	if err != nil {
		return nil
	}
	return &d
}
