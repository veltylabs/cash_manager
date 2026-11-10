package tests

import (
	"testing"

	cashmanager "github.com/veltylabs/cash_manager"
	"webtyp.com/crud"
	"webtyp.com/events"
	"webtyp.com/events/mock"
	"webtyp.com/model"
	"webtyp.com/money"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	"webtyp.com/time"
	"webtyp.com/unixid"
)

// storedFacts reads every stored fact through the crud read the screens use.
func storedFacts(t *testing.T, m *cashmanager.Module) cashmanager.StoredFactList {
	t.Helper()
	var list cashmanager.StoredFactList
	if err := call("u1", nil, m, "stored_fact", model.Read, &crud.ReadArgs{}, &list); err != nil {
		t.Fatalf("read stored facts: %v", err)
	}
	return list
}

func TestAccept_SameFactTwiceIsStoredOnce(t *testing.T) {
	m := setup()
	fact := cashFact("src1", 1000, "u1", time.NowUnix())
	mustAccept(t, m, fact)
	mustAccept(t, m, fact)
	if n := len(storedFacts(t, m)); n != 1 {
		t.Errorf("a fact that arrives twice is stored once, got %d rows", n)
	}
}

func TestAccept_InvalidFactIsRefused(t *testing.T) {
	m := setup()
	if err := m.Accept(money.Fact{}); err == nil {
		t.Fatal("an empty fact must be refused")
	}
	if n := len(storedFacts(t, m)); n != 0 {
		t.Errorf("nothing is stored for a refused fact, got %d rows", n)
	}
}

func TestAccept_FeeFactAppearsInTheThirtyDayView(t *testing.T) {
	m := setup()
	now := time.NowUnix()
	fee := money.Fact{
		Kind: money.Accrued, Side: money.Outflow, Amount: money.Money{Amount: 447, Currency: "CLP"}, Scope: "test",
		Source: money.Ref{Type: "payment_fee", ID: "p1"}, At: now, DueAt: now + time.Day, Means: money.MeansCard, Actor: "sys",
	}
	mustAccept(t, m, fee)
	var list cashmanager.ThirtyDayList
	err := call("u1", nil, m, "thirty_day", model.Read, &crud.ReadArgs{Filter: &cashmanager.ThirtyDay{From: now - 100, To: now + 2*time.Day}}, &list)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].RowType != "fact" || list[0].Amount != 447 {
		t.Errorf("expected the fee as one fact row of 447, got %+v", list)
	}
}

func TestAccept_ArrivesThroughTheMoneyTopic(t *testing.T) {
	broker := &mock.Broker{}
	ids, _ := unixid.NewUnixID()
	m := cashmanager.New(orm.New(mem.New()), cashmanager.Deps{IDs: ids, TenantID: "t1", Subscriber: broker})
	fact := cashFact("src2", 1000, "u1", time.NowUnix())
	broker.Publish(events.Event{Topic: "money.fact", Payload: &fact})
	if n := len(storedFacts(t, m)); n != 1 {
		t.Errorf("a fact published on the money topic is stored, got %d rows", n)
	}
}
