package refunds

import (
	"errors"
	"time"
)

const (
	RefundFullDays    = 7
	RefundPartialDays = 3

	RefundPercentFull    = 1.00
	RefundPercentPartial = 0.70
	RefundPercentNone    = 0.00

	RefundType = "reembolso"

	RefundMethod = "transferencia"

	PaymentTypeAnticipo = "anticipo"
	PaymentTypePago     = "pago"

	KindTotal   = "total"
	KindParcial = "parcial"
)

var (
	ErrReservationNotFound      = errors.New("reservation not found")
	ErrReservationNotRefundable = errors.New("reservation not refundable")
	ErrRefundExceedsAvailable   = errors.New("refund exceeds available")
	ErrRefundInvalidAmount      = errors.New("refund invalid amount")
	ErrNothingRefundable        = errors.New("nothing refundable")
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type Refund struct {
	ID            int       `json:"id"`
	ReservationID int       `json:"reservationId"`
	HostID        uint      `json:"hostId"`
	Amount        float64   `json:"amount"`
	Method        string    `json:"method"`
	CreatedAt     time.Time `json:"createdAt"`
}

type FinancialMovement struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	Subtype   string    `json:"subtype"`
	Amount    float64   `json:"amount"`
	Method    string    `json:"method"`
	CreatedAt time.Time `json:"createdAt"`
}

type RefundsHistoryResponse struct {
	Movements           []FinancialMovement `json:"movements"`
	TotalPaid           float64             `json:"totalPaid"`
	TotalRefunded       float64             `json:"totalRefunded"`
	RefundableAvailable float64             `json:"refundableAvailable"`
}

type CreateRefundRequest struct {
	Amount *float64 `json:"amount,omitempty"`
}

type CreateRefundResponse struct {
	Message string  `json:"message"`
	Refund  *Refund `json:"refund"`
}
