← [Stage 2](PLAN_STAGE_2_MODULE_STORE.md) | Next → [Stage 4](PLAN_STAGE_4_BROWSER.md)

# Stage 3 — generate, then write the rules

## 3.1 Generate

Delete `model_orm.go` and `definition_crud.go` if they exist (not `ops.go` yet: Stage 4). Run `definitionc` in the repo root: it writes `definition_gen.go` (types `Shift`, `StoredFact`, `TodayShift`, `Reorder`, `ThirtyDay` and their lists, the constants `Shift_`, `StoredFact_`, `ThirtyDay_`…, `ModelName()` = `"cash_manager"`, `Definition()`, `MountOperations()`) and one skeleton per action declared `AddMyBusinessRule`: `rule_cash_shift_create.go`, `rule_cash_shift_update.go`, `rule_cash_today_read.go`, `rule_reorder_read.go`, `rule_thirty_day_read.go`. Then repeat the `go get` line of Stage 1 and `go mod tidy`. **Keep the signatures the skeletons have; replace only the bodies.** `rows.go` and `store.go` compile now.

## 3.2 `rule_cash_shift_create.go` — open a shift (the cashier is the caller; one open shift per cashier)

```go
func (p *Shift) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	_, found, err := firstRow(tx, func() model.Model { return &Shift{} },
		storage.Eq(Shift_.CashierId, a.UserID), storage.Eq(Shift_.Status, ShiftOpen))
	if err != nil {
		return 0, err
	}
	if found {
		return 0, ErrShiftAlreadyOpen
	}
	p.CashierId = a.UserID
	p.Status = ShiftOpen
	p.OpenedAt = time.NowUnix()
	return crud.ApplyAndPublish, nil
}
```
(`crud` stamps `id`, `tenant_id` and `updated_at`; `opening_cash` and `currency` are validated by the schema before the rule runs.)

## 3.3 `rule_cash_shift_update.go` — close a shift

Only `closing_cash_counted` may be sent; only the shift's own cashier closes it. The rule computes `status`, `closing_cash_expected`, `difference` and `closed_at` and **writes them itself** with `tx.UpdateFields` (they are not in `Fields`, so `crud` would not write them); `crud` then writes `closing_cash_counted` and `updated_at`.

```go
func (next *Shift) Update(a crud.CurrentUser, tx crud.Tx, old *Shift, fields []string, m *Module) (crud.RuleAnswer, error) {
	if err := onlyFields(fields, Shift_.ClosingCashCounted); err != nil {
		return 0, err
	}
	if old.CashierId != a.UserID {
		return 0, &model.FieldError{Field: Shift_.CashierId, Err: ErrNotYourShift}
	}
	if old.Status == ShiftClosed {
		return 0, ErrShiftAlreadyClosed
	}
	now := time.NowUnix()
	facts, err := factsBetween(tx, old.OpenedAt, now)
	if err != nil {
		return 0, err
	}
	expected := old.OpeningCash
	for _, f := range facts {
		if f.Currency != old.Currency || f.Means != string(money.MeansCash) || f.Actor != old.CashierId {
			continue // another currency, card or transfer, or another cashier: not this shift's cash
		}
		switch f.Kind {
		case string(money.Charged):
			expected += f.Amount
		case string(money.Refunded):
			expected -= f.Amount
		}
	}
	derived := Shift{
		Id:                  old.Id,
		Status:              ShiftClosed,
		ClosingCashExpected: expected,
		Difference:          next.ClosingCashCounted - expected,
		ClosedAt:            now,
	}
	err = tx.UpdateFields(&derived, old.Id, []string{Shift_.Status, Shift_.ClosingCashExpected, Shift_.Difference, Shift_.ClosedAt})
	if err != nil {
		return 0, err
	}
	return crud.ApplyAndPublish, nil
}
```
Imports: `webtyp.com/crud`, `webtyp.com/model`, `webtyp.com/money`, `webtyp.com/time`.

## 3.4 The three read models (signatures are the generated ones, with `filter`)

Each returns a pointer to the generated list type. They read through `m.store` (it is a `crud.Tx`).

**`rule_cash_today_read.go`** — the caller's open shift, zero or one row
```go
func (p *TodayShift) Read(a crud.CurrentUser, filter *TodayShift, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	list := TodayShiftList{}
	row, found, err := firstRow(m.store, func() model.Model { return &Shift{} },
		storage.Eq(Shift_.CashierId, a.UserID), storage.Eq(Shift_.Status, ShiftOpen))
	if err != nil {
		return nil, err
	}
	if !found {
		return &list, nil
	}
	shift := row.(*Shift)
	facts, err := factsBetween(m.store, shift.OpenedAt, time.NowUnix())
	if err != nil {
		return nil, err
	}
	list = append(list, &TodayShift{
		Id: shift.Id, CashierId: shift.CashierId, Status: shift.Status, OpeningCash: shift.OpeningCash,
		Currency: shift.Currency, OpenedAt: shift.OpenedAt, FactsCount: int64(len(facts)),
	})
	return &list, nil
}
```
**`rule_reorder_read.go`** — one row per item to buy again
```go
func (p *Reorder) Read(a crud.CurrentUser, filter *Reorder, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	list := ReorderList{}
	if m.deps.LowStock == nil {
		return &list, nil
	}
	items, err := m.deps.LowStock.LowStock(m.deps.TenantID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		list = append(list, &Reorder{ItemId: items[i].ItemId, ItemName: items[i].ItemName, Total: items[i].Total, MinQuantity: items[i].MinQuantity})
	}
	return &list, nil
}
```
**`rule_thirty_day_read.go`** — facts due in the range, then forecast lines; `from` and `to` are required
```go
func (p *ThirtyDay) Read(a crud.CurrentUser, filter *ThirtyDay, args crud.ReadArgs, m *Module) (model.ModelSlice, error) {
	if filter == nil || filter.From == 0 || filter.To == 0 {
		return nil, &model.FieldError{Field: ThirtyDay_.From, Err: ErrRangeRequired}
	}
	conditions := []storage.Condition{
		storage.Gte(StoredFact_.DueAt, int64(filter.From)),
		storage.Lte(StoredFact_.DueAt, int64(filter.To)),
	}
	if filter.Scope != "" {
		conditions = append(conditions, storage.Eq(StoredFact_.Scope, filter.Scope))
	}
	facts, err := factsWhere(m.store, conditions...)
	if err != nil {
		return nil, err
	}
	list := ThirtyDayList{}
	for i := range facts {
		list = append(list, &ThirtyDay{RowType: RowFact, Label: facts[i].Kind, Amount: facts[i].Amount, Currency: facts[i].Currency, DueAt: facts[i].DueAt})
	}
	if m.deps.Forecast != nil {
		lines, err := m.deps.Forecast.PendingAccruals(m.deps.TenantID, filter.From, filter.To)
		if err != nil {
			return nil, err
		}
		for i := range lines {
			list = append(list, &ThirtyDay{RowType: RowForecast, Label: lines[i].ThirdPartyName, Amount: lines[i].Amount, Currency: lines[i].Currency, DueAt: lines[i].DueAt})
		}
	}
	return &list, nil
}
```
A caller (and a test) passes the filter as `&crud.ReadArgs{Filter: &cashmanager.ThirtyDay{From: a, To: b}}`.

## 3.5 Done when

`go vet ./...` compiles the rule files (the module may still fail on `ops.go` until Stage 4). Commit, push, and **open the pull request now** (title: the `PLAN:` line).
