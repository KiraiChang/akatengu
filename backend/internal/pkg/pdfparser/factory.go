package pdfparser

import "fmt"

// New 根據銀行類型回傳對應的 BankPDFParser。
// bankType 對應 enums.BankType 的字串值。
func New(bankType string) (BankPDFParser, error) {
	switch bankType {
	// 等 PDF 確認銀行後在此加入 case，例如：
	// case "E_SUN":
	//     return esun.New(), nil
	default:
		return nil, fmt.Errorf("unsupported bank type: %s", bankType)
	}
}
