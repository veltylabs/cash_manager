# cash_manager
<img src="docs/img/badges.svg">

Cash shifts per person, opening/closing with cash count, and money.Sink for operational facts.

## Operations

| Name | Input | Permission | Output |
|---|---|---|---|
| `OpenShift` | `OpenShiftArgs` | `Create` | `Shift` |
| `CloseShift` | `CloseShiftArgs` | `Update` | `Shift` |
| `GetToday` | `GetTodayArgs` | `Read` | `TodayResponse` |
| `ListStoredFacts` | `ListStoredFactsArgs` | `Read` | `StoredFactList` |
| `Reorder` | - | `Read` | `ReorderResponse` |
| `ThirtyDay` | `ListStoredFactsArgs` | `Read` | `ThirtyDayResponse` |

## Cierre de turno

Un turno (`Shift`) calcula su efectivo esperado (`ClosingCashExpected`) considerando exclusivamente los hechos financieros (`money.Fact`):
- En la misma moneda (`Currency == shift.Currency`).
- Con medio de pago en efectivo (`Means == money.MeansCash`).
- Operados por la misma cajera (`Actor == shift.CashierId`).
- Ocurridos durante la ventana del turno (`At >= shift.OpenedAt` y `At <= shift.ClosedAt`).

Los pagos con tarjeta (`MeansCard`), transferencias (`MeansTransfer`) u operados por otro personal no alteran el arqueo de efectivo del turno en caja.

## Key Files

- `module.go`, `ops.go`: Module operations and `money.Sink` implementation.
- `model.go`, `model_orm.go`: Models and database persistence.
