
# CabañasApp

> Plataforma web para la reserva, gestión y renta de cabañas, desarrollada para la asignatura de **Teoría de Sistemas**.

---

## Stack Tecnológico

El proyecto está diseñado sobre una arquitectura desacoplada utilizando el siguiente stack:

* **Frontend:** [Angular](https://angular.io/) (Single Page Application / TypeScript)
* **Backend:** [Go / Golang](https://go.dev/) (API REST)
* **Base de Datos:** [PostgreSQL](https://www.postgresql.org/) (Base de datos relacional)
* **Contenedorización:** [Docker](https://www.docker.com/) & Docker Compose
* **Diagrama ER:** [Enlace al diagrama ER](https://drive.google.com/file/d/1gaKiyiXwn-TsxSD6TGDgG0UB_bjFq39I/view?usp=drivesdk)
---

## Arquitectura y Módulos

El sistema se divide en los siguientes subsistemas: 

* **Módulo de Autenticación y Usuarios:** Control de acceso, perfiles de huésped y anfitrión.
* **Módulo de Catálogo:** Gestión de cabañas, amenidades y filtros de búsqueda.
* **Módulo de Reservas y Disponibilidad:** Gestión de fechas y estados de reserva.
* **Módulo de Pagos:** Simulación / gestión del flujo transaccional.
* **Módulo de Reseñas y Feedback:** Calificación del servicio por parte de los usuarios.

---

## Ejecución con Docker (Entorno Local)

Todo el entorno (Base de Datos, Backend y Frontend) se puede levantar con **Docker Compose**.

### Prerrequisitos

- [Docker](https://www.docker.com/products/docker-desktop/) instalado
- [Docker Compose](https://docs.docker.com/compose/) disponible


### Pasos para iniciar el entorno de desarrollo

1. **Clonar el repositorio:**
   ```bash
   git clone https://github.com/SigmaAssociation/CasaDelValle.git
   ```

2. **Desde este repositorio (`CasaDelValle`), iniciar los servicios:**

   ```bash
   docker compose up --build
   ```

3. **Acceder a los servicios:**
   - Frontend: `http://localhost:4200`
   - Backend API: `http://localhost:8080`
   - PostgreSQL: `postgres-db:5432` dentro de la red de Docker (sin puerto publicado al equipo local)

4. **Detener el entorno:**

   ```bash
   docker compose down
   ```

### Migraciones de base de datos

El esquema inicial se crea con `database/init.sql` solamente cuando el volumen de PostgreSQL está vacío. Después, tanto para una instalación nueva como para una base existente, aplica las migraciones de `database/migrations/` con el script de desarrollo.

1. Configura `POSTGRES_USER`, `POSTGRES_PASSWORD` y `POSTGRES_DB` en el archivo `.env` de la raíz (y las variables `DB_*` correspondientes para la API).
2. Levanta PostgreSQL y espera a que esté listo:

   ```bash
   docker compose up -d --build --wait postgres-db
   ```

   Si ya tenías el contenedor creado, este comando también incorpora el montaje de los archivos de migraciones.

3. Ejecuta desde la raíz del repositorio:

   ```bash
   ./database/migrate.sh
   ```

   El script usa el servicio `postgres-db` de `docker-compose.yml` y las credenciales `POSTGRES_USER` y `POSTGRES_DB` del contenedor. No requiere instalar `psql` en tu equipo. También puedes invocarlo desde otra carpeta usando su ruta absoluta. Si tu instalación requiere `sudo` para acceder a Docker, usa `sudo ./database/migrate.sh`.

   Opcionalmente puedes indicar ambos valores: `./database/migrate.sh USUARIO BASE_DE_DATOS`.

4. Inicia los demás servicios:

   ```bash
   docker compose up -d --build
   ```

Los archivos se ejecutan en orden alfabético. El historial se guarda en la tabla `public.schema_migrations` de cada base de datos, dentro del mismo volumen que sus datos. Puedes volver a ejecutar el script: omite las migraciones registradas. Un bloqueo en PostgreSQL evita ejecuciones simultáneas. Si una migración falla, se revierten sus cambios y su registro, el script termina con error y las migraciones posteriores no se ejecutan; las anteriores permanecen aplicadas.

Para consultar el historial:

```bash
docker compose exec postgres-db sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "TABLE public.schema_migrations;"'
```

`docker compose down` conserva los datos y el historial. **`docker compose down -v` los elimina**; no es necesario borrar el volumen para ejecutar migraciones.

### Gestionar precios por temporada

Desde **Mis Cabañas → Ver detalles → Precios por temporada** se abre un apartado independiente de **Editar cabaña**, reservado al propietario o administrador.

Selecciona una temporada global para completar sus fechas y mostrar el **precio recomendado por noche** calculado sobre la tarifa base de la cabaña. Puedes usar esa recomendación, cambiar el porcentaje o elegir un precio fijo. Para un periodo propio, selecciona **Personalizada** e ingresa la primera y última noche; ambas fechas están incluidas. La vista previa muestra el precio que se guardará, y la lista inferior permite consultar o eliminar las tarifas configuradas.

Los abonos posteriores al anticipo se registran desde **Mis reservaciones → Agregar pago**, con tarjeta, transferencia o efectivo, hasta cubrir el saldo pendiente.

### Documentación Swagger

La UI de Swagger está en `http://localhost:8080/swagger/index.html` (solo con el backend en ejecución). Los archivos de `cdv-api/docs/` se generan a partir de las anotaciones del código, **no se editan a mano**. Cada vez que agregues o modifiques un endpoint, regenéralos desde `cdv-api/`:

```bash
cd cdv-api
swag init -g cmd/api/main.go -o docs
```

> El CI verifica que `cdv-api/docs/` esté al día y falla si olvidas regenerarlo.

### Manual de usuario

La guía para huéspedes está disponible en [docs/manual-usuario.md](docs/manual-usuario.md).

## Estilo de diseño

La aplicación usara la siguiente paleta de colores y se puede utilizar con su respectiva clase:
- #4F5D2F -----> bg-green
- #423629 -----> bg-brown
- #EC4E20 -----> bg-orange
- #EFE6D2 -----> bg-cream
- #151515 -----> bg-black

Se utilizará la siguiente fuente para titulos y encabezados: Serif ---> font-serif

Se utilizará la siguiente fuente para cuerpo de texto, formularios, etc: Sans ---> font-sans

## Guias

Se habilitaron paginas temporales con una guía más detallada de como cumplir con los requerimientos manteniendo un mismo estándar.

Se podran encontrar en las siguientes rutas del proyecto FE:
- guide
- guide/user-story
- guide/model
- guide/design
- guide/backend
