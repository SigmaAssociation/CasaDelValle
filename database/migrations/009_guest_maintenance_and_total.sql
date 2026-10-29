-- Una cotización y una reserva usan exactamente las mismas reglas por noche.
-- La configuración particular de la cabaña tiene prioridad sobre la global,
-- incluso si su porcentaje es cero. Los intervalos de temporadas son inclusivos.
CREATE FUNCTION cotizar_noches(p_cabana INT, p_inicio DATE, p_fin DATE)
RETURNS TABLE(dia DATE, base NUMERIC, precio NUMERIC, fuente TEXT, temporada_nombre TEXT)
LANGUAGE sql STABLE AS $$
    SELECT d::date, c.precio,
        ROUND(CASE WHEN cp.id IS NOT NULL THEN
            COALESCE(cp.precio_fijo, c.precio * (1 + cp.porcentaje_ajuste / 100))
            WHEN g.id IS NOT NULL THEN c.precio * (1 + g.porcentaje_ajuste_sugerido / 100)
            ELSE c.precio END, 2),
        CASE WHEN cp.precio_fijo IS NOT NULL THEN 'cabin_fixed'
             WHEN cp.id IS NOT NULL THEN 'cabin_percent'
             WHEN g.id IS NOT NULL THEN 'global_season' ELSE 'normal' END,
        CASE WHEN cp.id IS NOT NULL THEN cs.nombre ELSE g.nombre END
    FROM cabanas c
    CROSS JOIN generate_series(p_inicio::timestamp, (p_fin - 1)::timestamp, interval '1 day') d
    LEFT JOIN LATERAL (
        SELECT * FROM precios_temporada p WHERE p.id_cabana=c.id
        AND d::date BETWEEN p.fecha_inicio AND p.fecha_fin ORDER BY p.id LIMIT 1
    ) cp ON TRUE
    LEFT JOIN temporada cs ON cs.id=cp.id_temporada
    LEFT JOIN LATERAL (
        SELECT * FROM temporada t WHERE t.activo AND d::date BETWEEN t.fecha_inicio AND t.fecha_fin
        ORDER BY t.id LIMIT 1
    ) g ON TRUE
    WHERE c.id=p_cabana
    ORDER BY d;
$$;
CREATE FUNCTION calcular_precio_total(p_cabana INT, p_inicio DATE, p_fin DATE)
RETURNS NUMERIC LANGUAGE sql STABLE AS $$
    SELECT COALESCE(SUM(precio),0) FROM cotizar_noches(p_cabana,p_inicio,p_fin);
$$;

ALTER TABLE reservaciones ADD COLUMN precio_total NUMERIC(10,2);
-- No se pueden reconstruir tarifas históricas eliminadas. Se prefiere el
-- importe acordado existente y, si falta, se cotiza con las reglas actuales.
UPDATE reservaciones SET precio_total=COALESCE(precio_total_acordado,
    CASE WHEN precio_noche_acordado IS NOT NULL THEN precio_noche_acordado*(fecha_fin-fecha_inicio)
    ELSE calcular_precio_total(id_cabana,fecha_inicio,fecha_fin) END);
ALTER TABLE reservaciones ALTER COLUMN precio_total SET NOT NULL;
ALTER TABLE reservaciones ADD CONSTRAINT chk_precio_total CHECK (precio_total>=0);

CREATE FUNCTION guardar_precio_reserva() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP='INSERT' THEN
        NEW.precio_total:=calcular_precio_total(NEW.id_cabana,NEW.fecha_inicio,NEW.fecha_fin);
    ELSIF NEW.id_cabana IS DISTINCT FROM OLD.id_cabana
       OR NEW.fecha_inicio IS DISTINCT FROM OLD.fecha_inicio OR NEW.fecha_fin IS DISTINCT FROM OLD.fecha_fin THEN
        IF OLD.estado='por_mantenimiento' AND NEW.id_cabana=OLD.id_cabana THEN
            NEW.precio_total:=OLD.precio_total;
        ELSE
            NEW.precio_total:=calcular_precio_total(NEW.id_cabana,NEW.fecha_inicio,NEW.fecha_fin);
        END IF;
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER guardar_precio_reserva BEFORE INSERT OR UPDATE OF id_cabana,fecha_inicio,fecha_fin ON reservaciones
    FOR EACH ROW EXECUTE FUNCTION guardar_precio_reserva();

ALTER TABLE reservaciones DROP CONSTRAINT chk_estado;
ALTER TABLE reservaciones ADD CONSTRAINT chk_estado CHECK
    (estado IN ('pendiente','confirmada','en_curso','finalizada','cancelada','por_mantenimiento'));
ALTER TABLE mantenimiento_reservaciones DROP CONSTRAINT mantenimiento_reservaciones_accion_check;
ALTER TABLE mantenimiento_reservaciones ADD CONSTRAINT mantenimiento_reservaciones_accion_check CHECK
    (accion IN ('pendiente','reprogramar','reembolsar','sin_efecto','reasignar'));
ALTER TABLE mantenimiento_reservaciones
    ADD COLUMN estado_anterior VARCHAR(20),
    ADD COLUMN fecha_resolucion TIMESTAMPTZ;
CREATE UNIQUE INDEX idx_mantenimiento_decision_pendiente ON mantenimiento_reservaciones(id_reservacion) WHERE accion='pendiente';

-- Una reserva afectada conserva su espacio mientras el huésped decide.
-- Los estados suspendidos pueden coexistir con el bloqueo de mantenimiento.
CREATE OR REPLACE FUNCTION validar_reserva_mantenimiento() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    PERFORM pg_advisory_xact_lock(947,1);
    IF NEW.estado IN ('pendiente','confirmada','en_curso') AND EXISTS (
        SELECT 1 FROM mantenimientos m WHERE m.id_cabana=NEW.id_cabana AND m.fecha_cancelacion IS NULL
        AND m.fecha_inicio<NEW.fecha_fin AND m.fecha_fin>NEW.fecha_inicio
    ) THEN
        RAISE EXCEPTION 'La cabaña está bloqueada por mantenimiento en esas fechas'
            USING ERRCODE='23P01',CONSTRAINT='reserva_mantenimiento';
    END IF;
    IF NEW.estado IN ('pendiente','confirmada','en_curso')
    AND NOT (TG_OP='UPDATE' AND OLD.estado='por_mantenimiento') AND EXISTS (
        SELECT 1 FROM reservaciones r WHERE r.id_cabana=NEW.id_cabana AND r.id<>NEW.id
        AND r.estado='por_mantenimiento' AND r.fecha_inicio<NEW.fecha_fin AND r.fecha_fin>NEW.fecha_inicio
    ) THEN
        RAISE EXCEPTION 'La cabaña tiene una reserva afectada por mantenimiento en esas fechas'
            USING ERRCODE='23P01',CONSTRAINT='reserva_mantenimiento';
    END IF;
    RETURN NEW;
END $$;

ALTER TABLE notificaciones DROP CONSTRAINT chk_notificacion_tipo;
ALTER TABLE notificaciones ADD CONSTRAINT chk_notificacion_tipo CHECK(tipo IN (
    'bienvenida','reserva_creada','reserva_confirmada','reserva_cancelada','reserva_actualizada',
    'reserva_finalizada','reserva_checkin','pago_registrado','cabana_creada','cabana_actualizada',
    'cabana_eliminada','rol_actualizado','reserva_mantenimiento'
));
