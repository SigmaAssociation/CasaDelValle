package reservations

type Reservation struct {
	ID        int    `json:"id"`
	UserID    uint   `json:"user_id"`
	CabinID   uint   `json:"cabin_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Available *bool  `json:"available,omitempty"`
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
