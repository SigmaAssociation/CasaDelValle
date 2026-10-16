package reservations

import "errors"

var (
	// ErrReservationNotCheckoutable: la reservación no está en curso
	// (pendiente, confirmada, ya finalizada o cancelada).
	ErrReservationNotCheckoutable = errors.New("reservacion no aplica para checkout")

	// ErrCheckinRequired: la reservación está en curso pero no tiene
	// registrado el check-in, requisito para poder hacer el check-out.
	ErrCheckinRequired = errors.New("checkin requerido antes de checkout")
)

type CheckoutReservationResponse struct {
	Message     string             `json:"message"`
	Reservation *ReservationDetail `json:"reservation"`
}
