package tests

import (
	"strings"
	"testing"

	cashmanager "github.com/veltylabs/cash_manager"
	"webtyp.com/crud"
	"webtyp.com/events"
	"webtyp.com/model"
	"webtyp.com/money"
	"webtyp.com/orm"
	"webtyp.com/router"
	"webtyp.com/storage/mem"
	"webtyp.com/time"
	"webtyp.com/unixid"
)

// call runs one operation of the module as `user` through the loopback the demo uses and answers the callback error.
// The module name is ALWAYS cashmanager.ModelName, never a literal.
func call(user string, pub events.Publisher, mod *cashmanager.Module, entity string, action model.Action, args model.Encodable, into model.Decodable) error {
	var callErr error
	crud.LoopbackActingAs(pub, user, mod).Call(
		router.Operation{Module: cashmanager.ModelName, Entity: entity, Action: action},
		args, into, func(err error) { callErr = err })
	return callErr
}

// failedField answers the field a refusal points at, or "".
func failedField(err error) string {
	if f, ok := err.(interface{ FailedField() string }); ok {
		return f.FailedField()
	}
	return ""
}

// openShift opens a shift as `user` with 1000 of opening cash.
func openShift(t *testing.T, mod *cashmanager.Module, user string) error {
	t.Helper()
	return call(user, nil, mod, "cash_shift", model.Create,
		&crud.CreateArgs{Records: []model.Model{&cashmanager.Shift{OpeningCash: 1000, Currency: "CLP"}}}, nil)
}

// myShift answers the open shift of `user` through the read model.
func myShift(t *testing.T, mod *cashmanager.Module, user string) cashmanager.TodayShiftList {
	t.Helper()
	var list cashmanager.TodayShiftList
	if err := call(user, nil, mod, "cash_today", model.Read, &crud.ReadArgs{}, &list); err != nil {
		t.Fatalf("cash_today as %s: %v", user, err)
	}
	return list
}

func cashFact(key string, amount int64, actor string, at time.Unix) money.Fact {
	return money.Fact{
		Kind: money.Charged, Side: money.Inflow, Amount: money.Money{Amount: amount, Currency: "CLP"}, Scope: "test",
		Source: money.Ref{Type: "consultation", ID: key}, At: at, Means: money.MeansCash, Actor: actor,
	}
}

func setup() *cashmanager.Module {
	db := orm.New(mem.New())
	ids, _ := unixid.NewUnixID()
	return cashmanager.New(db, cashmanager.Deps{IDs: ids, TenantID: "t1"})
}

func TestShift_OneOpenPerCashier(t *testing.T) {
	m := setup()
	if err := openShift(t, m, "u1"); err != nil {
		t.Fatalf("openShift: %v", err)
	}
	err := openShift(t, m, "u1")
	if model.KindOf(err) != model.ErrorConflict {
		t.Errorf("expected ErrorConflict, got %v", err)
	}
}

func TestShift_OnlyOwnCashierCloses(t *testing.T) {
	m := setup()
	openShift(t, m, "u1")
	id := myShift(t, m, "u1")[0].Id
	err := call("u2", nil, m, "cash_shift", model.Update, &crud.UpdateArgs{IDs: []string{id}, Fields: []string{"closing_cash_counted"}, Record: &cashmanager.Shift{Id: id, ClosingCashCounted: 1}}, nil)
	if failedField(err) != "cashier_id" {
		t.Errorf("expected failed field cashier_id, got %q", failedField(err))
	}
	if myShift(t, m, "u1")[0].Status != "open" {
		t.Errorf("shift is not open")
	}
}

func TestShift_CloseComputesDifference(t *testing.T) {
	m := setup()
	openShift(t, m, "u1")
	openedAt := myShift(t, m, "u1")[0].OpenedAt
	mustAccept(t, m, cashFact("f1", 5000, "u1", openedAt))
	id := myShift(t, m, "u1")[0].Id
	err := call("u1", nil, m, "cash_shift", model.Update, &crud.UpdateArgs{IDs: []string{id}, Fields: []string{"closing_cash_counted"}, Record: &cashmanager.Shift{ClosingCashCounted: 5900}}, nil)
	if err != nil {
		t.Fatalf("close shift: %v", err)
	}
	var shiftList cashmanager.ShiftList
	if err := call("u1", nil, m, "cash_shift", model.Read, &crud.ReadArgs{ID: id}, &shiftList); err != nil {
		t.Fatalf("read shift: %v", err)
	}
	shift := shiftList[0]
	if shift.Status != "closed" {
		t.Errorf("expected closed, got %s", shift.Status)
	}
	if shift.ClosingCashExpected != 6000 {
		t.Errorf("expected 6000, got %d", shift.ClosingCashExpected)
	}
	if shift.Difference != -100 {
		t.Errorf("expected -100, got %d", shift.Difference)
	}
	if shift.ClosingCashCounted != 5900 {
		t.Errorf("expected 5900, got %d", shift.ClosingCashCounted)
	}
}

func TestShift_CloseCountsOnlyItsCashiersCash(t *testing.T) {
	m := setup()
	openShift(t, m, "u1")
	openedAt := myShift(t, m, "u1")[0].OpenedAt
	mustAccept(t, m, cashFact("f1", 5000, "u1", openedAt))

	cardFact := cashFact("f2", 1000, "u1", openedAt)
	cardFact.Means = money.MeansCard
	mustAccept(t, m, cardFact)

	mustAccept(t, m, cashFact("f3", 1000, "u2", openedAt))

	refundFact := cashFact("f4", 500, "u1", openedAt)
	refundFact.Kind = money.Refunded
	refundFact.Side = money.Outflow
	mustAccept(t, m, refundFact)

	id := myShift(t, m, "u1")[0].Id
	call("u1", nil, m, "cash_shift", model.Update, &crud.UpdateArgs{IDs: []string{id}, Fields: []string{"closing_cash_counted"}, Record: &cashmanager.Shift{ClosingCashCounted: 5500}}, nil)

	var shiftList cashmanager.ShiftList
	call("u1", nil, m, "cash_shift", model.Read, &crud.ReadArgs{ID: id}, &shiftList)
	shift := shiftList[0]

	expected := int64(1000 + 5000 - 500)
	if shift.ClosingCashExpected != expected {
		t.Errorf("expected %d, got %d", expected, shift.ClosingCashExpected)
	}
}

func TestShift_CloseOnlyCountedField(t *testing.T) {
	m := setup()
	openShift(t, m, "u1")
	id := myShift(t, m, "u1")[0].Id
	err := call("u1", nil, m, "cash_shift", model.Update, &crud.UpdateArgs{IDs: []string{id}, Fields: []string{"status"}, Record: &cashmanager.Shift{Status: "closed"}}, nil)
	if model.KindOf(err) != model.ErrorInvalid {
		t.Errorf("a server-only field must be refused by crud, got %v", err)
	}
}

func TestReadModel_TodayIsCallers(t *testing.T) {
	m := setup()
	openShift(t, m, "u1")
	openShift(t, m, "u2")
	if myShift(t, m, "u1")[0].CashierId != "u1" {
		t.Errorf("expected u1")
	}
	if len(myShift(t, m, "u3")) != 0 {
		t.Errorf("expected empty")
	}
}

func TestReadModel_ThirtyDayRange(t *testing.T) {
	m := setup()
	now := time.NowUnix()
	from := now - 100
	to := now + 100
	due := money.Fact{
		Kind: money.Accrued, Side: money.Outflow, Amount: money.Money{Amount: 100, Currency: "CLP"}, Scope: "test",
		Source: money.Ref{Type: "accrual", ID: "f1"}, At: time.Unix(now), DueAt: time.Unix(now), // only an accrual carries a due time
	}
	mustAccept(t, m, due)
	var list cashmanager.ThirtyDayList
	call("u1", nil, m, "thirty_day", model.Read, &crud.ReadArgs{Filter: &cashmanager.ThirtyDay{From: from, To: to}}, &list)
	if len(list) != 1 || list[0].RowType != "fact" {
		t.Errorf("expected 1 fact row")
	}
}

func TestReadModel_ThirtyDayNeedsRange(t *testing.T) {
	m := setup()
	var list cashmanager.ThirtyDayList
	err := call("u1", nil, m, "thirty_day", model.Read, &crud.ReadArgs{Filter: &cashmanager.ThirtyDay{}}, &list)
	if failedField(err) != "from" {
		t.Errorf("expected failed field from")
	}
}

type lowStockDouble struct{}

func (l *lowStockDouble) LowStock(tenantId string) ([]cashmanager.LowStockItem, error) {
	return []cashmanager.LowStockItem{
		{ItemId: "i1", ItemName: "Item 1", Total: 10, MinQuantity: 5},
		{ItemId: "i2", ItemName: "Item 2", Total: 20, MinQuantity: 15},
	}, nil
}

func TestReadModel_ReorderRows(t *testing.T) {
	m := setup()
	m.Deps().LowStock = &lowStockDouble{}
	var list cashmanager.ReorderList
	call("u1", nil, m, "reorder", model.Read, &crud.ReadArgs{}, &list)
	if len(list) != 2 {
		t.Errorf("expected 2 rows")
	}
	if list[0].ItemName != "Item 1" || list[1].ItemName != "Item 2" {
		t.Errorf("unexpected rows")
	}
}

func TestFact_ReadOnly(t *testing.T) {
	m := setup()
	err := call("u1", nil, m, "stored_fact", model.Create, &crud.CreateArgs{Records: []model.Model{&cashmanager.StoredFact{}}}, nil)
	if !strings.Contains(err.Error(), "unknown operation") && !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("expected unknown operation error, got %v", err)
	}
}

func TestMount_MatchesDefinition(t *testing.T) {
	m := setup()
	if err := crud.Mount(nil, nil, nil, m); err != nil {
		t.Errorf("mount failed: %v", err)
	}
}

type dummyPub struct {
	events []events.Event
}

func (d *dummyPub) Publish(e events.Event) { d.events = append(d.events, e) }
func (d *dummyPub) Start() error           { return nil }

func TestCreate_PublishesChange(t *testing.T) {
	m := setup()
	pub := &dummyPub{}
	call("u1", pub, m, "cash_shift", model.Create,
		&crud.CreateArgs{Records: []model.Model{&cashmanager.Shift{OpeningCash: 1000, Currency: "CLP"}}}, nil)
	if len(pub.events) == 0 {
		t.Fatalf("expected event")
	}
	e := pub.events[0]
	if e.Topic != crud.Topic(cashmanager.ModelName, (&cashmanager.Shift{}).ModelName(), model.Create) {
		t.Errorf("unexpected topic %s", e.Topic)
	}
	if _, ok := e.Payload.(*crud.Change); !ok {
		t.Errorf("unexpected payload type %T", e.Payload)
	}
}

// mustAccept stores a fact through the module and fails the test if it is refused.
func mustAccept(t *testing.T, m *cashmanager.Module, f money.Fact) {
	t.Helper()
	if err := m.Accept(f); err != nil {
		t.Fatalf("Accept: %v", err)
	}
}
