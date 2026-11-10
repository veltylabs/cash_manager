package cashmanager

import (
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/storage"
)

func (p *ThirtyDay) Read(a crud.CurrentUser, filter *ThirtyDay, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	if filter == nil || filter.From == 0 || filter.To == 0 {
		return nil, &model.FieldError{Field: "from", Err: ErrRangeRequired}
	}
	conditions := []storage.Condition{
		storage.Gte(StoredFact_.DueAt, int64(filter.From)),
		storage.Lte(StoredFact_.DueAt, int64(filter.To)),
	}
	if filter.Scope != "" {
		conditions = append(conditions, storage.Eq(StoredFact_.Scope, filter.Scope))
	}
	facts, err := factsWhere(m.store, conditions...)
	if err != nil {
		return nil, err
	}
	list := ThirtyDayList{}
	for i := range facts {
		list = append(list, &ThirtyDay{RowType: RowFact, Label: facts[i].Kind, Amount: facts[i].Amount, Currency: facts[i].Currency, DueAt: facts[i].DueAt})
	}
	if m.deps.Forecast != nil {
		lines, err := m.deps.Forecast.PendingAccruals(m.deps.TenantID, filter.From, filter.To)
		if err != nil {
			return nil, err
		}
		for i := range lines {
			list = append(list, &ThirtyDay{RowType: RowForecast, Label: lines[i].ThirdPartyName, Amount: lines[i].Amount, Currency: lines[i].Currency, DueAt: lines[i].DueAt})
		}
	}
	return &list, nil
}
