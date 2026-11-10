package cashmanager

import (
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/money"
	"webtyp.com/time"
)

func (next *Shift) Update(a crud.CurrentUser, tx crud.Tx, old *Shift, fields []string, m *Module) (crud.RuleAnswer, error) {
	if err := onlyFields(fields, Shift_.ClosingCashCounted); err != nil {
		return 0, err
	}
	if old.CashierId != a.UserID {
		return 0, &model.FieldError{Field: Shift_.CashierId, Err: ErrNotYourShift}
	}
	if old.Status == ShiftClosed {
		return 0, ErrShiftAlreadyClosed
	}
	now := time.NowUnix()
	facts, err := factsBetween(tx, old.OpenedAt, now)
	if err != nil {
		return 0, err
	}
	expected := old.OpeningCash
	for _, f := range facts {
		if f.Currency != old.Currency || f.Means != string(money.MeansCash) || f.Actor != old.CashierId {
			continue // another currency, card or transfer, or another cashier: not this shift's cash
		}
		switch f.Kind {
		case string(money.Charged):
			expected += f.Amount
		case string(money.Refunded):
			expected -= f.Amount
		}
	}

	derived := Shift{
		Id:                  old.Id,
		Status:              ShiftClosed,
		ClosingCashExpected: expected,
		Difference:          next.ClosingCashCounted - expected,
		ClosedAt:            now,
	}
	err = tx.UpdateFields(&derived, old.Id, []string{Shift_.Status, Shift_.ClosingCashExpected, Shift_.Difference, Shift_.ClosedAt})
	if err != nil {
		return 0, err
	}
	return crud.ApplyAndPublish, nil
}
