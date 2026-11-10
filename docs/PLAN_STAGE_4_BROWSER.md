← [Stage 3](PLAN_STAGE_3_RULES.md) | Next → [Stage 5](PLAN_STAGE_5_TESTS_DOCS.md)

# Stage 4 — delete `ops.go`; `view.go`; demo

## 4.1 Delete

`ops.go` (with the `TodayResponse`, `ReorderResponse`, `ThirtyDayResponse` types and the six handlers it holds), `web/` (`rm -r web`), and every import only they used (`webtyp.com/router`, `webtyp.com/rpc`, `webtyp.com/view` in `module.go`). `var _ router.OperationModule` is already gone (Stage 2). Verify: `grep -rn "reg.Operation\|router.Context\|rpc.WriteError" --include=*.go . | grep -v _test` → empty.

## 4.2 `view.go` — replace the whole file with exactly

```go
package cashmanager

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

const titleCashManager = "Cash management"

// NewView builds the Presenter that lists the stored cash facts.
//
// Example: presenter := cashmanager.NewView(caller, view.WithDescription(descCash))
func NewView(caller router.Caller, opts ...view.Option) view.Presenter {
	lister := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, Page: view.Page{First: 50, Size: 50, More: true}},
		func() model.TableSlice { return &StoredFactList{} })
	allOpts := append([]view.Option{view.WithTitle(titleCashManager)}, opts...)
	return view.New(lister, &StoredFact{}, allOpts...)
}
```
(`OpListStoredFacts` is deleted; `view.Ops` has no `List`, `Save`, `Delete` or `Allow`.) If `view.NewCallerLister`/`view.Ops` differ in `view@v0.6.74`, run `go doc webtyp.com/view Ops` and copy https://github.com/veltylabs/patient_directory/blob/main/view.go; name the difference in the PR.

## 4.3 Demo (replaces `web/browser.go`): `seed/demo.go`, `//go:build wasm`, written exactly

```go
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
```
Read `ui/browser.go` and the deleted `web/browser.go` (`git show main:web/browser.go`) for the exact names of the label and the screens function and for what the old demo passed to them; use the existing names (add none). `migrate/migrate.go`: its `Sync(...)` list is `&cashmanager.Shift{}, &cashmanager.StoredFact{}` (the read models are not tables).

`gotest` compiles on every destination now except the tests Stage 5 rewrites. Commit and push.
