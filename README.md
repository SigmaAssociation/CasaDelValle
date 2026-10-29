
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
   - PostgreSQL: `localhost:5432`

4. **Detener el entorno:**

   ```bash
   docker compose down
   ```

### Migraciones de base de datos

Las migraciones se ejecutan manualmente con el script `database/migrate.sh`. El script recorre los archivos `.sql` de `database/migrations/` en orden alfabético y guarda los ya aplicados en `database/completed_migrations.txt`.

Levantar el proyecto con docker, ubicarse en la carpeta raiz y ejecutar:

```bash
sudo ./database/migrate.sh USUARIO BASE_DE_DATOS
```

Por ejemplo, `./database/migrate.sh devuser appdb`. Para agregar una migración, cree un archivo con el siguiente consecutivo en `database/migrations/`, por ejemplo `007_descripcion_del_cambio.sql`. No modifique ni elimine archivos ya registrados en `database/completed_migrations.txt`.

### Documentación Swagger

La UI de Swagger está en `http://localhost:8080/swagger/index.html` (solo con el backend en ejecución). Los archivos de `cdv-api/docs/` se generan a partir de las anotaciones del código, **no se editan a mano**. Cada vez que agregues o modifiques un endpoint, regenéralos desde `cdv-api/`:

```bash
cd cdv-api
swag init -g cmd/api/main.go -o docs
```

> El CI verifica que `cdv-api/docs/` esté al día y falla si olvidas regenerarlo.

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
