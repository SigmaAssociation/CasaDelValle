package reservations

import (
	"context"
	"math"
)

func toCents(v float64) int64   { return int64(math.Round(v * 100)) }
func fromCents(c int64) float64 { return float64(c) / 100 }

// GetReservationBalance calcula el saldo en cada consulta a partir de los
// movimientos registrados y del precio_total contratado; así refleja el
// último pago, cargo o reembolso sin recotizar las noches de la reserva.
// Los montos se suman en centavos (enteros) para evitar errores de float.
func (s *Service) GetReservationBalance(ctx context.Context, reservation *ReservationDetail) (*ReservationBalance, error) {
	if reservation == nil {
		return nil, ErrReservationNotFound
	}

	ledger, err := s.repository.GetReservationLedger(ctx, reservation.ID, reservation.CabinID)
	if err != nil {
		return nil, err
	}

	nights, err := calculateReservationNights(reservation.StartDate, reservation.EndDate)
	if err != nil {
		return nil, err
	}

	baseCents := int64(nights) * toCents(ledger.NightlyRate)
	if ledger.TotalPrice != nil {
		baseCents = toCents(*ledger.TotalPrice)
	}

	var chargesCents, paymentsCents, refundsCents int64
	for _, c := range ledger.ExtraCharges {
		chargesCents += toCents(c.Amount)
	}
	for _, p := range ledger.Payments {
		paymentsCents += toCents(p.Amount)
	}
	for _, rf := range ledger.Refunds {
		refundsCents += toCents(rf.Amount)
	}

	grandCents := baseCents + chargesCents
	netPaidCents := paymentsCents - refundsCents
	pendingCents := grandCents - netPaidCents

	status := PaymentStatusPending
	switch {
	case pendingCents == 0:
		status = PaymentStatusPaid
	case pendingCents < 0:
		status = PaymentStatusCredited
	}

	return &ReservationBalance{
		ReservationID:     reservation.ID,
		Status:            reservation.Status,
		Nights:            nights,
		NightlyRate:       ledger.NightlyRate,
		BaseTotal:         fromCents(baseCents),
		ExtraChargesTotal: fromCents(chargesCents),
		GrandTotal:        fromCents(grandCents),
		PaymentsTotal:     fromCents(paymentsCents),
		RefundsTotal:      fromCents(refundsCents),
		NetPaid:           fromCents(netPaidCents),
		PendingBalance:    fromCents(pendingCents),
		PaymentStatus:     status,
		ExtraCharges:      ledger.ExtraCharges,
		Payments:          ledger.Payments,
		Refunds:           ledger.Refunds,
	}, nil
}
