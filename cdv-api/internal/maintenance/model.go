package maintenance

import (
	"errors"
	"time"
)

var (
	ErrNotFound    = errors.New("Cabaña, mantenimiento o solicitud pendiente no encontrada")
	ErrForbidden   = errors.New("No tienes permiso para realizar esta operación")
	ErrConflict    = errors.New("La disponibilidad o el estado de la reserva cambió; actualiza la información")
	ErrUnavailable = errors.New("Las nuevas fechas no están disponibles para esta cabaña o se cruzan con otra de tus reservas")
)

type Maintenance struct {
	ID          int        `json:"id"`
	CabinID     int        `json:"cabin_id"`
	StartDate   string     `json:"start_date"`
	EndDate     string     `json:"end_date"`
	Reason      string     `json:"reason"`
	CreatedAt   time.Time  `json:"created_at"`
	CancelledAt *time.Time `json:"cancelled_at"`
}
type DateRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}
type CreateRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Reason    string `json:"reason"`
}
type AffectedReservation struct {
	ID         int     `json:"id"`
	GuestID    int     `json:"guest_id"`
	GuestName  string  `json:"guest_name"`
	StartDate  string  `json:"start_date"`
	EndDate    string  `json:"end_date"`
	Status     string  `json:"status"`
	TotalPrice float64 `json:"total_price"`
}
type Preview struct {
	Reservations []AffectedReservation `json:"reservations"`
}
type DecisionRequest struct {
	Action    string `json:"action"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}
type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}
type PendingDecision struct {
	ReservationID  int         `json:"reservation_id"`
	CabinID        int         `json:"cabin_id"`
	Reason         string      `json:"reason"`
	StartDate      string      `json:"start_date"`
	EndDate        string      `json:"end_date"`
	Nights         int         `json:"nights"`
	TotalPrice     float64     `json:"total_price"`
	Refundable     float64     `json:"refundable"`
	BlockedRanges  []DateRange `json:"blocked_ranges"`
	GuestID        uint        `json:"-"`
	HostID         uint        `json:"-"`
	MaintenanceID  int         `json:"-"`
	PreviousStatus string      `json:"-"`
}
type DecisionResponse struct {
	Message string `json:"message"`
}
type ErrorResponse struct {
	Message string `json:"message"`
}
type validationError string

func (e validationError) Error() string { return string(e) }
