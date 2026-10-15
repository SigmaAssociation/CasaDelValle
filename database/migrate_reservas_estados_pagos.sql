BEGIN;


ALTER TABLE reservaciones
    DROP CONSTRAINT IF EXISTS chk_estado;

-- Fecha de hoy en Guatemala
UPDATE reservaciones
SET estado = CASE
    WHEN fecha_fin    <  (now() AT TIME ZONE 'America/Guatemala')::date THEN 'finalizada'
    WHEN fecha_inicio <= (now() AT TIME ZONE 'America/Guatemala')::date THEN 'en_curso'
    ELSE 'proxima'
END
WHERE estado = 'activa';

ALTER TABLE reservaciones
    ADD CONSTRAINT chk_estado
    CHECK (estado IN ('proxima', 'en_curso', 'finalizada', 'cancelada'));

ALTER TABLE reservaciones
    ALTER COLUMN estado SET DEFAULT 'proxima';


ALTER TABLE reservaciones
    ADD COLUMN IF NOT EXISTS fecha_checkin  TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS fecha_checkout TIMESTAMPTZ;

ALTER TABLE reservaciones
    DROP CONSTRAINT IF EXISTS chk_checkin_checkout;

ALTER TABLE reservaciones
    ADD CONSTRAINT chk_checkin_checkout
    CHECK (fecha_checkout IS NULL
           OR (fecha_checkin IS NOT NULL AND fecha_checkout >= fecha_checkin));

CREATE INDEX IF NOT EXISTS idx_reservaciones_estado_fechas
    ON reservaciones (estado, fecha_inicio, fecha_fin);


ALTER TABLE pagos
    ADD COLUMN IF NOT EXISTS method VARCHAR(20) NOT NULL DEFAULT 'CASH';

ALTER TABLE pagos
    ALTER COLUMN method DROP DEFAULT;

UPDATE pagos SET method = 'CASH' WHERE method IS NULL;
    ALTER TABLE pagos ALTER COLUMN method SET NOT NULL;

ALTER TABLE pagos
    DROP CONSTRAINT IF EXISTS chk_pago_method;

ALTER TABLE pagos
    ADD CONSTRAINT chk_pago_method
    CHECK (method IN ('CARD_SIMULATED', 'TRANSFER_SIMULATED', 'CASH'));

-- NOT VALID: se aplica a pagos nuevos sin rechazar filas históricas.
ALTER TABLE pagos
    DROP CONSTRAINT IF EXISTS chk_pago_tipo;

ALTER TABLE pagos
    ADD CONSTRAINT chk_pago_tipo
    CHECK (tipo_pago IN ('anticipo', 'pago')) NOT VALID;

ALTER TABLE pagos
    DROP CONSTRAINT IF EXISTS chk_pago_cantidad;

ALTER TABLE pagos
    ADD CONSTRAINT chk_pago_cantidad
    CHECK (cantidad > 0) NOT VALID;

CREATE INDEX IF NOT EXISTS idx_pagos_reservacion
    ON pagos (id_reservacion);


CREATE OR REPLACE FUNCTION actualizar_estados_reservaciones() RETURNS void
LANGUAGE plpgsql AS $$
DECLARE
    tz   CONSTANT text := 'America/Guatemala';
    hoy  date := (now() AT TIME ZONE tz)::date;
BEGIN
    -- Vencidas: proxima/en_curso cuya fecha_fin ya pasó -> finalizada
    UPDATE reservaciones
    SET estado         = 'finalizada',
        fecha_checkin  = COALESCE(fecha_checkin,  fecha_inicio::timestamp AT TIME ZONE tz),
        fecha_checkout = COALESCE(fecha_checkout, fecha_fin::timestamp    AT TIME ZONE tz)
    WHERE estado IN ('proxima', 'en_curso') AND fecha_fin < hoy;

    -- Iniciadas: proxima cuya fecha_inicio llegó -> en_curso
    UPDATE reservaciones
    SET estado        = 'en_curso',
        fecha_checkin = COALESCE(fecha_checkin, now())
    WHERE estado = 'proxima' AND fecha_inicio <= hoy;
END $$;

-- Job de pg_cron: actualiza estados cada 15 minutos.
-- Se omite (sin error) si pg_cron no está disponible o si esta base no es
-- la configurada en cron.database_name, para no romper entornos antiguos.
DO $$
BEGIN
    IF current_setting('cron.database_name', true) IS DISTINCT FROM current_database() THEN
        RAISE NOTICE 'pg_cron no está configurado para la base %; se omite el job.', current_database();
    ELSIF NOT EXISTS (SELECT 1 FROM pg_available_extensions WHERE name = 'pg_cron') THEN
        RAISE NOTICE 'La extensión pg_cron no está disponible; se omite el job.';
    ELSE
        CREATE EXTENSION IF NOT EXISTS pg_cron;
        PERFORM cron.schedule(
            'actualizar-estados-reservaciones',
            '*/15 * * * *',
            'SELECT actualizar_estados_reservaciones()'
        );
    END IF;
END $$;

COMMIT;
