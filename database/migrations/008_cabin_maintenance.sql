-- Las fechas usan [inicio, fin), igual que las noches de una reserva.
CREATE TABLE mantenimientos (
    id SERIAL PRIMARY KEY,
    id_cabana INT NOT NULL REFERENCES cabanas(id) ON DELETE RESTRICT,
    fecha_inicio DATE NOT NULL,
    fecha_fin DATE NOT NULL,
    motivo VARCHAR(500) NOT NULL,
    fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    fecha_cancelacion TIMESTAMPTZ,
    CONSTRAINT chk_mantenimiento_fechas CHECK (fecha_fin > fecha_inicio)
);
CREATE INDEX idx_mantenimientos_cabana_fechas
    ON mantenimientos (id_cabana, fecha_inicio, fecha_fin) WHERE fecha_cancelacion IS NULL;

CREATE TABLE mantenimiento_reservaciones (
    id_mantenimiento INT NOT NULL REFERENCES mantenimientos(id),
    id_reservacion INT NOT NULL REFERENCES reservaciones(id),
    accion VARCHAR(20) NOT NULL CHECK (accion IN ('reasignar', 'reembolsar')),
    id_cabana_destino INT REFERENCES cabanas(id),
    monto_reembolsado NUMERIC(10,2) NOT NULL DEFAULT 0,
    PRIMARY KEY (id_mantenimiento, id_reservacion)
);

-- El mismo bloqueo transaccional coordina altas de mantenimientos y cambios
-- de reservas, incluso entre varias instancias de la API.
CREATE FUNCTION validar_reserva_mantenimiento() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    PERFORM pg_advisory_xact_lock(947, 1);
    IF NEW.estado IN ('pendiente', 'confirmada', 'en_curso') AND EXISTS (
        SELECT 1 FROM mantenimientos m
        WHERE m.id_cabana = NEW.id_cabana AND m.fecha_cancelacion IS NULL
          AND m.fecha_inicio < NEW.fecha_fin AND m.fecha_fin > NEW.fecha_inicio
    ) THEN
        RAISE EXCEPTION 'La cabaña está bloqueada por mantenimiento en esas fechas'
            USING ERRCODE = '23P01', CONSTRAINT = 'reserva_mantenimiento';
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER reserva_mantenimiento
    BEFORE INSERT OR UPDATE OF id_cabana, fecha_inicio, fecha_fin, estado ON reservaciones
    FOR EACH ROW EXECUTE FUNCTION validar_reserva_mantenimiento();
