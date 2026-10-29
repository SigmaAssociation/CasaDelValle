CREATE TABLE IF NOT EXISTS temporada (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL,
    porcentaje_ajuste_sugerido NUMERIC(6, 2) NOT NULL DEFAULT 0,
    activo BOOLEAN NOT NULL DEFAULT TRUE,
    fecha_inicio DATE NOT NULL,
    fecha_fin DATE NOT NULL,
    CONSTRAINT chk_temporada_fechas CHECK (fecha_fin >= fecha_inicio)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_temporada_sin_cruce
    ON temporada (fecha_inicio, fecha_fin)
    WHERE activo = TRUE;

CREATE TABLE IF NOT EXISTS precios_temporada (
    id SERIAL PRIMARY KEY,
    id_cabana INT NOT NULL,
    id_temporada INT,
    fecha_inicio DATE NOT NULL,
    fecha_fin DATE NOT NULL,
    porcentaje_ajuste NUMERIC(6, 2) NOT NULL DEFAULT 0,
    precio_fijo NUMERIC(10, 2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_precios_temporada_cabana FOREIGN KEY (id_cabana) REFERENCES cabanas(id) ON DELETE CASCADE,
    CONSTRAINT fk_precios_temporada_temporada FOREIGN KEY (id_temporada) REFERENCES temporada(id) ON DELETE SET NULL,
    CONSTRAINT chk_precios_temporada_fechas CHECK (fecha_fin >= fecha_inicio)
);

CREATE INDEX IF NOT EXISTS idx_precios_temporada_cabana_fechas
    ON precios_temporada (id_cabana, fecha_inicio, fecha_fin);

ALTER TABLE reservaciones
    ADD COLUMN IF NOT EXISTS precio_noche_acordado NUMERIC(10, 2),
    ADD COLUMN IF NOT EXISTS precio_total_acordado NUMERIC(10, 2);

INSERT INTO temporada (nombre, porcentaje_ajuste_sugerido, activo, fecha_inicio, fecha_fin)
VALUES
    ('Navidad', 25.00, TRUE, '2026-12-20', '2026-12-31'),
    ('Semana Santa', 20.00, TRUE, '2026-04-01', '2026-04-12'),
    ('Temporada Alta', 15.00, TRUE, '2026-07-01', '2026-08-31')
ON CONFLICT DO NOTHING;
