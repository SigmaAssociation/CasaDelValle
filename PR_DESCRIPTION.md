# Nueva gestión de estados de reservaciones + pagos (BD)

## Resumen
Adapta BD y API al ciclo `proxima → en_curso → finalizada`, agrega check-in/check-out y validaciones de pagos.

## Cambios

**BD**
- `reservaciones.estado`: default `proxima`; `chk_estado` = `proxima|en_curso|finalizada|cancelada`.
- Nuevas columnas `fecha_checkin`, `fecha_checkout` + constraint e índice `(estado, fecha_inicio, fecha_fin)`.
- `pagos.method` (`CARD_SIMULATED|TRANSFER_SIMULATED|CASH`) + constraints `tipo_pago`, `cantidad > 0` + índice.
- `migrate_reservas_estados_pagos.sql`: migra `activa`, crea función y job **pg_cron** (cada 15 min).
- Docker: `Dockerfile.postgres` con pg_cron y montaje de la migración.

**API**
- `model.go`: `StatusActive` → `StatusUpcoming`/`StatusInProgress` + helper `IsActiveStatus`.
- `repository.go`/`service.go`: cancelar, overlap y editar usan `IN ('proxima','en_curso')`; `FinalizePastReservations` sella checkin/checkout.
- Tests actualizados.

## ⚠️ Antes de desplegar
- Aplicar `database/migrate_reservas_estados_pagos.sql` **manualmente** en BD existentes (el `docker-entrypoint-initdb.d` solo corre con volumen vacío).
- Breaking change: `GET /reservations/*` ahora devuelve `status = proxima|en_curso` (ya no `activa`).

## Pendiente Frontend
- `my-reservations-page.ts:30` — filtrar `upcoming` por `proxima`/`en_curso`.
- `reservation-card.html:19,52,61` — reemplazar `'activa'` por `proxima` y añadir caso `en_curso`.
- `models/reservation.ts:11` y `reservation-detail.ts:8` — tipar `status`.
- Specs: `reservation-card.spec.ts:14`, `cancel-reservation-modal.spec.ts:11`, `notification-navigation.service.spec.ts:58`.
- Sin cambios necesarios: `reservation-card.ts:53` (usa `cancelada`).

## Notas
- pg_cron finaliza cada 15 min; el job Go (6 h) podría quedar sin filas → notificaciones de "finalizada" podrían no enviarse.
- Pagos: columna/constraints en BD, pero **sin implementar** en API ni FE.
