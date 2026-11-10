package cashmanager

import "webtyp.com/model"

const (
	ErrNotFound           model.NotFound = "shift not found"
	ErrValidationFailed   model.Invalid  = "validation failed"
	ErrShiftAlreadyOpen   model.Conflict = "a shift is already open for this cashier"
	ErrShiftAlreadyClosed model.Conflict = "shift already closed"
	ErrCannotRead         model.Invalid  = "cash_manager: cannot read the request"
	ErrNotYourShift     model.Invalid = "cash_manager: that shift belongs to another cashier, so it was not changed. Close only the shift you opened."
	ErrFieldNotEditable model.Invalid = "cash_manager: that field cannot be changed. Send only closing_cash_counted when closing a shift."
	ErrRangeRequired    model.Invalid = "cash_manager: from and to are required. Send both dates of the range."
)
