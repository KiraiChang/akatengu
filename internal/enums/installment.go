package enums

import "akatengu/internal/pkg/enumx"

type InterestType = enumx.Enum[interestTypeVal]

//enumx:enum
type interestTypeVal string

const (
	InterestTypeFree      interestTypeVal = "FREE"
	InterestTypeFixedRate interestTypeVal = "FIXED_RATE"
)

type InstallmentStatus = enumx.Enum[installmentStatusVal]

//enumx:enum
type installmentStatusVal string

const (
	InstallmentStatusActive    installmentStatusVal = "ACTIVE"
	InstallmentStatusCompleted installmentStatusVal = "COMPLETED"
	InstallmentStatusCanceled  installmentStatusVal = "CANCELED"
)

type InstallmentPaymentStatus = enumx.Enum[installmentPaymentStatusVal]

//enumx:enum
type installmentPaymentStatusVal string

const (
	InstallmentPaymentStatusPending installmentPaymentStatusVal = "PENDING"
	InstallmentPaymentStatusPaid    installmentPaymentStatusVal = "PAID"
)
