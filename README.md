
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
2. **Clonar también el frontend en la carpeta hermana esperada por `docker-compose.yml`:**
   ```bash
   git clone https://github.com/SigmaAssociation/CDV_Frontend.git
   ```

   ```bash
   # Estructura esperada
   <carpeta-base>/
   ├── CasaDelValle
   └── CDV_Frontend/
       └── Casa-Del-Valle-FE
   ```

3. **Desde este repositorio (`CasaDelValle`), iniciar los servicios:**

   ```bash
   docker compose up --build
   ```

4. **Acceder a los servicios:**
   - Frontend: `http://localhost:4200`
   - Backend API: `http://localhost:8080`
   - PostgreSQL: `localhost:5432`

5. **Detener el entorno:**

   ```bash
   docker compose down
   ```
