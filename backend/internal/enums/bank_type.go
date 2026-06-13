package enums

//go:generate go run ../../cmd/enumx-gen
import "akatengu/internal/pkg/enumx"

type BankType = enumx.Enum[bankTypeVal]

//enumx:enum
type bankTypeVal string

const (
	BankTypeTBD     bankTypeVal = "TBD"
	BankTypeSinoPac bankTypeVal = "SINOPAC"
)
