# Cash Manager
<img src="docs/img/badges.svg">

Módulo de gestión de caja.

## ¿Por qué existe?

Este módulo permite a los cajeros abrir y cerrar turnos, controlando el dinero esperado en caja comparándolo con los [hechos de dinero](https://webtyp.com/money).
**Sin esto:** no se sabe cuánto efectivo debería haber al final del día.
**Con esto:** cada cajero tiene un [turno (cash_shift)](#) donde se acumulan las transacciones de efectivo para su conciliación final.

## Antes de comenzar

1. Requiere **Go** 1.26 o superior.
2. Configurar `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`.
3. Para instanciar el módulo `New`, se requiere un `*orm.DB` que contenga las tablas creadas por `migrate`, un inyector de dependencias para `Deps.IDs` y un `Deps.TenantID`.

## Guía paso a paso

1. Instanciar el módulo con la base de datos y un ID generator. Ejemplo:
```go
package main

import (
	"webtyp.com/fmt"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	"webtyp.com/unixid"
	cashmanager "github.com/veltylabs/cash_manager"
)

func main() {
    ids, _ := unixid.NewUnixID()
	mod := cashmanager.New(orm.New(mem.New()), cashmanager.Deps{IDs: ids, TenantID: "t1"})
	fmt.Println(mod.ModelName())
}
```

2. Abrir y cerrar turnos mediante el API CRUD. Ejemplo detallado en [tests/crud_test.go](tests/crud_test.go).
3. Para una interfaz completa, usar la demo ejecutable: [seed/demo.go](seed/demo.go).

## Mapa de documentación

| Orden | Documento | Qué pregunta responde y cuándo leerlo |
|---|---|---|
| 1 | [README.md](README.md) | ¿Qué es y cómo empezar a usarlo? Leer para setup rápido. |
| 2 | [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | ¿Cómo están organizados los flujos y datos? Leer para extender funcionalidades. |
| 3 | [docs/TESTS.md](docs/TESTS.md) | ¿Qué reglas de negocio están cubiertas y probadas? Leer antes de modificar lógica o entender el dominio. |

## Operaciones

| Operación | Definición |
|---|---|
| `OpenShift` | `cash_shift.create` with `{opening_cash, currency}` (el cajero es el invocador) |
| `CloseShift` | `cash_shift.update` with `Fields: ["closing_cash_counted"]` |
| `GetToday` | read model `cash_today` |
| `ListStoredFacts` | `stored_fact.read` |
| `Reorder` | read model `reorder` |
| `ThirtyDay` | read model `thirty_day` con filter `{from, to, scope}` |

## Permisos requeridos

| Recurso | Roles recomendados |
|---|---|
| `cash_shift` | `cashier` |
| `stored_fact` | `auditor`, `cashier` |
| `cash_today` | `cashier` |
| `reorder` | `cashier`, `manager` |
| `thirty_day` | `auditor`, `manager` |

