package reportincome

import "time"

type ReportFilter struct {
	StartDate time.Time
	EndDate   time.Time
	CabinID   *int
	HostID    *uint
}

type IncomeMovement struct {
	ID            int       `json:"id"`
	ReservationID int       `json:"reservation_id"`
	CabinID       int       `json:"cabin_id"`
	CabinName     string    `json:"cabin_name"`
	GuestName     string    `json:"guest_name"`
	Type          string    `json:"type"`
	Method        string    `json:"method"`
	Amount        float64   `json:"amount"`
	Date          time.Time `json:"date"`
}

type IncomeReport struct {
	Movements      []IncomeMovement `json:"movements"`
	TotalPayments  float64          `json:"total_payments"`
	TotalRefunds   float64          `json:"total_refunds"`
	TotalIncome    float64          `json:"total_income"`
	TotalMovements int              `json:"total_movements"`
	StartDate      string           `json:"start_date"`
	EndDate        string           `json:"end_date"`
	Message        string           `json:"message,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}
