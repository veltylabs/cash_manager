//go:build wasm

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

	deps := cashmanager.Deps{
		IDs:        unixid.NewGenerator(),
		TenantID:   "t1",
		LowStock:   &fakeLowStock{},
		Forecast:   &fakeForecast{},
		Subscriber: broker,
	}

	module := cashmanager.New(db, deps)

	reg := loopback.NewRegistry()
	module.MountOperations(reg)

	// Normally this connects to DOM
}
