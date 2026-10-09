package notifications

import (
	"errors"
	"time"
)

type Notification struct {
	ID            int       `json:"id"`
	UserID        uint      `json:"user_id"`
	Tipo          string    `json:"tipo"`
	Mensaje       string    `json:"mensaje"`
	EntidadTipo   string    `json:"entidad_tipo"`
	EntidadID     *int      `json:"entidad_id,omitempty"`
	Leida         bool      `json:"leida"`
	FechaCreacion time.Time `json:"fecha_creacion"`
}

type NotificationInput struct {
	UserID      uint
	Tipo        string
	Mensaje     string
	EntidadTipo string
	EntidadID   int
}

type NotificationListResponse struct {
	Data    []Notification `json:"data"`
	Total   int            `json:"total"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
	HasMore bool           `json:"has_more"`
}

type UnreadCountResponse struct {
	UnreadCount int `json:"unread_count"`
}

type MarkReadResponse struct {
	Message string `json:"message"`
}

type MarkAllReadResponse struct {
	Message string `json:"message"`
	Updated int    `json:"updated"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

const (
	TipoBienvenida         = "bienvenida"
	TipoReservaCreada      = "reserva_creada"
	TipoReservaCancelada   = "reserva_cancelada"
	TipoReservaActualizada = "reserva_actualizada"
	TipoReservaFinalizada  = "reserva_finalizada"
	TipoCabanaCreada       = "cabana_creada"
	TipoCabanaActualizada  = "cabana_actualizada"
	TipoCabanaEliminada    = "cabana_eliminada"
	TipoRolActualizado     = "rol_actualizado"
)

const (
	EntidadReservacion = "reservacion"
	EntidadCabana      = "cabana"
	EntidadUsuario     = "usuario"
)

const (
	MaxTipoLength  = 40
	MaxMensajeSize = 1000

	// Paginación del buzón personal.
	DefaultPageLimit = 20
	MaxPageLimit     = 100
)

var (
	validTipos = map[string]bool{
		TipoBienvenida:         true,
		TipoReservaCreada:      true,
		TipoReservaCancelada:   true,
		TipoReservaActualizada: true,
		TipoReservaFinalizada:  true,
		TipoCabanaCreada:       true,
		TipoCabanaActualizada:  true,
		TipoCabanaEliminada:    true,
		TipoRolActualizado:     true,
	}

	validEntidades = map[string]bool{
		EntidadReservacion: true,
		EntidadCabana:      true,
		EntidadUsuario:     true,
	}

	ErrNotificationNotFound = errors.New("notification not found")
	ErrUserRequired         = errors.New("user required")
)

func IsValidTipo(tipo string) bool {
	return validTipos[tipo]
}

func IsValidEntidad(entidadTipo string) bool {
	if entidadTipo == "" {
		return true
	}
	return validEntidades[entidadTipo]
}

// NormalizePagination aplica los límites del buzón: una página nunca es mayor
// que MaxPageLimit ni el offset negativo.
func NormalizePagination(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = DefaultPageLimit
	}
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
