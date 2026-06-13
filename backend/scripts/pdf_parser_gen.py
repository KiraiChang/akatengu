#!/usr/bin/env python3
"""
pdf_parser_gen.py — 銀行 PDF 對帳單解析腳本 & Go parser 骨架產生器

使用方式:
  python3 scripts/pdf_parser_gen.py <pdf_path> [--password PW] [--bank-name BANK] [--generate]

選項:
  --password   PDF 開啟密碼
  --bank-name  銀行英文代號（用於產生 Go package 名稱），預設 unknown
  --generate   印出 Go parser 骨架程式碼

範例:
  # 解析並預覽交易
  python3 scripts/pdf_parser_gen.py ~/Downloads/statement.pdf --password abc123

  # 同時產生 Go 骨架
  python3 scripts/pdf_parser_gen.py ~/Downloads/statement.pdf --password abc123 --bank-name esun --generate
"""

import argparse
import re
import subprocess
import sys
import textwrap
from dataclasses import dataclass, field
from decimal import Decimal, InvalidOperation
from typing import Optional


# ─── PDF text extraction ───────────────────────────────────────────────────────

def extract_text(pdf_path: str, password: Optional[str] = None) -> str:
    args = ["pdftotext", "-layout", "-enc", "UTF-8"]
    if password:
        args += ["-upw", password]
    args += [pdf_path, "-"]
    result = subprocess.run(args, capture_output=True, text=True)
    if result.returncode != 0:
        stderr = result.stderr.strip()
        if "Incorrect password" in stderr:
            print("✗ 密碼錯誤", file=sys.stderr)
        else:
            print(f"✗ pdftotext 失敗: {stderr}", file=sys.stderr)
        sys.exit(1)
    return result.stdout


# ─── Column detection ──────────────────────────────────────────────────────────

DEBIT_KEYWORDS  = ["支出", "借方", "Debit",  "DR"]
CREDIT_KEYWORDS = ["存入", "貸方", "Credit", "CR"]
BALANCE_KEYWORDS = ["餘額", "結餘", "Balance"]
DATE_RE = re.compile(r"^\d{4}[/\-]\d{2}[/\-]\d{2}")
NUM_RE  = re.compile(r"\d+(?:,\d+)*")


@dataclass
class Columns:
    debit_start:   int = -1
    credit_start:  int = -1
    balance_start: int = -1
    note_start:    int = -1
    header_line:   str = ""


def find_header(lines: list[str]) -> Optional[tuple[int, str]]:
    """尋找包含借方/貸方/餘額關鍵字的表頭行，回傳 (行號, 行內容)"""
    for i, line in enumerate(lines):
        has_debit   = any(k in line for k in DEBIT_KEYWORDS)
        has_credit  = any(k in line for k in CREDIT_KEYWORDS)
        has_balance = any(k in line for k in BALANCE_KEYWORDS)
        if has_debit and has_credit and has_balance:
            return i, line
    return None


def detect_columns(header: str) -> Columns:
    """從表頭行的關鍵字位置推算各欄 rune 起點"""
    cols = Columns(header_line=header)
    runes = list(header)

    for kw in DEBIT_KEYWORDS:
        idx = header.find(kw)
        if idx >= 0:
            cols.debit_start = len(list(header[:idx]))  # rune 位置
            break
    for kw in CREDIT_KEYWORDS:
        idx = header.find(kw)
        if idx >= 0:
            cols.credit_start = len(list(header[:idx]))
            break
    for kw in BALANCE_KEYWORDS:
        idx = header.find(kw)
        if idx >= 0:
            cols.balance_start = len(list(header[:idx]))
            break

    # 備註欄：餘額右邊的第一個非空白字詞位置
    if cols.balance_start >= 0:
        after = runes[cols.balance_start + 2:]  # 跳過 "餘額"
        for j, c in enumerate(after):
            if c.strip():
                cols.note_start = cols.balance_start + 2 + j
                break

    return cols


# ─── Account section detection ────────────────────────────────────────────────

ACCOUNT_RE = re.compile(r"(?:帳號|Account)[：:]\s*[\w*\-]+")


def detect_account_sections(lines: list[str]) -> list[tuple[int, str]]:
    """回傳 (行號, 帳號字串) 清單"""
    sections = []
    for i, line in enumerate(lines):
        m = ACCOUNT_RE.search(line)
        if m:
            sections.append((i, m.group().strip()))
    return sections


# ─── Transaction parsing ───────────────────────────────────────────────────────

@dataclass
class Transaction:
    account: str
    date: str
    description: str
    debit:   Optional[Decimal]
    credit:  Optional[Decimal]
    balance: Optional[Decimal]
    note:    str = ""


def extract_amount(runes: list[str], start: int, end: int) -> Optional[Decimal]:
    if start < 0 or start >= len(runes):
        return None
    end = min(end, len(runes))
    segment = "".join(runes[start:end])
    m = NUM_RE.search(segment)
    if not m:
        return None
    cleaned = m.group().replace(",", "")
    try:
        return Decimal(cleaned)
    except InvalidOperation:
        return None


def parse_transactions(lines: list[str], cols: Columns, sections: list[tuple[int, str]]) -> list[Transaction]:
    txns = []
    section_idx = -1
    section_lines = {ln for ln, _ in sections}
    section_account = {ln: acc for ln, acc in sections}
    current_account = "unknown"

    for i, line in enumerate(lines):
        if i in section_lines:
            section_idx += 1
            current_account = section_account[i]
            continue

        if not DATE_RE.match(line):
            continue

        runes = list(line)
        if len(runes) < cols.balance_start:
            continue

        date = "".join(runes[0:10]).replace("/", "-").strip()

        desc_end = cols.debit_start if cols.debit_start > 0 else cols.credit_start
        desc = "".join(runes[10:desc_end]).strip() if desc_end > 10 else ""

        debit   = extract_amount(runes, cols.debit_start,   cols.credit_start)
        credit  = extract_amount(runes, cols.credit_start,  cols.balance_start)

        balance: Optional[Decimal] = None
        note = ""
        if len(runes) > cols.balance_start:
            after_bal = "".join(runes[cols.balance_start:])
            m = NUM_RE.search(after_bal)
            if m:
                try:
                    balance = Decimal(m.group().replace(",", ""))
                except InvalidOperation:
                    pass
                note = after_bal[m.end():].strip()

        txns.append(Transaction(
            account=current_account,
            date=date,
            description=desc,
            debit=debit,
            credit=credit,
            balance=balance,
            note=note,
        ))

    return txns


# ─── Output ────────────────────────────────────────────────────────────────────

def print_analysis(cols: Columns, sections: list, txns: list[Transaction]):
    print("=" * 70)
    print("▌ 欄位偵測結果")
    print("=" * 70)
    print(f"  表頭行 : {cols.header_line.strip()}")
    print(f"  支出 (Debit)   起始 rune : {cols.debit_start}")
    print(f"  存入 (Credit)  起始 rune : {cols.credit_start}")
    print(f"  餘額 (Balance) 起始 rune : {cols.balance_start}")
    print(f"  備註 (Note)    起始 rune : {cols.note_start}")
    print()

    print("▌ 帳戶段落")
    for _, acc in sections:
        print(f"  • {acc}")
    print()

    print(f"▌ 交易明細（共 {len(txns)} 筆）")
    print("-" * 70)
    cur_acc = None
    for t in txns:
        if t.account != cur_acc:
            cur_acc = t.account
            print(f"\n  [{cur_acc}]")
        dr = str(t.debit) if t.debit is not None else ""
        cr = str(t.credit) if t.credit is not None else ""
        bal = str(t.balance) if t.balance is not None else "-"
        print(f"    {t.date}  {t.description:<10}  支出={dr:<10} 存入={cr:<10} 餘額={bal:<10} {t.note}")
    print()


# ─── Go parser code generation ────────────────────────────────────────────────

GO_TEMPLATE = '''\
package {pkg}

import (
\t"bufio"
\t"fmt"
\t"io"
\t"os"
\t"os/exec"
\t"regexp"
\t"strings"

\t"akatengu/internal/pkg/pdfparser"

\t"github.com/shopspring/decimal"
)

// Column rune boundaries detected from pdftotext -layout output.
// Header: {header}
const (
\tdebitStart   = {debit}
\tcreditStart  = {credit}
\tbalanceStart = {balance}
)

var (
\tdateRe    = regexp.MustCompile(`^\d{{4}}[/\\-]\d{{2}}[/\\-]\d{{2}}`)
\taccountRe = regexp.MustCompile(`TODO: 填入帳號行的正則`) // TODO: 依實際 PDF 調整
\tnumRe     = regexp.MustCompile(`\d+(?:,\d+)*`)
)

type Parser struct{{}}

func (p *Parser) Parse(r io.Reader, ledgers []pdfparser.LedgerHint, opts pdfparser.ParseOptions) ([]pdfparser.ParsedRow, error) {{
\tif len(ledgers) == 0 {{
\t\treturn nil, fmt.Errorf("{pkg} parser: at least one ledger hint is required")
\t}}

\ttmpFile, err := os.CreateTemp("", "{pkg}-*.pdf")
\tif err != nil {{
\t\treturn nil, fmt.Errorf("{pkg} parser: create temp file: %w", err)
\t}}
\tdefer os.Remove(tmpFile.Name())

\tif _, err := io.Copy(tmpFile, r); err != nil {{
\t\ttmpFile.Close()
\t\treturn nil, fmt.Errorf("{pkg} parser: write temp file: %w", err)
\t}}
\ttmpFile.Close()

\targs := []string{{"-layout", "-enc", "UTF-8"}}
\tif opts.Password != nil && *opts.Password != "" {{
\t\targs = append(args, "-upw", *opts.Password)
\t}}
\targs = append(args, tmpFile.Name(), "-")

\tout, err := exec.Command("pdftotext", args...).Output()
\tif err != nil {{
\t\treturn nil, fmt.Errorf("{pkg} parser: pdftotext failed: %w", err)
\t}}

\treturn parseText(string(out), ledgers)
}}

func parseText(text string, ledgers []pdfparser.LedgerHint) ([]pdfparser.ParsedRow, error) {{
\tvar rows []pdfparser.ParsedRow
\tsectionIdx := -1

\tscanner := bufio.NewScanner(strings.NewReader(text))
\tfor scanner.Scan() {{
\t\tline := scanner.Text()

\t\tif accountRe.MatchString(line) {{
\t\t\tsectionIdx++
\t\t\tcontinue
\t\t}}
\t\tif !dateRe.MatchString(line) {{
\t\t\tcontinue
\t\t}}

\t\tledgerUUID := ledgers[len(ledgers)-1].LedgerUUID
\t\tif sectionIdx >= 0 && sectionIdx < len(ledgers) {{
\t\t\tledgerUUID = ledgers[sectionIdx].LedgerUUID
\t\t}}

\t\trow, err := parseLine(line, ledgerUUID)
\t\tif err != nil {{
\t\t\tcontinue
\t\t}}
\t\trows = append(rows, *row)
\t}}
\treturn rows, nil
}}

func parseLine(line, ledgerUUID string) (*pdfparser.ParsedRow, error) {{
\trunes := []rune(line)
\tif len(runes) < balanceStart {{
\t\treturn nil, fmt.Errorf("line too short")
\t}}

\tdate := strings.TrimSpace(string(runes[0:10]))
\tdate = strings.NewReplacer("/", "-").Replace(date)

\tdescEnd := debitStart
\tif descEnd > len(runes) {{
\t\tdescEnd = len(runes)
\t}}
\tdesc := strings.TrimSpace(string(runes[10:descEnd]))

\tdebit  := extractAmount(runes, debitStart, creditStart)
\tcredit := extractAmount(runes, creditStart, balanceStart)

\tvar balance *decimal.Decimal
\tvar refNo *string

\tif len(runes) > balanceStart {{
\t\tafterBal := string(runes[balanceStart:])
\t\tm := numRe.FindStringIndex(afterBal)
\t\tif m != nil {{
\t\t\tbalStr := strings.ReplaceAll(afterBal[m[0]:m[1]], ",", "")
\t\t\tb, err := decimal.NewFromString(balStr)
\t\t\tif err == nil {{
\t\t\t\tbalance = &b
\t\t\t}}
\t\t\trest := strings.TrimSpace(afterBal[m[1]:])
\t\t\tif rest != "" {{
\t\t\t\trefNo = &rest
\t\t\t}}
\t\t}}
\t}}

\tvar debitAmt, creditAmt decimal.Decimal
\tif debit != nil {{
\t\tdebitAmt = *debit
\t}}
\tif credit != nil {{
\t\tcreditAmt = *credit
\t}}

\treturn &pdfparser.ParsedRow{{
\t\tLedgerUUID:  ledgerUUID,
\t\tTxnDate:     date,
\t\tDescription: desc,
\t\tDebit:       debitAmt,
\t\tCredit:      creditAmt,
\t\tBalance:     balance,
\t\tReferenceNo: refNo,
\t}}, nil
}}

func extractAmount(runes []rune, start, end int) *decimal.Decimal {{
\tif start < 0 || start >= len(runes) {{
\t\treturn nil
\t}}
\tif end > len(runes) {{
\t\tend = len(runes)
\t}}
\tsegment := string(runes[start:end])
\tm := numRe.FindString(segment)
\tif m == "" {{
\t\treturn nil
\t}}
\tcleaned := strings.ReplaceAll(m, ",", "")
\td, err := decimal.NewFromString(cleaned)
\tif err != nil {{
\t\treturn nil
\t}}
\treturn &d
}}
'''


def generate_go(pkg: str, cols: Columns):
    print("=" * 70)
    print(f"▌ 產生 Go parser 骨架  →  internal/pkg/pdfparser/{pkg}/parser.go")
    print("=" * 70)
    code = GO_TEMPLATE.format(
        pkg=pkg,
        header=repr(cols.header_line.strip()),
        debit=cols.debit_start,
        credit=cols.credit_start,
        balance=cols.balance_start,
    )
    print(code)

    # 同時寫入檔案
    import os
    out_dir = os.path.join(
        os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
        "internal", "pkg", "pdfparser", pkg,
    )
    os.makedirs(out_dir, exist_ok=True)
    out_path = os.path.join(out_dir, "parser.go")
    if os.path.exists(out_path):
        print(f"⚠  {out_path} 已存在，跳過寫入（請手動確認）")
    else:
        with open(out_path, "w", encoding="utf-8") as f:
            f.write(code)
        print(f"✓ 已寫入 {out_path}")
    print()
    print("後續步驟：")
    print(f"  1. 在 internal/enums/bank_type.go 新增  BankType{pkg.capitalize()} bankTypeVal = \"{pkg.upper()}\"")
    print(f"  2. 執行 go generate ./internal/enums/...")
    print(f"  3. 在 internal/pkg/pdfparserfactory/factory.go 加入  case \"{pkg.upper()}\": return &{pkg}.Parser{{}}, nil")
    print(f"  4. 調整 parser.go 中 accountRe 正則與欄位邊界後執行 go build ./...")


# ─── Entry point ───────────────────────────────────────────────────────────────

def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("pdf_path", help="PDF 檔案路徑")
    ap.add_argument("--password", "-p", default=None, help="PDF 密碼")
    ap.add_argument("--bank-name", "-b", default="unknown", help="銀行英文代號（小寫）")
    ap.add_argument("--generate", "-g", action="store_true", help="產生 Go parser 骨架")
    args = ap.parse_args()

    print(f"▶ 讀取 PDF：{args.pdf_path}")
    text = extract_text(args.pdf_path, args.password)
    lines = text.split("\n")
    print(f"  共 {len(lines)} 行文字\n")

    result = find_header(lines)
    if result is None:
        print("✗ 找不到含借貸欄的表頭行，請確認 PDF 格式或手動設定欄位邊界", file=sys.stderr)
        sys.exit(1)

    header_line_no, header = result
    cols = detect_columns(header)

    if cols.debit_start < 0 or cols.credit_start < 0 or cols.balance_start < 0:
        print("✗ 無法自動偵測所有欄位邊界，請手動設定", file=sys.stderr)
        print(f"  偵測結果：debit={cols.debit_start} credit={cols.credit_start} balance={cols.balance_start}")
        sys.exit(1)

    sections = detect_account_sections(lines)
    txns = parse_transactions(lines, cols, sections)

    print_analysis(cols, sections, txns)

    if args.generate:
        generate_go(args.bank_name, cols)


if __name__ == "__main__":
    main()
