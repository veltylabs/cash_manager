package cashmanager

import (
	"webtyp.com/events"
	"webtyp.com/model"
	"webtyp.com/time"
)

type LowStockReader interface {
	LowStock(tenantId string) ([]LowStockItem, error)
}

type LowStockItem struct {
	ItemId      string  `json:"item_id"`
	ItemName    string  `json:"item_name"`
	Total       float64 `json:"total"`
	MinQuantity float64 `json:"min_quantity"`
}

func (i *LowStockItem) Schema() []model.Field { return nil }
func (i *LowStockItem) Pointers() []any { return []any{&i.ItemId, &i.ItemName, &i.Total, &i.MinQuantity} }
func (i *LowStockItem) EncodeFields(w model.FieldWriter) {
	w.String("item_id", i.ItemId)
	w.String("item_name", i.ItemName)
	w.Float("total", i.Total)
	w.Float("min_quantity", i.MinQuantity)
}
func (i *LowStockItem) DecodeFields(r model.FieldReader) {
	i.ItemId, _ = r.String("item_id")
	i.ItemName, _ = r.String("item_name")
	i.Total, _ = r.Float("total")
	i.MinQuantity, _ = r.Float("min_quantity")
}
func (i *LowStockItem) IsNil() bool { return i == nil }

type AgreementForecastReader interface {
	PendingAccruals(tenantId string, from, to time.Unix) ([]ForecastLine, error)
}

type ForecastLine struct {
	ThirdPartyName string    `json:"third_party_name"`
	Amount         int64     `json:"amount"`
	Currency       string    `json:"currency"`
	DueAt          time.Unix `json:"due_at"`
}

func (f *ForecastLine) Schema() []model.Field { return nil }
func (f *ForecastLine) Pointers() []any { return []any{&f.ThirdPartyName, &f.Amount, &f.Currency, (*int64)(&f.DueAt)} }
func (f *ForecastLine) EncodeFields(w model.FieldWriter) {
	w.String("third_party_name", f.ThirdPartyName)
	w.Int("amount", f.Amount)
	w.String("currency", f.Currency)
	w.Int("due_at", int64(f.DueAt))
}
func (f *ForecastLine) DecodeFields(r model.FieldReader) {
	f.ThirdPartyName, _ = r.String("third_party_name")
	f.Amount, _ = r.Int("amount")
	f.Currency, _ = r.String("currency")
	v, _ := r.Int("due_at")
	f.DueAt = time.Unix(v)
}
func (f *ForecastLine) IsNil() bool { return f == nil }

type Deps struct {
	IDs        model.IDGenerator
	TenantID   string
	LowStock   LowStockReader
	Forecast   AgreementForecastReader
	Subscriber events.Subscriber
}
