package pdfparserfactory

import (
	"fmt"

	"akatengu/internal/pkg/pdfparser"
	"akatengu/internal/pkg/pdfparser/sinopac"
)

func New(bankType string) (pdfparser.BankPDFParser, error) {
	switch bankType {
	case "SINOPAC":
		return &sinopac.Parser{}, nil
	default:
		return nil, fmt.Errorf("unsupported bank type: %s", bankType)
	}
}
