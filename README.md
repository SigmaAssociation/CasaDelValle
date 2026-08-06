
# CabañasApp

> Plataforma web para la reserva, gestión y renta de cabañas, desarrollada para la asignatura de **Teoría de Sistemas**.

---

## Stack Tecnológico

El proyecto está diseñado sobre una arquitectura desacoplada utilizando el siguiente stack:

* **Frontend:** [Angular](https://angular.io/) (Single Page Application / TypeScript)
* **Backend:** [Go / Golang](https://go.dev/) (API REST)
* **Base de Datos:** [PostgreSQL](https://www.postgresql.org/) (Base de datos relacional)
* **Contenedorización:** [Docker](https://www.docker.com/) & Docker Compose
* [Enlace a diagrama ER](https://drive.google.com/file/d/1gaKiyiXwn-TsxSD6TGDgG0UB_bjFq39I/view?usp=drivesdk)
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


### Pasos para iniciar el entorno de desarrollo

1. **Clonar el repositorio:**
