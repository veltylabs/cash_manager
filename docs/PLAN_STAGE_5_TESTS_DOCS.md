← [Stage 4](PLAN_STAGE_4_BROWSER.md)

# Stage 5 — tests, `docs/TESTS.md`, documentation

## 5.1 Existing tests

`TestAcceptIdempotency`, `TestAcceptInvalidFact`, `TestFeeFactIsStored`, `TestMoneySubscribeWiring`, `TestErrorsStatuses`, `TestUIScreens` and `TestItemsDeclaredOnFields` keep testing `Accept`, the errors and the screens: they stay. `setup()` in `tests/cash_manager_test.go` builds the module with `cashmanager.New(db, Deps{IDs, TenantID: "t1", Subscriber})` and returns `router/loopback` callers: replace the returned caller by nothing (`func setup() (*cashmanager.Module, *mock.Broker)`) and use `call` below. A test that drives the removed handlers by a string (`caller.Call("cashmanager.OpenShift", …)`) is rewritten as one of the tests below or deleted (`TestShiftLifecycle`, `TestCloseShiftCountsOnlyItsCashiersCash`, `TestNilSafetyReaders`, `TestCashManager_ListAcceptsListArgs`, `TestTenantIsolation` map to rows of 5.2). `tests/ui_view_browser_test.go` (`browser` tag) must compile against the `router.Operation` signature of `Call`/`Dispatch` (`Call(op router.Operation, args model.Encodable, into model.Decodable, done func(error))`).

## 5.2 New tests, `tests/crud_test.go` (package `tests`), with these helpers written exactly

```go
package tests

import (
	"testing"

	cashmanager "github.com/veltylabs/cash_manager"
	"webtyp.com/crud"
	"webtyp.com/events"
	"webtyp.com/model"
	"webtyp.com/money"
	"webtyp.com/router"
	"webtyp.com/time"
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
```
**The acting user is the cashier:** a shift opened as `u1` is `u1`'s; `cash_today` read as `u2` is empty. The first thing a test does with a callback error is assert it (`if err != nil { t.Fatalf(...) }`) BEFORE reading the result: an unchecked error looks like «empty result».

| Test (all in `tests/crud_test.go`) | setup → call → expected |
|---|---|
| `TestShift_OneOpenPerCashier` | `openShift(t, m, "u1")` → nil; again → `model.KindOf(err) == model.ErrorConflict` |
| `TestShift_OnlyOwnCashierCloses` | `u1` opens; `call("u2", nil, m, "cash_shift", model.Update, &crud.UpdateArgs{IDs: []string{id}, Fields: []string{"closing_cash_counted"}, Record: &cashmanager.Shift{ClosingCashCounted: 1}}, nil)` → `failedField(err) == "cashier_id"`; `myShift(t, m, "u1")` still has `Status == "open"` (`id` = `myShift(...)[0].Id`) |
| `TestShift_CloseComputesDifference` | `u1` opens (opening 1000); `m.Accept(cashFact("f1", 5000, "u1", openedAt))` with `openedAt = myShift(...)[0].OpenedAt`; `u1` closes with `ClosingCashCounted: 5900` → no error; reading the shift (`call("u1", nil, m, "cash_shift", model.Read, &crud.ReadArgs{ID: id}, &shiftList)` with `var shiftList cashmanager.ShiftList`) → `Status == "closed"`, `ClosingCashExpected == 6000`, `Difference == -100`, `ClosingCashCounted == 5900` |
| `TestShift_CloseCountsOnlyItsCashiersCash` | as above plus a card fact, a fact of another cashier (`actor "u2"`) and a refund of 500 (`Kind: money.Refunded`, `Side: money.Outflow`) of `u1`'s cash → `ClosingCashExpected == 1000 + 5000 - 500` |
| `TestShift_CloseOnlyCountedField` | `Fields: []string{"status"}` → `failedField(err) == "status"` |
| `TestReadModel_TodayIsCallers` | `u1` and `u2` open shifts; `myShift(t, m, "u1")` has one row with `CashierId == "u1"`; `myShift(t, m, "u3")` is empty |
| `TestReadModel_ThirtyDayRange` | three facts with `DueAt` inside, before and after the range `[from, to]`; `call("u1", nil, m, "thirty_day", model.Read, &crud.ReadArgs{Filter: &cashmanager.ThirtyDay{From: from, To: to}}, &list)` with `var list cashmanager.ThirtyDayList` → only the inside fact, `RowType == "fact"` |
| `TestReadModel_ThirtyDayNeedsRange` | the same read with an empty `ThirtyDay{}` → `failedField(err) == "from"` |
| `TestReadModel_ReorderRows` | `Deps.LowStock` = a double returning two `LowStockItem` → `reorder` read has two rows with their names |
| `TestFact_ReadOnly` | `call("u1", nil, m, "stored_fact", model.Create, &crud.CreateArgs{Records: []model.Model{&cashmanager.StoredFact{}}}, nil)` → error text contains `unknown operation` |
| `TestMount_MatchesDefinition` | `crud.Mount(nil, nil, nil, m)` returns nil |
| `TestCreate_PublishesChange` | open a shift with a publisher given as the FIRST argument: one event on `crud.Topic(cashmanager.ModelName, (&cashmanager.Shift{}).ModelName(), model.Create)` whose payload is `*crud.Change` |
| `TestDemo_Valid` (`tests/demo_test.go`, `//go:build wasm`) | copy https://github.com/veltylabs/patient_directory/blob/main/tests/demo_test.go with the import path `github.com/veltylabs/cash_manager/seed` |

Facts need an `At` inside the shift window: `openedAt` as above, never `time.Now()` of the stdlib. `TestShift_CloseComputesDifference` is the one that failed twice: if the closed shift still shows `open`, the rule is not writing the derived columns with `tx.UpdateFields` (Stage 3.3).

## 5.3 `docs/TESTS.md`

One row per `func Test…` of the repo (old and new), header `| Test | Rule |`, separator `|---|---|`, each row `| [TestName](../tests/file_test.go) | one line rule |` — the link IS the file; **no File column** (it fails the `gotest` tests-index check). Rules in Spanish.

## 5.4 Documentation (Spanish, `veltylabs`)

`langc sync .`: zero pending for `es`, zero reuse warnings. `README.md` ≤150 lines, four sections in this order: `## ¿Por qué existe?` (cash shifts and the facts of money events; **Sin esto vs. Con esto**; first mentions linked), `## Antes de comenzar` (Go, `GOPRIVATE`, a `*orm.DB` with the tables of `migrate`, `Deps.IDs`, `Deps.TenantID`), `## Guía paso a paso` (Step 1 create the module over `orm.New(mem.New())` — a complete `main` importing `"webtyp.com/fmt"` and printing `mod.ModelName()` → `cash_manager`; Step 2 open and close a shift, linking [tests/crud_test.go](tests/crud_test.go); Step 3 open the demo, linking [seed/demo.go](seed/demo.go)), `## Mapa de documentación` (table `Orden | Documento | Qué pregunta responde y cuándo leerlo`). Include the operations table of the mapping in `PLAN.md` and a permissions table with the resources `cash_shift`, `stored_fact`, `cash_today`, `reorder`, `thirty_day` (roles must be granted them). `docs/ARCHITECTURE.md`: four sections (`## 1. Visión general en un minuto`, `## 2. Piezas y estructura`, `## 3. Flujos paso a paso`, `## 4. Contratos y límites de seguridad`): why closing a shift writes its derived columns itself, why the read models are flat. Translate to Spanish every README, `docs/*`, diagram and code comment still in English; exported symbols without a usage example get one. Permanent docs never link `docs/PLAN.md` or its stage files.

Run `gotest` and `langc sync .` one last time, read their output, check the **Delivery check** of `PLAN.md`, commit, push and make sure the pull request is open.
