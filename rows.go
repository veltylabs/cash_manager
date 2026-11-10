package cashmanager

import (
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/storage"
	"webtyp.com/time"
)

// firstRow answers the first row that matches conds, and whether there is one.
//
// Example: row, found, err := firstRow(tx, func() model.Model { return &Shift{} }, storage.Eq(Shift_.CashierId, "u1"))
func firstRow(tx crud.Tx, newRow func() model.Model, conds ...storage.Condition) (model.Model, bool, error) {
	var first model.Model
	err := tx.ReadAll(newRow, crud.Query{Conditions: conds, Limit: 1}, func(r model.Model) {
		if first == nil {
			first = r
		}
	})
	return first, first != nil, err
}

// factsWhere answers every stored fact that matches conds, paging by id because one read answers at most crud.MaxPage rows.
func factsWhere(tx crud.Tx, conds ...storage.Condition) ([]StoredFact, error) {
	var facts []StoredFact
	after := ""
	for {
		var page []StoredFact
		err := tx.ReadAll(func() model.Model { return &StoredFact{} },
			crud.Query{Conditions: conds, After: after, Limit: crud.MaxPage},
			func(r model.Model) { page = append(page, *r.(*StoredFact)) })
		if err != nil {
			return nil, err
		}
		facts = append(facts, page...)
		if len(page) < crud.MaxPage {
			return facts, nil
		}
		after = page[len(page)-1].Id
	}
}

// factsBetween answers the facts whose `at` is in [from, to].
func factsBetween(tx crud.Tx, from, to time.Unix) ([]StoredFact, error) {
	return factsWhere(tx, storage.Gte(StoredFact_.At, int64(from)), storage.Lte(StoredFact_.At, int64(to)))
}

// onlyFields refuses a field that is not in allowed.
func onlyFields(fields []string, allowed ...string) error {
	for _, field := range fields {
		permitted := false
		for _, name := range allowed {
			if field == name {
				permitted = true
			}
		}
		if !permitted {
			return &model.FieldError{Field: field, Err: ErrFieldNotEditable}
		}
	}
	return nil
}
