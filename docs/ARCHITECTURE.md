# cash_manager Architecture

Domain scope: Manages cash shifts per person, opening/closing with cash count, and acts as the first consumer of `money.Sink` for operational facts.

## Entities

- `cash_shift`: A person's active cash shift, recording opening and closing cash counts.
- `stored_fact`: Raw storage of accepted `money.Fact`s keyed by their idempotency key. Used to trace expectations for cash shift close difference.

## Patterns

- The module implements `webtyp.com/money`'s `Sink` to accept money facts directly.
- The shift lifecycle (`OpenShift`, `CloseShift`) computes differences when a shift is closed based on stored facts inside its time window.
- Read operations are strictly scoped to the `cashier_id` and the `tenant_id` respectively.

## Operations

| Name | Method | Action |
|---|---|---|
| OpenShift | `m.openShift` | `Create` |
| CloseShift | `m.closeShift` | `Update` |
| GetToday | `m.getToday` | `Read` |
| ListStoredFacts | `m.listStoredFacts` | `Read` |
| Reorder | `m.reorder` | `Read` |
| ThirtyDay | `m.thirtyDay` | `Read` |

## Cierre de turno y arqueo de caja

Al cerrar un turno mediante `CloseShift`:
1. Se recupera el turno por `TenantId` e `Id`.
2. Se consultan los `stored_fact` del tenant ocurridos entre `shift.OpenedAt` y la hora de cierre.
3. Se filtra estrictamente por:
   - `Currency == shift.Currency`
   - `Means == money.MeansCash`
   - `Actor == shift.CashierId`
4. Se calcula el efectivo esperado: `OpeningCash` sumando cobros (`money.Charged`) y restando devoluciones (`money.Refunded`).
5. La diferencia queda registrada en `shift.Difference = ClosingCashCounted - ClosingCashExpected`.

Cualquier cobro con tarjeta, transferencia o hecho registrado por otro cajero permanece persistido en `stored_fact` para reportería y conciliación global, pero no ingresa al conteo físico de la gaveta de caja de ese turno.

## Composition Root Example

```go
broker := events.NewBroker()
deps := cashmanager.Deps{
	IDs:        unixid.NewGenerator(),
	TenantID:   "tenant123",
	Subscriber: broker,
	// Optional dependencies for views: LowStock, Forecast
}
m := cashmanager.New(db, deps)
m.MountOperations(registry)
```
