package cashmanager

import (
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/storage"
	"webtyp.com/time"
)

func (p *Shift) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	_, found, err := firstRow(tx, func() model.Model { return &Shift{} },
		storage.Eq(Shift_.CashierId, a.UserID), storage.Eq(Shift_.Status, ShiftOpen))
	if err != nil {
		return 0, err
	}
	if found {
		return 0, ErrShiftAlreadyOpen
	}
	p.CashierId = a.UserID
	p.Status = ShiftOpen
	p.OpenedAt = time.NowUnix()
	return crud.ApplyAndPublish, nil
}
