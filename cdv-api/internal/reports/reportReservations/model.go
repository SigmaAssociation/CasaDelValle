package reportreservations

import "time"


type FiltroReporte struct {
	IDAnfitrion *int
	IDCabana    *int
}


type UsuarioResumen struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

type ReservacionResumen struct {
	ID            int            `json:"id"`
	FechaInicio   time.Time      `json:"fecha_inicio"`
	FechaFin      time.Time      `json:"fecha_fin"`
	Estado        string         `json:"estado"`
	FechaCreacion time.Time      `json:"fecha_creacion"`
	Huesped       UsuarioResumen `json:"huesped"`
}


type CabanaReporte struct {
	ID                 int                  `json:"id"`
	Nombre             string               `json:"nombre"`
	Direccion          string               `json:"direccion"`
	Precio             float64              `json:"precio"`
	Capacidad          int                  `json:"capacidad"`
	Anfitrion          UsuarioResumen       `json:"anfitrion"`
	TotalReservaciones int                  `json:"total_reservaciones"`
	Reservaciones      []ReservacionResumen `json:"reservaciones"`
}


type ReporteReservaciones struct {
	Cabanas            []CabanaReporte `json:"cabanas"`
	TotalCabanas       int             `json:"total_cabanas"`
	TotalReservaciones int             `json:"total_reservaciones"`
	Mensaje            string          `json:"mensaje,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}