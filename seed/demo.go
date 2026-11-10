//go:build wasm

package seed

import (
	cashmanager "github.com/veltylabs/cash_manager"
	"github.com/veltylabs/cash_manager/migrate"
	"github.com/veltylabs/cash_manager/ui"
	"webtyp.com/crud"
	"webtyp.com/ddl"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	"webtyp.com/unixid"
	"webtyp.com/view"
)

// Demo crea una demostración ejecutable del módulo cash_manager sobre una base en memoria.
//
// Ejemplo: demo, err := seed.Demo()
func Demo() (view.Demo, error) {
	db := orm.New(mem.New())
	ids, err := unixid.NewUnixID()
	if err != nil {
		return view.Demo{}, err
	}
	mod := cashmanager.New(db, cashmanager.Deps{IDs: ids, TenantID: "demo"})
	return view.Demo{
		Label:      ui.Label,
		Timezone:   "America/Santiago",
		ActingAs:   "u1",
		Modules:    []crud.Module{mod},
		Migrations: []ddl.Step{{Name: cashmanager.ModelName, Run: migrate.Migrate}},
		Screens: func(d view.BrowserDeps) []view.Screen {
			return ui.Screens(d.Caller)
		},
	}, nil
}
