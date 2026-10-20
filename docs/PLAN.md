---
PLAN: "feat: cash_manager — per-person cash shifts, the first money.Sink, and the Today/Close/Reorder/30-day screens"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 767920940306581773
PR: https://github.com/veltylabs/cash_manager/pull/1
REVIEWED_COMMIT: e21d75d5a1f0d1a7a16218172613886d3f2cf78d
---

> Read AGENTS.md at the repo root before any change (this is a new repo: AGENTS.md and go.mod already exist,
> `cash_manager.go`'s stub `CashManager`/`New()` is scaffolding from `gonew` — delete it, it is not the real
> module). This plan is complete: do not guess; if something does not fit, stop and ask in the PR. Design
> history: `webtyp/docs/MONEY_FLOWS_DESIGN.md` §4 (recommendation), §7 (P9: the shift replaces a fixed
> register), §6 (G4: reorder + 30-day cash view).

## Goal

The first consumer of `webtyp.com/money`'s `Sink` port (already published, v0.1.3). A cash **shift** belongs
to one **person**, not a physical register — P9 of `MONEY_FLOWS_MASTER_PLAN.md`: "no hay 2 cajas fijas; la
unidad es el turno de una persona". The module:

1. Lets a cashier **open** a shift (starting cash counted), **close** it with a cash count (`arqueo`) that
   reports the difference against what the facts said should be there, and never edits a closed shift.
2. Implements `money.Sink` and subscribes (`money.Subscribe`) to every `money.Fact` the application's broker
   carries — stores them, idempotent by `Fact.Key()`.
3. Draws four read screens: **Today** (totals since the cashier's shift opened), **Close** (the arqueo
   form), **Reorder** (items under their minimum, from `pharmacy_manager`/`supply_manager`), **30-day cash**
   (known inflows/outflows with a date, method-direct).
4. Gives a read-only "accountant" role access to every screen, no write access anywhere.

**Deliberately out of scope — do not build it:** any automatic purchase order (G4 of the design: manual
only), any cash-flow prediction beyond known-dated items (no forecasting model), a `shift_id` field written
back into `payment_manager`/`agreement_manager` (correlation is by `Fact.At` falling inside
`[Shift.OpenedAt, Shift.ClosedAt)` plus the cashier identity the application passes when it asks this module
to open a shift — never a cross-module foreign key).

## Git rules (strict)

- Work on the branch codejob created. Never `git reset`/`checkout -- <file>`/`clean`/`stash`. Commit often.
- **Never pause to ask permission or to report progress.** Delivery check passes → commit, push, open the PR.
  Still failing after 30 minutes → open the PR anyway, list exactly what fails.
- Do not edit `model_orm.go` by hand: it is generated.
- Module-boundary rule (`devskills/skills/module-layout/SKILL.md`): this module's root package never
  imports `payment_manager`, `agreement_manager`, `pharmacy_manager` or `supply_manager`. It declares narrow
  interfaces (below) that the application wires.

## Tools

```bash
go install webtyp.com/ormc/cmd/ormc@latest
go install webtyp.com/devflow/cmd/gotest@latest
```

## Stage 0 — dependencies

```bash
rm cash_manager.go
go get webtyp.com/money@v0.1.3 webtyp.com/model@v0.2.4 webtyp.com/orm@latest webtyp.com/events@latest webtyp.com/time@v0.6.2 webtyp.com/router@latest webtyp.com/view@latest webtyp.com/input@latest
go mod tidy
```

## Stage 1 — `model.go`

```go
package cashmanager

import "webtyp.com/model"

const (
	ShiftOpen   = "open"
	ShiftClosed = "closed"
)

var ShiftModel = model.Definition{
	Name: "cash_shift",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "cashier_id", Type: model.Text(), NotNull: true}, // soft reference — the application's own user id
		{Name: "status", Type: model.Text(), NotNull: true},
		{Name: "opening_cash", Type: model.Int(), NotNull: true},   // integer minor units, counted at open
		{Name: "closing_cash_counted", Type: model.Int()},          // counted at close; 0 while open
		{Name: "closing_cash_expected", Type: model.Int()},         // opening_cash + sum of Charged/Refunded cash facts in the window; computed at close
		{Name: "difference", Type: model.Int()},                    // closing_cash_counted - closing_cash_expected; can be negative
		{Name: "currency", Type: model.Text(), NotNull: true, Permitted: model.Permitted{Letters: true, Minimum: 3, Maximum: 3}},
		{Name: "opened_at", Type: model.UnixTime(), NotNull: true},
		{Name: "closed_at", Type: model.UnixTime(), OmitEmpty: true},
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true},
	},
}

// StoredFactModel is this module's own idempotent storage of every money.Fact it accepted, keyed by
// Fact.Key(). It is NOT an accounting ledger — just the raw facts, so Today/Close/30-day can query them.
var StoredFactModel = model.Definition{
	Name: "stored_fact",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}}, // Fact.Key()
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "kind", Type: model.Text(), NotNull: true},
		{Name: "side", Type: model.Int(), NotNull: true},
		{Name: "amount", Type: model.Int(), NotNull: true},
		{Name: "currency", Type: model.Text(), NotNull: true, Permitted: model.Permitted{Letters: true, Minimum: 3, Maximum: 3}},
		{Name: "scope", Type: model.Text(), NotNull: true},
		{Name: "source_type", Type: model.Text(), NotNull: true},
		{Name: "source_id", Type: model.Text(), NotNull: true},
		{Name: "party", Type: model.Text(), OmitEmpty: true},
		{Name: "at", Type: model.UnixTime(), NotNull: true},
		{Name: "due_at", Type: model.UnixTime(), OmitEmpty: true},
	},
}
```

Transport-only Args models for the ops in Stage 3: `OpenShiftArgs{TenantId, CashierId, OpeningCash,
Currency}`, `CloseShiftArgs{TenantId, Id, ClosingCashCounted}`, `GetTodayArgs{TenantId, CashierId}`,
`ListStoredFactsArgs{TenantId, Scope, From, To time.Unix}` (for the 30-day view).

## Stage 2 — regenerate

```bash
cp go.mod /tmp/x.mod.bak && ormc && cp /tmp/x.mod.bak go.mod && go mod tidy && gofmt -w model_orm.go
```

## Stage 3 — `module.go`, `interfaces.go`, `ops.go`: the Sink and the shift lifecycle

```go
// interfaces.go
package cashmanager

// LowStockReader answers which items are under their minimum right now, for the Reorder screen. Mirrors
// supply_manager's own ListLowStock/LowStock shape (check supply_manager/pharmacy_manager's current exact
// field names before wiring — do not assume without reading).
type LowStockReader interface {
	LowStock(tenantId string) ([]LowStockItem, error)
}

type LowStockItem struct {
	ItemId, ItemName string
	Total, MinQuantity float64
}

// AgreementForecastReader answers what a third party is known to owe/be owed, with a due date, for the
// 30-day view. The application wires this to agreement_manager once it exists; until then a nil reader
// means the 30-day screen shows only payment_manager's own stored facts (never a crash on a nil reader —
// treat it exactly like Deps.Publisher elsewhere in this ecosystem: optional, nil disables that source).
type AgreementForecastReader interface {
	PendingAccruals(tenantId string, from, to time.Unix) ([]ForecastLine, error)
}

type ForecastLine struct {
	ThirdPartyName string
	Amount         int64 // positive = expected in, negative = expected out
	Currency       string
	DueAt          time.Unix
}

type Deps struct {
	IDs        model.IDGenerator
	TenantID   string
	LowStock   LowStockReader           // optional
	Forecast   AgreementForecastReader  // optional
	Subscriber events.Subscriber        // optional — if set, the module subscribes itself to money.Fact in New
}
```

`*Module` implements `money.Sink`:

```go
// Accept implements money.Sink. Idempotent by f.Key(): a fact already stored is a no-op, not an error.
func (m *Module) Accept(f money.Fact) error {
	if err := f.Validate(); err != nil {
		return err
	}
	// ... store under id = f.Key() via a Create that treats "already exists" as success, not ErrIdTaken ...
}
```

If `Deps.Subscriber != nil`, `New` calls `money.Subscribe(deps.Subscriber, m, onError)` where `onError`
publishes (or logs, following whatever this module's own event-publication story ends up being — there is
none planned here beyond storing facts, so logging via `webtyp.com/fmt` is enough) — never silently drop a
rejected fact.

Shift lifecycle (`OpenShift`, `CloseShift`): `OpenShift` rejects a second open shift for the same
`cashier_id` (`ErrShiftAlreadyOpen`) — one open shift per cashier at a time, matching "cada turno debe hacer
su cierre" (K2 of the design, already accepted). `CloseShift` computes `closing_cash_expected` by summing
`StoredFact` rows where `Kind` is `Charged`/`Refunded`, `at` is inside `[shift.OpenedAt, now]`, and — since
`StoredFact` carries no cashier/shift reference by design (see Goal) — the application is responsible for
scoping cash-only facts to this cashier at the call site; this module assumes every `Charged`/`Refunded` row
whose `at` falls in the window belongs to the shift being closed. **This is a known simplification for a
single-cashier-per-scope setup; if the real deployment runs several simultaneous cashiers in the same scope,
say so in the PR instead of guessing a fix** — this plan does not resolve that case.

`GetToday`, `ListStoredFacts`, the Reorder op (`m.deps.LowStock.LowStock(tenantId)`, nil-safe), and the
30-day op (`m.deps.Forecast.PendingAccruals` merged with `ListStoredFacts`'s own facts that carry a
`DueAt`, nil-safe) follow the decode→validate→service→encode shape every other module in this ecosystem
uses (`ops.go` of `business_calendar`/`payment_manager` for the exact pattern).

A read-only "accountant" role: this module declares no special permission of its own — every read op is
`.Requires(resource, model.Read)` as usual; a write op additionally requires `model.Create`/`model.Update`.
The accountant-sees-everything-but-writes-nothing property is the application's RBAC configuration (grant
the accountant's role `Read` on this module's resources and nothing else), not code here.

## Stage 4 — `migrate/`, `ui/`, `seed/`, `web/`

Follow `business_calendar`'s exact file shape for each (read it first): `migrate/migrate.go` syncing
`ShiftModel`/`StoredFactModel`; `ui/browser.go` with three screens — "Hoy" (`GetToday`), "Cierre" (open/close
form), "Reposición y caja 30 días" (combines the Reorder and 30-day reads, since both are read-only
summaries with no independent navigation need); `seed/seed.go` with one open and one closed shift plus a
handful of stored facts; `web/client.go` demo wiring fake `LowStockReader`/`events.Subscriber` (see
`webtyp.com/events/mock`).

## Stage 5 — tests

Through the exported API, `webtyp.com/storage/mem`, `webtyp.com/events/mock`:

- `Accept` stores a fact once; calling it twice with the same `Fact` is a no-op, not an error (idempotency).
- `Accept` on an invalid `Fact` returns `Validate`'s error and stores nothing.
- `OpenShift` rejects a second concurrent open shift for the same cashier.
- `CloseShift` computes the right `closing_cash_expected`/`difference` from a scripted sequence of
  `Charged`/`Refunded` facts inside the shift window.
- `money.Subscribe` wiring: publishing a `Fact` on the broker reaches `Accept` (use `events/mock`'s Broker).
- `LowStock`/`Forecast` both nil → the Reorder/30-day ops still succeed, returning empty lists, never a
  panic or error.
- Tenant isolation: tenant A's shift/facts are invisible to tenant B.

## Stage 6 — docs

`README.md`, `docs/ARCHITECTURE.md`, `docs/diagrams/database.md` (Mermaid ERD of `cash_shift`/
`stored_fact`) — per `veltylabs/modules/AGENTS.md`'s "Documentation a module must carry" table.

## Delivery check

- `gotest` green (vet, race, tests, wasm).
- `git diff --stat main --diff-filter=D` shows only `cash_manager.go` (the `gonew` stub).
- `grep -rn "github.com/veltylabs/\(payment_manager\|agreement_manager\|pharmacy_manager\|supply_manager\)\""`
  outside `go.mod`/`go.sum` returns nothing (module-boundary rule).

## Stages

| # | File(s) | What |
|---|---|---|
| 0 | `go.mod` | dependencies; delete `cash_manager.go` stub |
| 1 | `model.go` | `ShiftModel`, `StoredFactModel`, transport Args |
| 2 | `model_orm.go` | regenerate |
| 3 | `interfaces.go` (new), `module.go`, `ops.go` | `money.Sink`, shift lifecycle, reads |
| 4 | `migrate/`, `ui/`, `seed/`, `web/` | schema, screens, demo data, runnable demo |
| 5 | `tests/*` | idempotency, shift lifecycle, Sink wiring, nil-safety, tenant isolation |
| 6 | `README.md`, `docs/ARCHITECTURE.md`, `docs/diagrams/database.md` | document the module |
