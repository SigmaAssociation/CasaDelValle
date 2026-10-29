CREATE EXTENSION IF NOT EXISTS btree_gist;

DROP INDEX IF EXISTS idx_temporada_sin_cruce;

ALTER TABLE temporada
    ADD CONSTRAINT excl_temporada_fechas_activas
    EXCLUDE USING gist (daterange(fecha_inicio, fecha_fin, '[]') WITH &&)
    WHERE (activo);

ALTER TABLE precios_temporada
    ADD CONSTRAINT excl_precios_temporada_cabana_fechas
    EXCLUDE USING gist (id_cabana WITH =, daterange(fecha_inicio, fecha_fin, '[]') WITH &&);

ALTER TABLE precios_temporada
    ADD CONSTRAINT chk_precios_temporada_precio_fijo
    CHECK (precio_fijo IS NULL OR precio_fijo > 0);
