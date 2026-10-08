-- Migración de roles: separa Huésped (2) y Anfitrión (3).
-- Los usuarios que ya publicaron cabañas se convierten en anfitriones;
-- el resto queda como huésped. Solo para BDs creadas antes de este cambio
-- (init.sql ya incluye los roles nuevos para BDs frescas).
-- Uso: docker exec -i postgres_db psql -U devuser -d appdb < database/migrate_roles.sql

INSERT INTO roles (id, tipo) VALUES (3, 'Anfitrión')
ON CONFLICT (id) DO NOTHING;

UPDATE roles SET tipo = 'Huésped' WHERE id = 2 AND tipo = 'Usuario';

UPDATE usuarios
SET id_rol = 3
WHERE id_rol = 2
  AND id IN (SELECT DISTINCT id_anfitrion FROM cabanas);
