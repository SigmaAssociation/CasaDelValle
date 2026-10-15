-- Nuevo modelo de estados de reservaciones:
--   pendiente -> confirmada (pago del anticipo) -> en_curso (inicio de la
--   estancia) -> finalizada (fin de la estancia), más cancelada como estado
--   terminal.
--
-- Las reservas 'activa' existentes pasan a 'confirmada' (ya estaban vigentes).
-- También agrega el tipo de notificación 'reserva_confirmada' al CHECK de la
-- tabla notificaciones. Es idempotente.
--
-- Uso:
--   docker exec -i postgres_db psql -U devuser -d appdb \
--     < database/migrate_reservation_states.sql

-- 1) Reservaciones: estados nuevos (la data existente se migra antes de volver
--    a imponer el CHECK, para no violarlo).
ALTER TABLE reservaciones DROP CONSTRAINT IF EXISTS chk_estado;

UPDATE reservaciones SET estado = 'confirmada' WHERE estado = 'activa';

ALTER TABLE reservaciones
    ADD CONSTRAINT chk_estado
    CHECK (estado IN ('pendiente', 'confirmada', 'en_curso', 'finalizada', 'cancelada'));

ALTER TABLE reservaciones ALTER COLUMN estado SET DEFAULT 'pendiente';

-- 2) Notificaciones: tipo 'reserva_confirmada' para el pago del anticipo.
ALTER TABLE notificaciones
    DROP CONSTRAINT IF EXISTS chk_notificacion_tipo;

ALTER TABLE notificaciones
    ADD CONSTRAINT chk_notificacion_tipo
    CHECK (tipo IN (
        'bienvenida',
        'reserva_creada',
        'reserva_confirmada',
        'reserva_cancelada',
        'reserva_actualizada',
        'reserva_finalizada',
        'cabana_creada',
        'cabana_actualizada',
        'cabana_eliminada',
        'rol_actualizado'
    ));
