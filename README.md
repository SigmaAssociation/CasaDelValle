
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
