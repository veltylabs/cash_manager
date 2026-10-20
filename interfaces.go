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

type AgreementForecastReader interface {
	PendingAccruals(tenantId string, from, to time.Unix) ([]ForecastLine, error)
}

type ForecastLine struct {
	ThirdPartyName string    `json:"third_party_name"`
	Amount         int64     `json:"amount"`
	Currency       string    `json:"currency"`
	DueAt          time.Unix `json:"due_at"`
}

type Deps struct {
	IDs        model.IDGenerator
	TenantID   string
	LowStock   LowStockReader
	Forecast   AgreementForecastReader
	Subscriber events.Subscriber
}
