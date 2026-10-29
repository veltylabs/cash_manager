package cashmanager

import (
	"webtyp.com/model"
	"webtyp.com/time"
)

var ResourceFeatureReorder = model.FeatureResource("reorder")
var ResourceFeatureThirtyDay = model.FeatureResource("thirty_day")

const (
	ShiftOpen   = "open"
	ShiftClosed = "closed"
)

var ShiftModel = model.Definition{
	Name: "cash_shift",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "cashier_id", Type: model.Text(), NotNull: true},
		{Name: "status", Type: model.Text(), NotNull: true},
		{Name: "opening_cash", Type: model.Int(), NotNull: true},
		{Name: "closing_cash_counted", Type: model.Int(), Label: "Counted cash"},
		{Name: "closing_cash_expected", Type: model.Int(), Label: "Expected cash"},
		{Name: "difference", Type: model.Int()},
		{Name: "currency", Type: model.Text(), NotNull: true, Permitted: model.Permitted{Letters: true, Minimum: 3, Maximum: 3}},
		{Name: "opened_at", Type: model.UnixTime(), NotNull: true},
		{Name: "closed_at", Type: model.UnixTime(), OmitEmpty: true},
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true},
	},
}

var StoredFactModel = model.Definition{
	Name: "stored_fact",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'|', '_', '-', '.'}}},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
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

type OpenShiftArgs struct {
	TenantId    string `json:"tenant_id"`
	CashierId   string `json:"cashier_id"`
	OpeningCash int64  `json:"opening_cash"`
	Currency    string `json:"currency"`
}
func (a *OpenShiftArgs) Schema() []model.Field { return nil }
func (a *OpenShiftArgs) Pointers() []any { return []any{&a.TenantId, &a.CashierId, &a.OpeningCash, &a.Currency} }
func (a *OpenShiftArgs) EncodeFields(w model.FieldWriter) { w.String("tenant_id", a.TenantId); w.String("cashier_id", a.CashierId); w.Int("opening_cash", a.OpeningCash); w.String("currency", a.Currency) }
func (a *OpenShiftArgs) DecodeFields(r model.FieldReader) { a.TenantId, _ = r.String("tenant_id"); a.CashierId, _ = r.String("cashier_id"); a.OpeningCash, _ = r.Int("opening_cash"); a.Currency, _ = r.String("currency") }
func (a *OpenShiftArgs) IsNil() bool { return a == nil }

type CloseShiftArgs struct {
	TenantId           string `json:"tenant_id"`
	Id                 string `json:"id"`
	ClosingCashCounted int64  `json:"closing_cash_counted"`
}
func (a *CloseShiftArgs) Schema() []model.Field { return nil }
func (a *CloseShiftArgs) Pointers() []any { return []any{&a.TenantId, &a.Id, &a.ClosingCashCounted} }
func (a *CloseShiftArgs) EncodeFields(w model.FieldWriter) { w.String("tenant_id", a.TenantId); w.String("id", a.Id); w.Int("closing_cash_counted", a.ClosingCashCounted) }
func (a *CloseShiftArgs) DecodeFields(r model.FieldReader) { a.TenantId, _ = r.String("tenant_id"); a.Id, _ = r.String("id"); a.ClosingCashCounted, _ = r.Int("closing_cash_counted") }
func (a *CloseShiftArgs) IsNil() bool { return a == nil }

type GetTodayArgs struct {
	TenantId  string `json:"tenant_id"`
	CashierId string `json:"cashier_id"`
}
func (a *GetTodayArgs) Schema() []model.Field { return nil }
func (a *GetTodayArgs) Pointers() []any { return []any{&a.TenantId, &a.CashierId} }
func (a *GetTodayArgs) EncodeFields(w model.FieldWriter) { w.String("tenant_id", a.TenantId); w.String("cashier_id", a.CashierId) }
func (a *GetTodayArgs) DecodeFields(r model.FieldReader) { a.TenantId, _ = r.String("tenant_id"); a.CashierId, _ = r.String("cashier_id") }
func (a *GetTodayArgs) IsNil() bool { return a == nil }

type ListStoredFactsArgs struct {
	TenantId string    `json:"tenant_id"`
	Scope    string    `json:"scope"`
	From     time.Unix `json:"from"`
	To       time.Unix `json:"to"`
}
func (a *ListStoredFactsArgs) Schema() []model.Field { return nil }
func (a *ListStoredFactsArgs) Pointers() []any { return []any{&a.TenantId, &a.Scope, &a.From, &a.To} }
func (a *ListStoredFactsArgs) EncodeFields(w model.FieldWriter) { w.String("tenant_id", a.TenantId); w.String("scope", a.Scope); w.Int("from", int64(a.From)); w.Int("to", int64(a.To)) }
func (a *ListStoredFactsArgs) DecodeFields(r model.FieldReader) { a.TenantId, _ = r.String("tenant_id"); a.Scope, _ = r.String("scope"); f, _ := r.Int("from"); a.From = time.Unix(f); t, _ := r.Int("to"); a.To = time.Unix(t) }
func (a *ListStoredFactsArgs) IsNil() bool { return a == nil }
