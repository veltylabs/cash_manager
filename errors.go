package cashmanager

type domainError string

func (e domainError) Error() string { return string(e) }

const (
	ErrNotFound           domainError = "shift not found"
	ErrValidationFailed   domainError = "validation failed"
	ErrShiftAlreadyOpen   domainError = "a shift is already open for this cashier"
	ErrShiftAlreadyClosed domainError = "shift already closed"
)
