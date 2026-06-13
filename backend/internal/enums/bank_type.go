package enums

//go:generate go run ../../cmd/enumx-gen
import "akatengu/internal/pkg/enumx"

type BankType = enumx.Enum[bankTypeVal]

//enumx:enum
type bankTypeVal string

const (
	// 等 PDF 確認後填入實際銀行名稱，目前以 TBD 佔位
	BankTypeTBD bankTypeVal = "TBD"
)
