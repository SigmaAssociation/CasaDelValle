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
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type CreateReservationResponse struct {
	Message       string `json:"message"`
	ReservationID int    `json:"reservation_id,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

const (
	StatusPending     = "pendiente"
	StatusConfirmed   = "confirmada"
	StatusInProgress  = "en_curso"
	StatusFinished    = "finalizada"
	StatusCancelled   = "cancelada"
	StatusMaintenance = "por_mantenimiento"

	MinCancellationDays = 3

	AnticipoPercent = 0.30
)

var (
	ErrReservationNotFound       = errors.New("reservation not found")
	ErrReservationNotCancellable = errors.New("reservation not cancellable")
	ErrReservationNotEditable    = errors.New("reservation not editable")
	ErrReservationNotPending     = errors.New("reservation not pending")
	ErrReservationNotCheckinable = errors.New("reservation not checkinable")
	ErrCheckinPaymentIncomplete  = errors.New("checkin payment incomplete")
	ErrCabinNotFound             = errors.New("cabin not found")
)

type UpdateReservationRequest struct {
	CabinID   uint   `json:"cabin_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type ReservationDetail struct {
	TotalPrice  *float64   `json:"precio_total,omitempty"`
	NightlyRate *float64   `json:"nightly_rate,omitempty"`
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
	Refund      *CancelRefund      `json:"refund,omitempty"`
}

type CancelRefund struct {
	Amount  float64 `json:"amount"`
	Percent float64 `json:"percent"`
	Kind    string  `json:"kind"`
}

type PayAnticipoResponse struct {
	Message     string             `json:"message"`
	Amount      float64            `json:"amount"`
	Reservation *ReservationDetail `json:"reservation"`
}

type ReservationCard struct {
	ID            int        `json:"id"`
	CabinID       int        `json:"cabinId"`
	HostID        int        `json:"hostId"`
	CabinName     string     `json:"cabinName"`
	CabinImageURL *string    `json:"cabinImageUrl,omitempty"`
	GuestID       int        `json:"guestId"`
	GuestName     string     `json:"guestName"`
	StartDate     string     `json:"startDate"`
	EndDate       string     `json:"endDate"`
	TotalPrice    float64    `json:"totalPrice"`
	PaidTotal     float64    `json:"paidTotal"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	CancelledAt   *time.Time `json:"cancelledAt"`
}
type UpdateReservationResponse struct {
	Message     string             `json:"message"`
	Reservation *ReservationDetail `json:"reservation"`
}

type CheckinEligibility struct {
	TotalPrice    *float64
	ReservationID int
	UserID        uint
	CabinID       uint
	HostID        uint
	Status        string
	StartDate     string
	EndDate       string
	CheckedInAt   *time.Time
	NightlyRate   float64
	PaidTotal     float64
}

type CheckinReservationResponse struct {
	Message     string             `json:"message"`
	Reservation *ReservationDetail `json:"reservation"`
}

const (
	PaymentMethodCard     = "tarjeta"
	PaymentMethodTransfer = "transferencia"
	PaymentMethodCash     = "efectivo"
)

func IsValidPaymentMethod(method string) bool {
	switch method {
	case PaymentMethodCard, PaymentMethodTransfer, PaymentMethodCash:
		return true
	default:
		return false
	}
}

type PayAnticipoRequest struct {
	Amount float64 `json:"amount,omitempty"`
	Method string  `json:"method,omitempty"`
}

type CheckinRequest struct {
	CashAmount float64 `json:"cash_amount,omitempty"`
}

type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type CabinAvailabilityResponse struct {
	CabinID       uint        `json:"cabinId"`
	BlockedRanges []DateRange `json:"blockedRanges"`
}
