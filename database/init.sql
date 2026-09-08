CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    tipo VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS usuarios (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(150) NOT NULL,
    telefono VARCHAR(20),
    direccion VARCHAR(255),
    dpi CHAR(13) NOT NULL UNIQUE,
    correo VARCHAR(150) NOT NULL UNIQUE,
    contrasena VARCHAR(255) NOT NULL,
    
    id_rol INT NOT NULL,
    CONSTRAINT fk_usuario_rol FOREIGN KEY (id_rol) REFERENCES roles(id) ON DELETE RESTRICT
);

-- Añadí Cabañas 
CREATE TABLE IF NOT EXISTS cabanas (
    id SERIAL PRIMARY KEY,
    direccion VARCHAR(255) NOT NULL,
    precio NUMERIC(10, 2) NOT NULL,
    descripcion TEXT,
    capacidad INT NOT NULL,
    reglas TEXT,
    id_anfitrion INT NOT NULL,
    id_comision INT, 
    
    CONSTRAINT fk_cabana_anfitrion 
        FOREIGN KEY (id_anfitrion) REFERENCES usuarios(id) ON DELETE RESTRICT
);

-- Añadí Comisiones 
CREATE TABLE IF NOT EXISTS comisiones (
    id SERIAL PRIMARY KEY,
    id_cabana INT NOT NULL,
    id_usuario INT NOT NULL,
    porcentaje_comision NUMERIC(5, 2) NOT NULL,
    fecha_de_cambio TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_comision_cabana 
        FOREIGN KEY (id_cabana) REFERENCES cabanas(id) ON DELETE CASCADE,
        
    CONSTRAINT fk_comision_usuario 
        FOREIGN KEY (id_usuario) REFERENCES usuarios(id) ON DELETE RESTRICT    
);

-- Relación Cabañas con Comisiones 
ALTER TABLE cabanas
ADD CONSTRAINT fk_cabana_comision 
    FOREIGN KEY (id_comision) REFERENCES comisiones(id) ON DELETE SET NULL;
    
-- Valoraciones cabañas y usuarios     
CREATE TABLE IF NOT EXISTS valoracion_cabana (
    id SERIAL PRIMARY KEY,
    puntuacion INT NOT NULL CHECK (puntuacion >= 1 AND puntuacion <= 5),
    id_usuario_emisor INT NOT NULL,
    id_cabana INT NOT NULL,
    
    CONSTRAINT fk_vc_usuario_emisor 
        FOREIGN KEY (id_usuario_emisor) REFERENCES usuarios(id) ON DELETE CASCADE,
        
    CONSTRAINT fk_vc_cabana 
        FOREIGN KEY (id_cabana) REFERENCES cabanas(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS valoracion_usuario (
    id SERIAL PRIMARY KEY,
    puntuacion INT NOT NULL CHECK (puntuacion >= 1 AND puntuacion <= 5),
    id_usuario_emisor INT NOT NULL,
    id_usuario_valorado INT NOT NULL,
    
    CONSTRAINT fk_vu_usuario_emisor 
        FOREIGN KEY (id_usuario_emisor) REFERENCES usuarios(id) ON DELETE CASCADE,
        
    CONSTRAINT fk_vu_usuario_valorado 
        FOREIGN KEY (id_usuario_valorado) REFERENCES usuarios(id) ON DELETE CASCADE
);

-- Añadi tabla imagenes
CREATE TABLE IF NOT EXISTS imagenes (
    id SERIAL PRIMARY KEY,
    ruta VARCHAR(255) NOT NULL,
    id_usuario INT,
    id_cabana INT,
    
    CONSTRAINT fk_imagen_usuario 
        FOREIGN KEY (id_usuario) REFERENCES usuarios(id) ON DELETE CASCADE,
        
    CONSTRAINT fk_imagen_cabana 
        FOREIGN KEY (id_cabana) REFERENCES cabanas(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS reservaciones (
    id SERIAL PRIMARY KEY,
    fecha_inicio DATE NOT NULL,
    fecha_fin DATE NOT NULL,
    id_cabana INT NOT NULL, 
    id_usuario INT NOT NULL,
    disponible BOOLEAN DEFAULT TRUE,

    CONSTRAINT fk_reservacion_usuario FOREIGN KEY (id_usuario) REFERENCES usuarios(id) ON DELETE RESTRICT,
    
-- Añadí esta fk
CONSTRAINT fk_reservacion_cabana
        FOREIGN KEY (id_cabana) REFERENCES cabanas(id) ON DELETE RESTRICT,
    
    CONSTRAINT chk_fechas CHECK (fecha_fin >= fecha_inicio)
);

CREATE TABLE IF NOT EXISTS pagos (
    id SERIAL PRIMARY KEY,
    id_usuario INT NOT NULL,
    cantidad NUMERIC(10, 2) NOT NULL,
    tipo_pago VARCHAR(50) NOT NULL,
    fecha TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    id_reservacion INT NOT NULL,

    CONSTRAINT fk_pago_usuario FOREIGN KEY (id_usuario) REFERENCES usuarios(id) ON DELETE RESTRICT,

    CONSTRAINT fk_pago_reservacion FOREIGN KEY (id_reservacion) REFERENCES reservaciones(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS comentario (
    id SERIAL PRIMARY KEY,
    id_usuario INT NOT NULL,          
    id_usuario_comentado INT NOT NULL,  
    id_cabana INT NOT NULL, -- Modifique la ñ por n            
    comentario TEXT NOT NULL,          
    fecha_creacion TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, 

    CONSTRAINT fk_comentario_usuario 
        FOREIGN KEY (id_usuario) REFERENCES usuarios(id) ON DELETE CASCADE,
        
    CONSTRAINT fk_comentario_usuario_comentado 
        FOREIGN KEY (id_usuario_comentado) REFERENCES usuarios(id) ON DELETE CASCADE,
        
    CONSTRAINT fk_comentario_cabana 
        FOREIGN KEY (id_cabana) REFERENCES cabanas(id) ON DELETE CASCADE -- Cambie la ñ por n
);

INSERT INTO roles (id, tipo) VALUES 
(1, 'Administrador'), 
(2, 'Usuario');

INSERT INTO usuarios (nombre, telefono, direccion, dpi, correo, contrasena, id_rol) VALUES 
('Admin', '123456789', 'Direccion Admin', '1234567890123', 'admin@admin.com', 'password', 1);
