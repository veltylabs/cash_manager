package cashmanager

import (
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/storage"
	"webtyp.com/time"
)

func (p *TodayShift) Read(a crud.CurrentUser, filter *TodayShift, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	list := TodayShiftList{}
	row, found, err := firstRow(m.store, func() model.Model { return &Shift{} },
		storage.Eq(Shift_.CashierId, a.UserID), storage.Eq(Shift_.Status, ShiftOpen))
	if err != nil {
		return nil, err
	}
	if !found {
		return &list, nil
	}
	shift := row.(*Shift)
	facts, err := factsBetween(m.store, shift.OpenedAt, time.NowUnix())
	if err != nil {
		return nil, err
	}
	list = append(list, &TodayShift{
		Id: shift.Id, CashierId: shift.CashierId, Status: shift.Status, OpeningCash: shift.OpeningCash,
		Currency: shift.Currency, OpenedAt: shift.OpenedAt, FactsCount: int64(len(facts)),
	})
	return &list, nil
}
