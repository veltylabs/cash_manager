← [Index](PLAN.md) | Next → [Stage 2](PLAN_STAGE_2_MODULE_STORE.md)

# Stage 1 — tools, `definition.go`, `errors.go`

## 1.1 Tools (run first, exactly)

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go install webtyp.com/definitionc/cmd/definitionc@v0.1.42
export PATH="$PATH:$(go env GOPATH)/bin"
go get webtyp.com/model@v0.2.19 webtyp.com/crud@v0.0.12 webtyp.com/orm@v0.12.33 webtyp.com/storage@v0.2.1 webtyp.com/ddl@v0.0.49 webtyp.com/router@v0.4.22 webtyp.com/definitionc@v0.1.42 webtyp.com/view@v0.6.74 webtyp.com/appkit@v0.2.35
go mod tidy
```

`GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`; remove every `replace` from `go.mod`. **`definitionc` rewrites `go.mod` to older versions every time it runs: after EVERY run repeat the `go get` line and `go mod tidy`.** If a module cannot be fetched, stop and say so in the PR; never invent a hash.

Facts you need (asked before; do not ask again):

- Arguments of an operation: `crud.CreateArgs{Records []model.Model}`, `crud.UpdateArgs{IDs []string, Fields []string, Record model.Model}`, `crud.DeleteArgs{IDs []string}`, `crud.ReadArgs{ID, Search, After string, Limit int64, Filter model.Model}`; never a bare entity pointer, a `*List` or a map. `model.Pointer`, `model.List` and `model.Map` do not exist: kinds are flat (`model.Text`, `Int`, `Float`, `Bool`, `UnixTime`, `NanoTime`, `Blob`, `Raw`).
- A rule's `tx` is a `crud.Tx`: `ReadOne(&rec, id)`, `ReadAll(newRow, crud.Query{Conditions, Order, After, Limit}, onRow)`, `Create`, `UpdateFields(&rec, id, cols)`, `Delete`. Conditions are `storage.Eq/Neq/Gt/Gte/Lt/Lte/In/Like` (AND only); the tenant is applied by the store: never add a `tenant_id` condition; give `Gte/Lte` an `int64(...)`.
- **Update rules:** `crud` writes `next` after the rule with exactly the columns of `Fields` (the ones the caller sent). A column the rule derives and the caller did NOT send (status, totals, a difference) must be written by the rule itself with `tx.UpdateFields(&row, id, []string{...})`; only assigning it to `next` leaves the row unchanged.

## 1.2 `definition.go` — exact edits

1. **Delete** `ResourceFeatureReorder`, `ResourceFeatureThirtyDay`, and the four argument types `OpenShiftArgs`, `CloseShiftArgs`, `GetTodayArgs`, `ListStoredFactsArgs` with all their methods (`Schema`, `Pointers`, `EncodeFields`, `DecodeFields`, `IsNil`). Keep `ShiftOpen`/`ShiftClosed`.
2. **`ShiftModel`** (`Name: "cash_shift"`): add `Allow: model.Create | model.Read | model.Update` and `Rules: model.Rules{Create: model.AddMyBusinessRule, Read: model.HaveNoBusinessRule, Update: model.AddMyBusinessRule}`; add `ServerOnly: true` to `tenant_id`, `cashier_id`, `status`, `closing_cash_expected`, `difference`, `opened_at`, `closed_at`, `updated_at`. `opening_cash`, `currency` and `closing_cash_counted` stay client-sendable.
3. **`StoredFactModel`** (`Name: "stored_fact"`): add `Allow: model.Read`, `Rules: model.Rules{Read: model.HaveNoBusinessRule}`, `ServerOnly: true` on `tenant_id`.
4. **Three new read models**, written exactly (flat, one row per result; `Wire: true`, no `DB:` marks). The Go type of each is its variable name without `Model` (`TodayShift`, `Reorder`, `ThirtyDay`):

```go
// TodayShiftModel is the caller's open shift (zero or one row).
var TodayShiftModel = model.Definition{
	Name:  "cash_today",
	Wire:  true,
	Allow: model.Read,
	Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "id", Type: model.Text()},
		{Name: "cashier_id", Type: model.Text()},
		{Name: "status", Type: model.Text()},
		{Name: "opening_cash", Type: model.Int()},
		{Name: "currency", Type: model.Text()},
		{Name: "opened_at", Type: model.UnixTime()},
		{Name: "facts_count", Type: model.Int()},
	},
}

// ReorderModel lists the items to buy again: one row per item (it replaces the old ReorderResponse).
var ReorderModel = model.Definition{
	Name:  "reorder",
	Wire:  true,
	Allow: model.Read,
	Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "item_id", Type: model.Text()},
		{Name: "item_name", Type: model.Text()},
		{Name: "total", Type: model.Float()},
		{Name: "min_quantity", Type: model.Float()},
	},
}

// ThirtyDayModel is the next-30-days view: one row per stored fact due in the range and one per forecast line.
// from, to and scope are the filter the caller sends; the other fields are the answer.
var ThirtyDayModel = model.Definition{
	Name:  "thirty_day",
	Wire:  true,
	Allow: model.Read,
	Rules: model.Rules{Read: model.AddMyBusinessRule},
	Fields: model.Fields{
		{Name: "from", Type: model.UnixTime(), OmitEmpty: true},
		{Name: "to", Type: model.UnixTime(), OmitEmpty: true},
		{Name: "scope", Type: model.Text(), OmitEmpty: true},
		{Name: "row_type", Type: model.Text(), OmitEmpty: true},
		{Name: "label", Type: model.Text(), OmitEmpty: true},
		{Name: "amount", Type: model.Int(), OmitEmpty: true},
		{Name: "currency", Type: model.Text(), OmitEmpty: true},
		{Name: "due_at", Type: model.UnixTime(), OmitEmpty: true},
	},
}
```
   `row_type` is the constant text `fact` or `forecast`: declare `const (RowFact = "fact"; RowForecast = "forecast")` next to `ShiftOpen`.
5. **`Definition`** (new), after the models:

```go
var Definition = model.ModuleDefinition{
	Name: "cash_manager",
	Entities: []model.Definition{ShiftModel, StoredFactModel, TodayShiftModel, ReorderModel, ThirtyDayModel},
}
```

## 1.3 `errors.go`

Keep the existing constants and add (exact text):

```go
	ErrNotYourShift     model.Invalid = "cash_manager: that shift belongs to another cashier, so it was not changed. Close only the shift you opened."
	ErrFieldNotEditable model.Invalid = "cash_manager: that field cannot be changed. Send only closing_cash_counted when closing a shift."
	ErrRangeRequired    model.Invalid = "cash_manager: from and to are required. Send both dates of the range."
```

`go vet ./...` still fails (`ops.go`, `module.go` reference removed types): expected until Stage 4. Commit and push.
