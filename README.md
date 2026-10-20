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

## Key Files

- `module.go`, `ops.go`: Module operations and `money.Sink` implementation.
- `model.go`, `model_orm.go`: Models and database persistence.
