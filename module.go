package cashmanager

import (
	"strings"


	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/money"
	"webtyp.com/orm"
	"webtyp.com/router"
)

var _ money.Sink = (*Module)(nil)
var _ router.OperationModule = (*Module)(nil)

type Module struct {
	db   *orm.DB
	deps Deps
}

func New(db *orm.DB, deps Deps) *Module {
	if deps.IDs == nil {
		panic("cashmanager: missing IDs dependency")
	}
	m := &Module{
		db:   db,
		deps: deps,
	}

	if deps.Subscriber != nil {
		money.Subscribe(deps.Subscriber, m, func(err error) {
			fmt.Println("cashmanager: failed to accept fact:", err)
		})
	}

	return m
}

func (m *Module) ModelName() string {
	return "cashmanager"
}

func (m *Module) MountOperations(reg router.OperationRegistry) {
	reg.Operation("OpenShift", m.openShift).Requires(model.Resource(m.ModelName()), model.Create).Accepts(&OpenShiftArgs{})
	reg.Operation("CloseShift", m.closeShift).Requires(model.Resource(m.ModelName()), model.Update).Accepts(&CloseShiftArgs{})
	reg.Operation("GetToday", m.getToday).Requires(model.Resource(m.ModelName()), model.Read).Accepts(&GetTodayArgs{})
	reg.Operation("ListStoredFacts", m.listStoredFacts).Requires(model.Resource(m.ModelName()), model.Read).Accepts(&ListStoredFactsArgs{})
	reg.Operation("Reorder", m.reorder).Requires(model.Resource(m.ModelName()), model.Read).Accepts(nil)
	reg.Operation("ThirtyDay", m.thirtyDay).Requires(model.Resource(m.ModelName()), model.Read).Accepts(&ListStoredFactsArgs{})
}

func (m *Module) Accept(f money.Fact) error {
	if err := f.Validate(); err != nil {
		return err
	}

	fact := &StoredFact{
		Id:         f.Key(),
		TenantId:   m.deps.TenantID, // using module's tenant
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
	}

	if err := fact.Validate(model.ActionCreate); err != nil {
		return err
	}

	err := m.db.Create(fact)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil // idempotent
		}
		return err
	}
	return nil
}
