| Test | Rule |
|---|---|
| [TestShift_OneOpenPerCashier](../tests/crud_test.go) | Un cajero solo puede tener un turno abierto a la vez |
| [TestShift_OnlyOwnCashierCloses](../tests/crud_test.go) | Solo el cajero que abrió el turno puede cerrarlo |
| [TestShift_CloseComputesDifference](../tests/crud_test.go) | Al cerrar el turno se calcula automáticamente la diferencia de efectivo |
| [TestShift_CloseCountsOnlyItsCashiersCash](../tests/crud_test.go) | Al cerrar el turno solo se cuenta el efectivo de ese cajero en particular |
| [TestShift_CloseOnlyCountedField](../tests/crud_test.go) | Solo se permite enviar el campo closing_cash_counted al cerrar un turno |
| [TestReadModel_TodayIsCallers](../tests/crud_test.go) | El modelo de lectura cash_today solo muestra el turno del cajero solicitante |
| [TestReadModel_ThirtyDayRange](../tests/crud_test.go) | La lectura de 30 días requiere y respeta el rango de fechas |
| [TestReadModel_ThirtyDayNeedsRange](../tests/crud_test.go) | El rango de fechas es obligatorio para la lectura de 30 días |
| [TestReadModel_ReorderRows](../tests/crud_test.go) | La lista de reorden usa correctamente la dependencia LowStock |
| [TestFact_ReadOnly](../tests/crud_test.go) | Los hechos almacenados son de solo lectura |
| [TestMount_MatchesDefinition](../tests/crud_test.go) | El módulo se puede montar exitosamente según su definición |
| [TestCreate_PublishesChange](../tests/crud_test.go) | Publicación de eventos de cambio al crear entidades |
| [TestDemo_Valid](../tests/demo_test.go) | La demostración inicia correctamente |
| [TestAccept_SameFactTwiceIsStoredOnce](../tests/accept_test.go) | un hecho que llega dos veces se guarda una sola vez (sin error la segunda) |
| [TestAccept_InvalidFactIsRefused](../tests/accept_test.go) | un hecho vacío se rechaza y no deja filas |
| [TestAccept_FeeFactAppearsInTheThirtyDayView](../tests/accept_test.go) | la comisión (un devengo con vencimiento) sale como una fila `fact` de 447 en la vista a 30 días |
| [TestAccept_ArrivesThroughTheMoneyTopic](../tests/accept_test.go) | un hecho publicado en el tópico de dinero lo guarda el módulo suscrito |
