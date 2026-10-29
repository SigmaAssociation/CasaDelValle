-- Agrega el tipo de notificación 'reserva_finalizada' al CHECK de la tabla
-- notificaciones. Las bases creadas antes de esta migración no lo tenían.
-- Es idempotente: se puede ejecutar varias veces sin efectos secundarios.

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
        'reserva_checkin',
        'pago_registrado',
        'cabana_creada',
        'cabana_actualizada',
        'cabana_eliminada',
        'rol_actualizado'
    ));
