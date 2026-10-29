package cashmanager

import "webtyp.com/model"

const (
	ErrNotFound           model.NotFound = "shift not found"
	ErrValidationFailed   model.Invalid  = "validation failed"
	ErrShiftAlreadyOpen   model.Conflict = "a shift is already open for this cashier"
	ErrShiftAlreadyClosed model.Conflict = "shift already closed"
	ErrCannotRead         model.Invalid  = "cash_manager: cannot read the request"
)
