package reportreservations

import "time"

type ReportFilter struct {
	HostID      *int
	CabinID     *int
	GuestID     *int
	Status      string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

type UserSummary struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ReservationSummary struct {
	ID        int         `json:"id"`
	StartDate time.Time   `json:"start_date"`
	EndDate   time.Time   `json:"end_date"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	Guest     UserSummary `json:"guest"`
}

type CabinReport struct {
	ID                int                  `json:"id"`
	Name              string               `json:"name"`
	Address           string               `json:"address"`
	Price             float64              `json:"price"`
	Capacity          int                  `json:"capacity"`
	Host              UserSummary          `json:"host"`
	TotalReservations int                  `json:"total_reservations"`
	Reservations      []ReservationSummary `json:"reservations"`
}

type ReservationsReport struct {
	Cabins            []CabinReport `json:"cabins"`
	TotalCabins       int           `json:"total_cabins"`
	TotalReservations int           `json:"total_reservations"`
	Message           string        `json:"message,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}