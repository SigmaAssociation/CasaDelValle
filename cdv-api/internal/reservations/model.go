package reservations

import (
	"errors"
	"time"
)

type Reservation struct {
	ID        int    `json:"id"`
	UserID    uint   `json:"user_id"`
	CabinID   uint   `json:"cabin_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Status    string `json:"status,omitempty"`
}

type ReservationRequest struct {
	UserID    uint   `json:"user_id"`
	CabinID   uint   `json:"cabin_id"`
	StartDate string `json:"start_date"` // Formato AAAA-MM-DD
	EndDate   string `json:"end_date"`   // Formato AAAA-MM-DD
}

type CreateReservationResponse struct {
	Message       string `json:"message"`
	ReservationID int    `json:"reservation_id,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

// Estados del ciclo de vida de una reservación. Nace pendiente; el pago del
// anticipo la confirma; el job en segundo plano la pasa a en_curso al iniciar
// la estancia y a finalizada al terminar. Cancelada es un estado terminal.
const (
	StatusPending    = "pendiente"
	StatusConfirmed  = "confirmada"
	StatusInProgress = "en_curso"
	StatusFinished   = "finalizada"
	StatusCancelled  = "cancelada"

	MinCancellationDays = 3

	// AnticipoPercent es la fracción del total de la reserva que debe
	// pagarse como anticipo para confirmarla.
	AnticipoPercent = 0.30
)

var (
	ErrReservationNotFound       = errors.New("reservation not found")
	ErrReservationNotCancellable = errors.New("reservation not cancellable")
	ErrReservationNotEditable    = errors.New("reservation not editable")
	ErrReservationNotPending     = errors.New("reservation not pending")
	ErrCancellationTooLate       = errors.New("cancellation too late")
	ErrCabinNotFound             = errors.New("cabin not found")
)

type UpdateReservationRequest struct {
	CabinID   uint   `json:"cabin_id"`
	StartDate string `json:"start_date"` // Formato AAAA-MM-DD
	EndDate   string `json:"end_date"`   // Formato AAAA-MM-DD
}

type ReservationDetail struct {
	ID          int        `json:"id"`
	UserID      uint       `json:"user_id"`
	CabinID     uint       `json:"cabin_id"`
	HostID      uint       `json:"host_id"`
	StartDate   string     `json:"start_date"`
	EndDate     string     `json:"end_date"`
	Status      string     `json:"status"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`
}

type CancelReservationResponse struct {
	Message     string             `json:"message"`
	Reservation *ReservationDetail `json:"reservation"`
}

// PayAnticipoResponse devuelve el monto del anticipo registrado y la
// reservación ya confirmada.
type PayAnticipoResponse struct {
	Message     string             `json:"message"`
	Amount      float64            `json:"amount"`
	Reservation *ReservationDetail `json:"reservation"`
}

type ReservationCard struct {
	ID            int        `json:"id"`
	CabinID       int        `json:"cabinId"`
	CabinName     string     `json:"cabinName"`
	CabinImageURL *string    `json:"cabinImageUrl,omitempty"`
	GuestID       int        `json:"guestId"`
	GuestName     string     `json:"guestName"`
	StartDate     string     `json:"startDate"` 
	EndDate       string     `json:"endDate"`
	TotalPrice    float64    `json:"totalPrice"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	CancelledAt   *time.Time `json:"cancelledAt"`
}
type UpdateReservationResponse struct {
	Message     string             `json:"message"`
	Reservation *ReservationDetail `json:"reservation"`
}
