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
		Means:      money.MeansCash,
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
		Means:      money.MeansCash,
	}

	broker.Publish(events.Event{Topic: "money.fact", Payload: &fact})
}

func TestCloseShiftCountsOnlyItsCashiersCash(t *testing.T) {
	_, broker, caller := setup()

	errCh := make(chan error, 1)
	var shift cashmanager.Shift
	caller.Call("cashmanager.OpenShift", &cashmanager.OpenShiftArgs{
		TenantId:    "t1",
		CashierId:   "c1",
		OpeningCash: 1000,
		Currency:    "CLP",
	}, &shift, func(err error) { errCh <- err })

	if err := <-errCh; err != nil {
		t.Fatalf("OpenShift failed: %v", err)
	}

	openedAt := shift.OpenedAt

	facts := []money.Fact{
		{
			Kind:   money.Charged,
			Side:   money.Inflow,
			Amount: money.Money{Amount: 5000, Currency: "CLP"},
			Scope:  "test",
			Source: money.Ref{Type: "consultation", ID: "cons-1"},
			At:     openedAt,
			Means:  money.MeansCash,
			Actor:  "c1",
		},
		{
			Kind:   money.Charged,
			Side:   money.Inflow,
			Amount: money.Money{Amount: 7000, Currency: "CLP"},
			Scope:  "test",
			Source: money.Ref{Type: "consultation", ID: "cons-2"},
			At:     openedAt,
			Means:  money.MeansCard,
			Actor:  "c1",
		},
		{
			Kind:   money.Charged,
			Side:   money.Inflow,
			Amount: money.Money{Amount: 3000, Currency: "CLP"},
			Scope:  "test",
			Source: money.Ref{Type: "consultation", ID: "cons-3"},
			At:     openedAt,
			Means:  money.MeansCash,
			Actor:  "c2",
		},
		{
			Kind:   money.Refunded,
			Side:   money.Outflow,
			Amount: money.Money{Amount: 500, Currency: "CLP"},
			Scope:  "test",
			Source: money.Ref{Type: "consultation", ID: "cons-4"},
			At:     openedAt,
			Means:  money.MeansCash,
			Actor:  "c1",
		},
	}

	for _, f := range facts {
		broker.Publish(events.Event{Topic: money.Topic, Payload: &f})
	}

	var closedShift cashmanager.Shift
	caller.Call("cashmanager.CloseShift", &cashmanager.CloseShiftArgs{
		TenantId:           "t1",
		Id:                 shift.Id,
		ClosingCashCounted: 5500,
	}, &closedShift, func(err error) { errCh <- err })

	if err := <-errCh; err != nil {
		t.Fatalf("CloseShift failed: %v", err)
	}

	if closedShift.ClosingCashExpected != 5500 {
		t.Errorf("ClosingCashExpected: got %d, want 5500", closedShift.ClosingCashExpected)
	}
	if closedShift.Difference != 0 {
		t.Errorf("Difference: got %d, want 0", closedShift.Difference)
	}
}

func TestFeeFactIsStored(t *testing.T) {
	m, _, caller := setup()

	now := time2.NowUnix()
	dueAt := now + 86400

	feeFact := money.Fact{
		Kind:   money.Accrued,
		Side:   money.Outflow,
		Amount: money.Money{Amount: 447, Currency: "CLP"},
		Scope:  "test",
		Source: money.Ref{Type: "payment_fee", ID: "p1"},
		At:     now,
		DueAt:  dueAt,
		Means:  money.MeansCard,
		Actor:  "sys",
	}

	if err := m.Accept(feeFact); err != nil {
		t.Fatalf("Accept fee fact failed: %v", err)
	}

	errCh := make(chan error, 1)
	var thirtyDayResp cashmanager.ThirtyDayResponse
	caller.Call("cashmanager.ThirtyDay", &cashmanager.ListStoredFactsArgs{
		TenantId: "t1",
		From:     now - 100,
		To:       dueAt + 100,
	}, &thirtyDayResp, func(err error) { errCh <- err })

	if err := <-errCh; err != nil {
		t.Fatalf("ThirtyDay failed: %v", err)
	}

	if len(thirtyDayResp.Facts) != 1 {
		t.Fatalf("expected 1 fact listed by ThirtyDay, got %d", len(thirtyDayResp.Facts))
	}
	got := thirtyDayResp.Facts[0]
	if got.SourceType != "payment_fee" || got.SourceId != "p1" || got.Means != string(money.MeansCard) {
		t.Errorf("unexpected fact: %+v", got)
	}
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
