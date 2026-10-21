//go:build browser

package main

import (
	"github.com/veltylabs/cash_manager"
	"webtyp.com/events/mock"
	"webtyp.com/orm"
	"webtyp.com/router/loopback"
	"webtyp.com/storage/mem"
	"webtyp.com/time"
	"webtyp.com/unixid"
)

type fakeLowStock struct{}

func (f *fakeLowStock) LowStock(tenantId string) ([]cashmanager.LowStockItem, error) {
	return []cashmanager.LowStockItem{}, nil
}

type fakeForecast struct{}

func (f *fakeForecast) PendingAccruals(tenantId string, from, to time.Unix) ([]cashmanager.ForecastLine, error) {
	return []cashmanager.ForecastLine{}, nil
}

func main() {
	db := orm.New(mem.New())
	broker := &mock.Broker{}
	idGen, _ := unixid.NewUnixID()

	deps := cashmanager.Deps{
		IDs:        idGen,
		TenantID:   "t1",
		LowStock:   &fakeLowStock{},
		Forecast:   &fakeForecast{},
		Subscriber: broker,
	}

	module := cashmanager.New(db, deps)
	_ = loopback.WithTenant("t1", module)

	// Normally this connects to DOM
}
