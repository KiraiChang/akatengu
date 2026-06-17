package errors

type ErrorCategory string

const (
	CategoryContract       ErrorCategory = "contract"
	CategoryRuntime        ErrorCategory = "runtime"
	CategoryBusiness       ErrorCategory = "business"
	CategoryInfrastructure ErrorCategory = "infra"
)
