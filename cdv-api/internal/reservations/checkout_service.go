package reservations

import (
	"context"
	"errors"
	"time"
)

// Checkout registra el fin de la estadía y pasa la reservación a finalizada.
// Solo aplica a reservaciones en curso que tengan check-in registrado.
// Reutiliza GetCheckinEligibility porque ya entrega estado y fecha de check-in.
func (s *Service) Checkout(ctx context.Context, id int) (*ReservationDetail, error) {
	if id <= 0 {
		return nil, errors.New("ID de reservación inválido")
	}

	state, err := s.repository.GetCheckinEligibility(ctx, id)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, ErrReservationNotFound
	}

	if state.Status != StatusInProgress {
		return nil, ErrReservationNotCheckoutable
	}
	if state.CheckedInAt == nil {
		return nil, ErrCheckinRequired
	}

	if err := s.repository.RecordCheckout(ctx, id, time.Now()); err != nil {
		return nil, err
	}

	updated, err := s.repository.GetReservationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.notifyReservationFinished(ctx, updated)

	return updated, nil
}
