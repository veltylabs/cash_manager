package cashmanager

import (
	"webtyp.com/crud"
	"webtyp.com/model"
)

func (p *Reorder) Read(a crud.CurrentUser, filter *Reorder, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	list := ReorderList{}
	if m.deps.LowStock == nil {
		return &list, nil
	}
	items, err := m.deps.LowStock.LowStock(m.deps.TenantID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		list = append(list, &Reorder{ItemId: items[i].ItemId, ItemName: items[i].ItemName, Total: items[i].Total, MinQuantity: items[i].MinQuantity})
	}
	return &list, nil
}
