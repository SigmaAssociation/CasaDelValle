# Manual Técnico - Casa del Valle

Sistema de gestión de cabañas: reservas con anticipo, check-in, check-out con pago, tarifas por temporada, bloqueo por mantenimiento y reportes.

## 1. Arquitectura del sistema

Servicios independientes desplegados dentro de la red interna `cdv-net`:

* **`frontend-app` (Angular):** Punto de entrada expuesto a internet en los puertos 80 (HTTP) y 443 (HTTPS). Procesa las peticiones del usuario y las redirige hacia el backend.
* **`backend-api` (Go):** API REST interna escuchando en el puerto 8080.
* **`postgres-db` (PostgreSQL 16):** Persistencia de datos. Solo acepta conexiones provenientes del servicio `backend-api`.

En producción solo el frontend queda expuesto (80 y 443); el backend es accesible únicamente dentro de la red Docker. En desarrollo se publican el frontend (4200) y la API (8080). PostgreSQL no publica ningún puerto al equipo: solo es accesible como `postgres-db:5432` dentro de la red de Docker.

## 2. Tecnologías

| Componente | Tecnología |
|---|---|
| Backend | Go 1.26 |
| Frontend | Angular |
| Base de datos | PostgreSQL 16 |
| Servidor web | Nginx 1.27 |
| Documentación de API | swag v1.16.6 |
| Contenedores | Docker y Docker Compose |
| CI/CD | GitHub Actions, imágenes en GHCR |
| Hosting | Máquina virtual en Azure, HTTPS con Let's Encrypt |

## 3. Estructura del proyecto

```
CasaDelValle/
├── cdv-api/                 # Backend en Go
│   ├── cmd/api/main.go      # Punto de entrada
│   ├── docs/                # Swagger generado
│   ├── Dockerfile.dev
│   └── Dockerfile.prod
├── Casa-Del-Valle-FE/       # Frontend Angular
│   ├── nginx.conf
│   ├── Dockerfile.dev
│   └── Dockerfile.prod
├── database/
│   ├── init.sql             # Esquema base y datos iniciales
│   ├── migrations/          # Cambios posteriores (*.sql)
│   ├── migrate.sh           # Aplica las migraciones pendientes
│   └── Dockerfile.postgres  # PostgreSQL con pg_cron (desarrollo)
├── docker-compose.yml       # Desarrollo
├── docker-compose.prod.yml  # Producción
└── .github/workflows/       # CI y CD
```

## 4. Variables de entorno

Se definen en un archivo `.env` en la raíz del repositorio.

| Variable | Descripción |
|---|---|
| `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | Credenciales y nombre de la base de datos |
| `DB_HOST`, `DB_PORT` | Host y puerto de la BD (desarrollo: `postgres-db` y `5432`; en producción están fijos en el compose) |
| `DB_USER`, `DB_PASSWORD`, `DB_NAME` | Credenciales con las que la API se conecta (deben coincidir con las anteriores) |
| `JWT_SECRET` | Clave para firmar los tokens JWT |
| `CORS_ALLOW_ORIGIN` | Origen permitido por CORS (por defecto `*` en producción) |
| `BACKEND_IMAGE_TAG`, `FRONTEND_IMAGE_TAG` | Etiqueta de las imágenes en producción; las asigna el CD (por defecto `latest`) |

Ejemplo para desarrollo:

```env
POSTGRES_USER=cdv_user
POSTGRES_PASSWORD=cambiar_clave
POSTGRES_DB=casadelvalle
DB_HOST=postgres-db
DB_PORT=5432
DB_USER=cdv_user
DB_PASSWORD=cambiar_clave
DB_NAME=casadelvalle
JWT_SECRET=valor_largo_y_aleatorio
```

Se puede generar un `JWT_SECRET` con `openssl rand -hex 32`.

## 5. Levantar el proyecto en desarrollo

Requisitos: Git y Docker con Compose v2. Si tu instalación de Docker requiere permisos de administrador, antepón `sudo` a los comandos.

```bash
git clone https://github.com/SigmaAssociation/CasaDelValle.git
cd CasaDelValle
# crear el archivo .env (sección 4)

docker compose up -d --build --wait postgres-db   # 1. base de datos lista
./database/migrate.sh                             # 2. migraciones
docker compose up -d --build                      # 3. resto de servicios
```

| Servicio | URL |
|---|---|
| Frontend | http://localhost:4200 |
| API | http://localhost:8080 |
| Swagger | http://localhost:8080/swagger/index.html (con el backend en ejecución) |

El código de `cdv-api/` y `Casa-Del-Valle-FE/` se monta como volumen, por lo que los cambios se reflejan sin reconstruir las imágenes.

```bash
docker compose ps                    # estado
docker compose logs -f backend-api   # logs
docker compose down                  # detener (conserva datos e historial de migraciones)
docker compose down -v               # detener y borrar la base de datos
```

## 6. Base de datos

**Esquema base.** `database/init.sql` crea las tablas (usuarios, roles, cabañas, reservaciones, pagos, cargos adicionales, reembolsos, comisiones, valoraciones, comentarios, imágenes y notificaciones) e inserta los tres roles (Administrador, Huésped, Anfitrión) y un usuario administrador inicial. La contraseña de ese usuario debe cambiarse fuera del entorno local. Este script solo se ejecuta cuando el volumen `pgdata` está vacío.

**Migraciones.** Los cambios posteriores están en `database/migrations/` como archivos `.sql` numerados, y se aplican tanto en instalaciones nuevas como en bases existentes con:

```bash
./database/migrate.sh
```

El script usa el servicio `postgres-db` de `docker-compose.yml` y las credenciales del contenedor, por lo que no requiere `psql` en el equipo. Opcionalmente acepta usuario y base de datos: `./database/migrate.sh USUARIO BASE_DE_DATOS`.

- Los archivos se ejecutan en orden alfabético.
- El historial se guarda en la tabla `public.schema_migrations`, dentro del mismo volumen que los datos. Se puede ejecutar varias veces: omite las ya registradas.
- Un bloqueo en PostgreSQL evita ejecuciones simultáneas.
- Si una migración falla, se revierten sus cambios y su registro, el script termina con error y las siguientes no se ejecutan. Las anteriores permanecen aplicadas.

Consultar el historial:

```bash
docker compose exec postgres-db sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "TABLE public.schema_migrations;"'
```

Para agregar un cambio, crear el siguiente archivo consecutivo (por ejemplo `008_descripcion.sql`) en `database/migrations/` y ejecutar de nuevo el script. Reglas:

- No modificar migraciones ya aplicadas.
- Escribir SQL sin `BEGIN`/`COMMIT`: el script administra la transacción.
- No usar operaciones incompatibles con transacciones, como `CREATE INDEX CONCURRENTLY`.

## 7. Documentación de la API

La API se documenta con anotaciones swaggo. Los archivos de `cdv-api/docs/` se generan a partir del código, no se editan a mano, y deben estar en el repositorio. Después de agregar o modificar endpoints o modelos:

```bash
cd cdv-api
go install github.com/swaggo/swag/cmd/swag@v1.16.6
swag init -g cmd/api/main.go -o docs
git add docs
```

## 8. Pipeline CI/CD

Los workflows están en `.github/workflows/`.

**CI.** Se ejecuta en pull requests hacia `main` y `develop`:

| Job | Acción |
|---|---|
| Backend tests | `go test ./...` |
| Frontend tests | `npm ci` y `npm test -- --watch=false` |
| Swagger docs up to date | Regenera Swagger y falla si `cdv-api/docs` difiere de lo commiteado |

**CD.** Se ejecuta en cada push a `main` y de forma manual:

1. `build-images`: construye las imágenes `Dockerfile.prod` del backend y del frontend y las publica en GHCR con el SHA del commit y `latest`.
2. `deploy`: entra por SSH a la VM, ejecuta `git pull --ff-only origin main` en `/home/CasaDelValle`, y levanta los servicios con `docker-compose.prod.yml` usando las imágenes del commit.

Secretos requeridos en GitHub (*Settings → Secrets and variables → Actions*): `GHCR_PAT`, `VM_SSH_PRIVATE_KEY`, `VM_HOST` y `VM_USER`.

## 9. Despliegue en producción

| Servicio | Imagen | Puertos | Volumen |
|---|---|---|---|
| `postgres-db` | `postgres:16-alpine` | ninguno | `pgdata` |
| `backend-api` | `ghcr.io/mynordaniel/casadelvalle-backend` | 8080 (interno) | `uploads_data` |
| `frontend-app` | `ghcr.io/mynordaniel/casadelvalle-frontend` | 80, 443 | `/etc/letsencrypt` (lectura) |

**Preparación de la VM (una vez):** instalar Docker, abrir los puertos 22, 80 y 443, clonar el repositorio en `/home/CasaDelValle`, crear el `.env` de producción en esa carpeta, obtener el certificado con certbot, autorizar la llave pública del pipeline en `~/.ssh/authorized_keys` y, si las imágenes son privadas, ejecutar `docker login ghcr.io`.

**Despliegue manual:**

```bash
cd /home/CasaDelValle
git pull --ff-only origin main
export BACKEND_IMAGE_TAG=latest FRONTEND_IMAGE_TAG=latest
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d --no-build --remove-orphans
```

**Migraciones.** El pipeline no las ejecuta. Cuando una versión incluya migraciones nuevas, se aplican desde la VM:

```bash
for f in database/migrations/*.sql; do
  docker exec -i postgres_db sh -c 'psql -X -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"' < "$f"
done
```

**Verificación:**

```bash
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs --tail=50 backend-api
curl -I https://<dominio>
```

**Volver a una versión anterior:** exportar `BACKEND_IMAGE_TAG` y `FRONTEND_IMAGE_TAG` con el SHA deseado y ejecutar `up -d --no-build`. Esto no revierte cambios de base de datos.

## 10. Respaldo y restauración

Se respaldan la base de datos y las imágenes subidas (`uploads_data`).

```bash
# Respaldo de la base de datos
mkdir -p backups
docker exec postgres_db sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" --clean --if-exists' \
  > backups/cdv_$(date +%F_%H%M).sql

# Restauración
docker compose -f docker-compose.prod.yml stop backend-api
docker exec -i postgres_db sh -c 'psql -X -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
  < backups/cdv_<fecha>.sql
docker compose -f docker-compose.prod.yml start backend-api

# Respaldo de imágenes subidas (confirmar el nombre del volumen con docker volume ls)
docker run --rm -v casadelvalle_uploads_data:/data -v "$(pwd)/backups":/backup alpine \
  tar czf /backup/uploads_$(date +%F).tar.gz -C /data .
```

Los respaldos contienen datos personales: no se suben al repositorio y conviene copiarlos fuera de la VM.

## 11. Solución de problemas

| Problema | Solución |
|---|---|
| CI falla en "Swagger docs up to date" | Regenerar con `swag init -g cmd/api/main.go -o docs` (swag v1.16.6) y commitear `docs/` |
| `git push` rechazado (*fetch first*) | `git pull --rebase origin <rama>` y volver a hacer push |
| La API no conecta a la BD | Revisar las variables `DB_*` del `.env` y que `postgres_db` esté *healthy* |
| Cambios en `init.sql` no se aplican | Solo corre con el volumen vacío: en desarrollo usar `docker compose down -v`; para cambios de esquema crear una migración |
| Faltan tablas | Aplicar las migraciones y revisar `public.schema_migrations` |
| Error de CORS | Ajustar `CORS_ALLOW_ORIGIN` al origen del frontend y reiniciar `backend-api` |
| HTTPS no funciona | Verificar el certificado en `/etc/letsencrypt` (`certbot renew`) y reiniciar `frontend-app` |
| El CD falla en el paso SSH | Revisar los secretos `VM_*` y `authorized_keys` en la VM |
| `git pull --ff-only` falla en la VM | Dejar limpio el repositorio de la VM (`git status`) |
| Se pierden imágenes subidas | No usar `down -v` en producción; restaurar desde respaldo |

