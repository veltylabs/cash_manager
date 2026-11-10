package cashmanager

import (
	"webtyp.com/model"
)

const (
	ShiftOpen   = "open"
	ShiftClosed = "closed"
	RowFact     = "fact"
	RowForecast = "forecast"
)

var ShiftModel = model.Definition{
	Name: "cash_shift",
	Allow: model.Create | model.Read | model.Update,
	Rules: model.Rules{Create: model.AddMyBusinessRule, Read: model.HaveNoBusinessRule, Update: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true, ServerOnly: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "cashier_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "status", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "opening_cash", Type: model.Int(), NotNull: true},
		{Name: "closing_cash_counted", Type: model.Int(), Label: "Counted cash"},
		{Name: "closing_cash_expected", Type: model.Int(), Label: "Expected cash", ServerOnly: true},
		{Name: "difference", Type: model.Int(), ServerOnly: true},
		{Name: "currency", Type: model.Text(), NotNull: true, Permitted: model.Permitted{Letters: true, Minimum: 3, Maximum: 3}},
		{Name: "opened_at", Type: model.UnixTime(), NotNull: true, ServerOnly: true},
		{Name: "closed_at", Type: model.UnixTime(), OmitEmpty: true, ServerOnly: true},
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true, ServerOnly: true},
	},
}

var StoredFactModel = model.Definition{
	Name: "stored_fact",
	Allow: model.Read,
	Rules: model.Rules{Read: model.HaveNoBusinessRule},
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'|', '_', '-', '.'}}},
		{Name: "tenant_id", Type: model.Text(), NotNull: true, ServerOnly: true},
		{Name: "kind", Type: model.Text(), NotNull: true, Item: model.ItemLabel},
		{Name: "side", Type: model.Int(), NotNull: true},
		{Name: "currency", Type: model.Text(), NotNull: true, Permitted: model.Permitted{Letters: true, Minimum: 3, Maximum: 3}, Item: model.ItemDescription},
		{Name: "amount", Type: model.Int(), NotNull: true, Item: model.ItemDescription},
		{Name: "scope", Type: model.Text(), NotNull: true, Item: model.ItemDescription},
		{Name: "source_type", Type: model.Text(), NotNull: true, Item: model.ItemDescription},
		{Name: "source_id", Type: model.Text(), NotNull: true, Item: model.ItemDescription},
		{Name: "party", Type: model.Text(), OmitEmpty: true},
		{Name: "at", Type: model.UnixTime(), NotNull: true},
		{Name: "due_at", Type: model.UnixTime(), OmitEmpty: true},
		{Name: "means", Type: model.Text(), OmitEmpty: true}, // CASH, CARD o TRANSFER
		{Name: "actor", Type: model.Text(), OmitEmpty: true}, // quién lo registró
	},
}

// TodayShiftModel is the caller's open shift (zero or one row).
var TodayShiftModel = model.Definition{
	Name:  "cash_today",
	Wire:  true,
	Allow: model.Read,
	Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "id", Type: model.Text()},
		{Name: "cashier_id", Type: model.Text()},
		{Name: "status", Type: model.Text()},
		{Name: "opening_cash", Type: model.Int()},
		{Name: "currency", Type: model.Text()},
		{Name: "opened_at", Type: model.UnixTime()},
		{Name: "facts_count", Type: model.Int()},
	},
}

// ReorderModel lists the items to buy again: one row per item (it replaces the old ReorderResponse).
var ReorderModel = model.Definition{
	Name:  "reorder",
	Wire:  true,
	Allow: model.Read,
	Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "item_id", Type: model.Text()},
		{Name: "item_name", Type: model.Text()},
		{Name: "total", Type: model.Float()},
		{Name: "min_quantity", Type: model.Float()},
	},
}

// ThirtyDayModel is the next-30-days view: one row per stored fact due in the range and one per forecast line.
// from, to and scope are the filter the caller sends; the other fields are the answer.
var ThirtyDayModel = model.Definition{
	Name:  "thirty_day",
	Wire:  true,
	Allow: model.Read,
	Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "from", Type: model.UnixTime(), OmitEmpty: true},
		{Name: "to", Type: model.UnixTime(), OmitEmpty: true},
		{Name: "scope", Type: model.Text(), OmitEmpty: true},
		{Name: "row_type", Type: model.Text(), OmitEmpty: true},
		{Name: "label", Type: model.Text(), OmitEmpty: true},
		{Name: "amount", Type: model.Int(), OmitEmpty: true},
		{Name: "currency", Type: model.Text(), OmitEmpty: true},
		{Name: "due_at", Type: model.UnixTime(), OmitEmpty: true},
	},
}

var Definition = model.ModuleDefinition{
	Name: "cash_manager",
	Entities: []model.Definition{ShiftModel, StoredFactModel, TodayShiftModel, ReorderModel, ThirtyDayModel},
}
