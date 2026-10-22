---
PLAN: "fix: a shift counts only its cashier's cash; facts keep means and actor"
TAG: v0.2.0
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 17131905432708726339
---

> Read AGENTS.md at the repo root before any change.

<!-- Ready: webtyp.com/money v0.2.0 published and checked against "APIs used" on 2026-10-22. -->

# Plan — the shift counts cash, only its cashier's

## Why

Closing a shift compares the cash counted with the cash expected. Today the expected cash is the opening cash plus
**every** `Charged` and minus every `Refunded` fact of the tenant since the shift opened (`ops.go`, `closeShift`):
card payments and other cashiers' payments inflate it, so the difference is wrong on any normal day. Since
`webtyp.com/money` v0.2.0 a fact says **how** it was paid (`Means`) and **who** registered it (`Actor`), so the shift
counts only `Means == CASH` facts whose `Actor` is the shift's cashier.

Also: a card payment is now an `Accrued` fact whose source type is `payment_fee` for the fee; the stored fact id is
`Kind|SourceType|SourceID`, and today its `Permitted` rejects `_` — the fee fact would fail to store.

Language: identifiers, keys and error strings in English; **README, docs, diagrams and comments in Spanish**, each
exported symbol with a `// Ejemplo: …` line.

## Delivery rules

- **Commit after every stage.** Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash`.
- Never pause to ask permission or to report progress. When the Delivery check passes: push and open the PR. Stuck
  for 30 minutes: open the PR anyway, listing what fails.
- A PR that does not compile, or whose tests could not run, is a failed delivery.
- Never edit `model_orm.go` by hand: regenerate it with `ormc`.
- The code blocks below are the implementation: copy them.

## Tools and pinned versions

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/ormc/cmd/ormc@latest
export GOPRIVATE='webtyp.com/*,github.com/veltylabs/*' GOFLAGS=-mod=mod
go get webtyp.com/money@v0.2.0 webtyp.com/storage@v0.1.10 && go mod tidy   # money v0.2.0 publicado y verificado 2026-10-22
```

## APIs used (`webtyp.com/money` v0.2.0)

`money.Fact` has `Means money.Means` and `Actor string`; `money.MeansCash = "CASH"`. `Charged`/`Refunded` facts
always carry a valid `Means`.

## Stage 1 — Store means and actor (`model.go`, `module.go`)

`StoredFactModel`:

- `id`: `Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'|', '_', '-', '.'}}`.
- after `due_at` add:

```go
		{Name: "means", Type: model.Text(), OmitEmpty: true}, // CASH, CARD o TRANSFER
		{Name: "actor", Type: model.Text(), OmitEmpty: true}, // quién lo registró
```

Run `go mod tidy && ormc`. In `Accept` (`module.go`), the `StoredFact` literal gains `Means: string(f.Means), Actor:
f.Actor`. **Commit.**

## Stage 2 — The shift counts only its cashier's cash (`ops.go`)

In `closeShift`, replace the loop that computes `expected` with:

```go
	expected := shift.OpeningCash
	for _, f := range facts {
		if f.Currency != shift.Currency || f.Means != string(money.MeansCash) || f.Actor != shift.CashierId {
			continue // otra moneda, tarjeta o transferencia, u otra cajera: no es efectivo de este turno
		}
		switch f.Kind {
		case string(money.Charged):
			expected += f.Amount
		case string(money.Refunded):
			expected -= f.Amount
		}
	}
```

Tests (`tests/cash_manager_test.go`): `TestMoneySubscribeWiring`'s fact gets `Means: money.MeansCash`. New
`TestCloseShiftCountsOnlyItsCashiersCash`: open a shift for cashier `c1` with 1000 CLP; publish through the broker
(topic `money.Topic`) four facts timestamped after the opening — `Charged` 5000 CASH by `c1`, `Charged` 7000 CARD by
`c1`, `Charged` 3000 CASH by `c2`, `Refunded` 500 CASH by `c1`; close with counted 5500 → `ClosingCashExpected ==
5500`, `Difference == 0`. And `TestFeeFactIsStored`: `Accept` an `Accrued` Outflow fact with
`Source{Type: "payment_fee", ID: "p1"}`, `Means: MeansCard`, `DueAt` set → it is listed by `ThirtyDay` in its due window.
**Commit.**

## Stage 3 — Docs (Spanish)

`README.md`: one paragraph "Cierre de turno": se cuenta solo el efectivo (`Means` CASH) registrado por la cajera del
turno; las tarjetas aparecen en «Caja 30 días» el día del depósito, con su comisión como gasto. `docs/ARCHITECTURE.md`:
same rule with the example of Stage 2. **Commit.**

## Stage 4 — Spanish (only what is still in English)

Every `README.md`, `docs/*` file, diagram and code comment in Spanish; translate only what is still English; each
exported symbol's comment has a `// Ejemplo: …` line. **Commit.**

## Delivery check

```bash
go mod tidy && ormc && gotest
```

## Stages

| # | What | Files |
|---|---|---|
| 1 | Store means and actor | `model.go`, `model_orm.go` (generated), `module.go` |
| 2 | Shift counts its cashier's cash | `ops.go`, `tests/cash_manager_test.go` |
| 3 | Docs | `README.md`, `docs/ARCHITECTURE.md` |
| 4 | Spanish | every English `*.md` and `*.go` comment |
