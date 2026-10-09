CREATE TABLE IF NOT EXISTS notificaciones (
    id SERIAL PRIMARY KEY,
    id_usuario INT NOT NULL,
    tipo VARCHAR(40) NOT NULL,
    mensaje TEXT NOT NULL,
    entidad_tipo VARCHAR(20),
    entidad_id INT,
    leida BOOLEAN NOT NULL DEFAULT FALSE,
    fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_notificacion_usuario
        FOREIGN KEY (id_usuario) REFERENCES usuarios(id) ON DELETE CASCADE,
    CONSTRAINT chk_notificacion_tipo
        CHECK (tipo IN (
            'bienvenida',
            'reserva_creada',
            'reserva_cancelada',
            'reserva_actualizada',
            'reserva_finalizada',
            'cabana_creada',
            'cabana_actualizada',
            'cabana_eliminada',
            'rol_actualizado'
        )),
    CONSTRAINT chk_notificacion_entidad_tipo
        CHECK (entidad_tipo IS NULL OR entidad_tipo IN ('reservacion', 'cabana', 'usuario'))
);

CREATE INDEX IF NOT EXISTS idx_notificaciones_usuario
    ON notificaciones (id_usuario, fecha_creacion DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_notificaciones_no_leidas
    ON notificaciones (id_usuario) WHERE leida = FALSE;
