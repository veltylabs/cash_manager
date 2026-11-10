# Arquitectura

## 1. Visión general en un minuto

Cash Manager es un módulo basado en crud, utilizando el framework `webtyp.com/crud`. Declara modelos persistentes y modelos de lectura plana, usando domain rules inyectadas a través de `webtyp.com/definitionc`.

## 2. Piezas y estructura

- **cash_shift**: Tabla transaccional para turnos.
- **stored_fact**: Tabla de solo lectura de hechos en un ledger.
- **Modelos de lectura plana** (`cash_today`, `reorder`, `thirty_day`): Consultan de manera eficiente el estado sin unir demasiadas tablas.
- **Rules (`rule_*.go`)**: Implementación específica de validaciones, escritura y lógica de negocio.

## 3. Flujos paso a paso

- Al aceptar (`Accept`) un hecho transaccional de dinero, este se valida y se inserta silenciosamente en `stored_fact`. Es idempotente.
- Para **cerrar un turno**, la regla en `rule_cash_shift_update.go` lee todos los hechos relevantes y determina automáticamente el `closing_cash_expected` y la `difference`. Al escribirlos por su cuenta, asegura la integridad, por lo que el cliente solo envía la cantidad en caja contada (`closing_cash_counted`).
- Los modelos de lectura devuelven directamente un arreglo (`Slice`) generado, ahorrando conversiones extrañas para la capa API.

## 4. Contratos y límites de seguridad

El módulo garantiza que un cajero no puede cerrar turnos que no ha abierto, no pueden modificar hechos ya almacenados, y ciertos campos administrativos (`tenant_id`, `status`) solo son modificables desde el servidor, definidos en la estructura con `ServerOnly: true`.

