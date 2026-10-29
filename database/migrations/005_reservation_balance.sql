-- Cargos adicionales y reembolsos para calcular el saldo de una reservación.

CREATE TABLE IF NOT EXISTS cargos_adicionales (
    id SERIAL PRIMARY KEY,
    id_reservacion INT NOT NULL,
    concepto VARCHAR(150) NOT NULL,
    monto NUMERIC(10, 2) NOT NULL CHECK (monto > 0),
    id_usuario_registro INT,
    fecha TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_cargo_reservacion
        FOREIGN KEY (id_reservacion) REFERENCES reservaciones(id) ON DELETE RESTRICT,
    CONSTRAINT fk_cargo_usuario_registro
        FOREIGN KEY (id_usuario_registro) REFERENCES usuarios(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS reembolsos (
    id SERIAL PRIMARY KEY,
    id_reservacion INT NOT NULL,
    id_pago INT,
    cantidad NUMERIC(10, 2) NOT NULL CHECK (cantidad > 0),
    motivo TEXT,
    id_usuario_registro INT,
    fecha TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_reembolso_reservacion
        FOREIGN KEY (id_reservacion) REFERENCES reservaciones(id) ON DELETE RESTRICT,
    CONSTRAINT fk_reembolso_pago
        FOREIGN KEY (id_pago) REFERENCES pagos(id) ON DELETE RESTRICT,
    CONSTRAINT fk_reembolso_usuario_registro
        FOREIGN KEY (id_usuario_registro) REFERENCES usuarios(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_cargos_reservacion ON cargos_adicionales (id_reservacion, fecha, id);
CREATE INDEX IF NOT EXISTS idx_reembolsos_reservacion ON reembolsos (id_reservacion, fecha, id);
CREATE INDEX IF NOT EXISTS idx_pagos_reservacion ON pagos (id_reservacion, fecha, id);
