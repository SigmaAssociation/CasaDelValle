package reservations

import "time"

const (
	PaymentStatusPending  = "pendiente"
	PaymentStatusPaid     = "pagado"
	PaymentStatusCredited = "saldo_a_favor"
)

type ExtraCharge struct {
	ID        int       `json:"id"`
	Concept   string    `json:"concept"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"createdAt"`
}

type Payment struct {
	ID     int        `json:"id"`
	Amount float64    `json:"amount"`
	Type   string     `json:"type"`
	PaidAt *time.Time `json:"paidAt"`
}

type Refund struct {
	ID         int       `json:"id"`
	PaymentID  *int      `json:"paymentId"`
	Amount     float64   `json:"amount"`
	Reason     *string   `json:"reason"`
	RefundedAt time.Time `json:"refundedAt"`
}

// ReservationLedger agrupa los datos crudos que alimentan el saldo.
type ReservationLedger struct {
	TotalPrice   *float64
	NightlyRate  float64
	ExtraCharges []ExtraCharge
	Payments     []Payment
	Refunds      []Refund
}

// ReservationBalance es la situación financiera de una reservación.
//
//	grandTotal     = baseTotal + extraChargesTotal
//	netPaid        = paymentsTotal - refundsTotal
//	pendingBalance = grandTotal - netPaid
//
// pendingBalance > 0: el huésped debe; = 0: saldada; < 0: saldo a favor.
type ReservationBalance struct {
	ReservationID     int           `json:"reservationId"`
	Status            string        `json:"status"`
	Nights            int           `json:"nights"`
	NightlyRate       float64       `json:"nightlyRate"`
	BaseTotal         float64       `json:"baseTotal"`
	ExtraChargesTotal float64       `json:"extraChargesTotal"`
	GrandTotal        float64       `json:"grandTotal"`
	PaymentsTotal     float64       `json:"paymentsTotal"`
	RefundsTotal      float64       `json:"refundsTotal"`
	NetPaid           float64       `json:"netPaid"`
	PendingBalance    float64       `json:"pendingBalance"`
	PaymentStatus     string        `json:"paymentStatus"`
	ExtraCharges      []ExtraCharge `json:"extraCharges"`
	Payments          []Payment     `json:"payments"`
	Refunds           []Refund      `json:"refunds"`
}
