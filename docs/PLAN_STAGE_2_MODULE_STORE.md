← [Stage 1](PLAN_STAGE_1_TOOLS_DEFINITION.md) | Next → [Stage 3](PLAN_STAGE_3_RULES.md)

# Stage 2 — `Module.store`, helpers, `Accept`

## 2.1 `module.go` — replace the whole file with exactly

```go
package cashmanager

import (
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/money"
	"webtyp.com/orm"
	"webtyp.com/storage"
)

var _ money.Sink = (*Module)(nil)

// Module is the cash module: shifts of the cashiers and the facts money events leave behind.
type Module struct {
	db    *orm.DB
	store *orm.Store
	deps  Deps
}

// New connects the module to a database. Deps.IDs and Deps.TenantID are required.
//
// Example: m := cashmanager.New(orm.New(mem.New()), cashmanager.Deps{IDs: ids, TenantID: "t1"})
func New(db *orm.DB, deps Deps) *Module {
	if deps.IDs == nil {
		panic("cash_manager: missing IDs dependency")
	}
	m := &Module{db: db, store: orm.TenantStore(db, deps.TenantID, deps.IDs), deps: deps}
	if deps.Subscriber != nil {
		money.Subscribe(deps.Subscriber, m, func(err error) {
			fmt.Println("cash_manager: failed to accept fact:", err)
		})
	}
	return m
}

// Accept stores a money fact once: a fact already stored is not an error.
//
// Example: err := m.Accept(money.Fact{Kind: money.Charged, ...})
func (m *Module) Accept(f money.Fact) error {
	if err := f.Validate(); err != nil {
		return err
	}
	fact := &StoredFact{
		Id:         f.Key(),
		Kind:       string(f.Kind),
		Side:       int64(f.Side),
		Amount:     f.Amount.Amount,
		Currency:   string(f.Amount.Currency),
		Scope:      f.Scope,
		SourceType: f.Source.Type,
		SourceId:   f.Source.ID,
		Party:      f.Party,
		At:         f.At,
		DueAt:      f.DueAt,
		Means:      string(f.Means),
		Actor:      f.Actor,
	}
	if err := model.ValidateFields(model.ActionCreate, fact); err != nil {
		return err
	}
	err := m.store.Atomic(func(tx crud.Tx) error { return tx.Create(fact) })
	if err != nil && storage.IsDuplicate(err) {
		return nil // idempotent: the same fact arrives twice
	}
	return err
}
```
Add the missing `"webtyp.com/crud"` import. The old `ModelName()` and `MountOperations()` of this file are deleted: `definitionc` writes both (Stage 3). The `tenant_id` of the fact is stamped by the store; do not set it. If `model.ValidateFields` or `storage.IsDuplicate` has another signature, run `go doc webtyp.com/model ValidateFields` / `go doc webtyp.com/storage IsDuplicate` and name the difference in the PR; the reference `import.go` of `patient_directory` shows the first one in use.

## 2.2 `store.go` (new)

```go
package cashmanager

import "webtyp.com/crud"

// StoreShift is the store of the shifts.
func (m *Module) StoreShift() crud.Store { return m.store }

// StoreStoredFact is the store of the stored facts.
func (m *Module) StoreStoredFact() crud.Store { return m.store }
```

## 2.3 `rows.go` (new): the helpers every rule uses, written exactly

```go
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
```
`Shift_`, `StoredFact_` and the types come from the generated `definition_gen.go` (Stage 3): this file compiles after Stage 3. Commit it with Stage 3.
