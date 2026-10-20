package tests

import (
	"webtyp.com/events"

	"testing"
	"time"

	"github.com/veltylabs/cash_manager"
	"webtyp.com/events/mock"
	"webtyp.com/money"
	"webtyp.com/orm"
	"webtyp.com/router/loopback"
	"webtyp.com/storage/mem"
	time2 "webtyp.com/time"
	"webtyp.com/router"
)

type mockID struct{ counter int }

func (m *mockID) NewID() string {
	m.counter++
	return "id123a" + string(rune(m.counter+64))
}

func setup() (*cashmanager.Module, *mock.Broker, router.Caller) {
	db := orm.New(mem.New())
	broker := &mock.Broker{}
	deps := cashmanager.Deps{
		IDs:        &mockID{},
		TenantID:   "t1",
		Subscriber: broker,
	}
	m := cashmanager.New(db, deps)
	return m, broker, loopback.ActingAs("t1", "user_1", m)
}

func TestAcceptIdempotency(t *testing.T) {
	m, _, _ := setup()

	fact := money.Fact{
		Kind:       money.Charged,
		Side:       1,
		Amount:     money.Money{Amount: 1000, Currency: "USD"},
		Scope:      "test",
		Source:     money.Ref{Type: "testsrc", ID: "src1"},
		At:         time2.Unix(time.Now().Unix()),
	}

	err := m.Accept(fact)
	if err != nil {
		t.Fatalf("first accept failed: %v", err)
	}

	err = m.Accept(fact)
	if err != nil {
		t.Fatalf("second accept failed: %v", err)
	}
}

func TestAcceptInvalidFact(t *testing.T) {
	m, _, _ := setup()
	fact := money.Fact{} // invalid missing fields

	err := m.Accept(fact)
	if err == nil {
		t.Fatal("expected error on invalid fact")
	}
}

func TestShiftLifecycle(t *testing.T) {
	_, _, caller := setup()

	// Create channels to receive results
	errCh := make(chan error, 1)

	var shift cashmanager.Shift
	caller.Call("cashmanager.OpenShift", &cashmanager.OpenShiftArgs{
		TenantId:    "t1",
		CashierId:   "c1",
		OpeningCash: 1000,
		Currency:    "USD",
	}, &shift, func(err error) { errCh <- err })

	if err := <-errCh; err != nil {
		t.Fatalf("OpenShift failed: %v", err)
	}

	var shift2 cashmanager.Shift
	caller.Call("cashmanager.OpenShift", &cashmanager.OpenShiftArgs{
		TenantId:    "t1",
		CashierId:   "c1",
		OpeningCash: 1000,
		Currency:    "USD",
	}, &shift2, func(err error) { errCh <- err })

	if err := <-errCh; err == nil {
		t.Fatal("expected ErrShiftAlreadyOpen")
	}

	var closedShift cashmanager.Shift
	caller.Call("cashmanager.CloseShift", &cashmanager.CloseShiftArgs{
		TenantId:           "t1",
		Id:                 shift.Id,
		ClosingCashCounted: 1000,
	}, &closedShift, func(err error) { errCh <- err })

	if err := <-errCh; err != nil {
		t.Fatalf("CloseShift failed: %v", err)
	}
}

func TestMoneySubscribeWiring(t *testing.T) {
	_, broker, _ := setup()

	fact := money.Fact{
		Kind:       money.Charged,
		Side:       1,
		Amount:     money.Money{Amount: 1000, Currency: "USD"},
		Scope:      "test",
		Source:     money.Ref{Type: "testsrc", ID: "src2"},
		At:         time2.Unix(time.Now().Unix()),
	}

	broker.Publish(events.Event{Topic: "money.fact", Payload: &fact})
}

func TestNilSafetyReaders(t *testing.T) {
	_, _, caller := setup()

	errCh := make(chan error, 1)

	var reorderResp cashmanager.ReorderResponse
	caller.Call("cashmanager.Reorder", nil, &reorderResp, func(err error) { errCh <- err })

	if err := <-errCh; err != nil {
		t.Fatalf("Reorder failed: %v", err)
	}

	var thirtyDayResp cashmanager.ThirtyDayResponse
	caller.Call("cashmanager.ThirtyDay", &cashmanager.ListStoredFactsArgs{
		TenantId: "t1",
		From:     time2.Unix(time.Now().Unix()),
		To:       time2.Unix(time.Now().AddDate(0, 1, 0).Unix()),
	}, &thirtyDayResp, func(err error) { errCh <- err })

	if err := <-errCh; err != nil {
		t.Fatalf("ThirtyDay failed: %v", err)
	}
}

func TestTenantIsolation(t *testing.T) {
	_, _, caller := setup()

	errCh := make(chan error, 1)

	var shift cashmanager.Shift
	caller.Call("cashmanager.OpenShift", &cashmanager.OpenShiftArgs{
		TenantId:    "t1",
		CashierId:   "c1",
		OpeningCash: 1000,
		Currency:    "USD",
	}, &shift, func(err error) { errCh <- err })

	if err := <-errCh; err != nil {
		t.Fatalf("OpenShift t1 failed: %v", err)
	}

	var todayResp cashmanager.TodayResponse
	caller.Call("cashmanager.GetToday", &cashmanager.GetTodayArgs{
		TenantId:  "t2",
		CashierId: "c1",
	}, &todayResp, func(err error) { errCh <- err })

	if err := <-errCh; err != nil {
		t.Fatalf("GetToday t2 failed: %v", err)
	}

	if todayResp.Shift != nil && todayResp.Shift.Id != "" {
		t.Fatal("tenant 2 saw tenant 1's shift")
	}
}
