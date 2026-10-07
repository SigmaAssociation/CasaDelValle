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

const (
	StatusActive        = "activa"
	StatusCancelled     = "cancelada"
	StatusFinished      = "finalizada"
	MinCancellationDays = 3
)

var (
	ErrReservationNotFound       = errors.New("reservation not found")
	ErrReservationNotCancellable = errors.New("reservation not cancellable")
	ErrReservationNotEditable    = errors.New("reservation not editable")
	ErrCancellationTooLate       = errors.New("cancellation too late")
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

type UpdateReservationResponse struct {
	Message     string             `json:"message"`
	Reservation *ReservationDetail `json:"reservation"`
}
