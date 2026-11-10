package cashmanager

import (
	"webtyp.com/crud"
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

func (m *Module) Deps() *Deps { return &m.deps }

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
		TenantId:   m.deps.TenantID,
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
