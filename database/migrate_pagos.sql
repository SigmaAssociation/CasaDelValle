-- Migracion: registrar pagos/abonos de una reserva.
-- Idempotente: puede ejecutarse varias veces sin efectos secundarios.
--
-- 1) Validaciones de la tabla pagos.
--    Diseno: tipo_pago distingue anticipo/pago; method guarda el medio de pago.
ALTER TABLE pagos DROP CONSTRAINT IF EXISTS chk_pago_cantidad;
ALTER TABLE pagos ADD CONSTRAINT chk_pago_cantidad CHECK (cantidad > 0);

ALTER TABLE pagos DROP CONSTRAINT IF EXISTS chk_pago_tipo;
ALTER TABLE pagos ADD CONSTRAINT chk_pago_tipo CHECK (tipo_pago IN ('anticipo', 'pago'));

ALTER TABLE pagos DROP CONSTRAINT IF EXISTS chk_pago_method;
ALTER TABLE pagos ADD CONSTRAINT chk_pago_method CHECK (method IN ('CARD_SIMULATED', 'TRANSFER_SIMULATED', 'CASH'));

-- 2) Nuevo tipo de notificacion "pago_registrado".
ALTER TABLE notificaciones DROP CONSTRAINT IF EXISTS chk_notificacion_tipo;
ALTER TABLE notificaciones ADD CONSTRAINT chk_notificacion_tipo CHECK (tipo IN (
    'bienvenida',
    'reserva_creada',
    'reserva_confirmada',
    'reserva_cancelada',
    'reserva_actualizada',
    'reserva_finalizada',
    'pago_registrado',
    'cabana_creada',
    'cabana_actualizada',
    'cabana_eliminada',
    'rol_actualizado'
));
