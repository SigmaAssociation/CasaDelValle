package reservations

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) GetReservationLedger(ctx context.Context, reservationID int, cabinID uint) (*ReservationLedger, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	ledger := &ReservationLedger{
		ExtraCharges: []ExtraCharge{},
		Payments:     []Payment{},
		Refunds:      []Refund{},
	}

	if err := tx.QueryRow(ctx, `SELECT r.precio_total/(r.fecha_fin-r.fecha_inicio),r.precio_total FROM reservaciones r JOIN cabanas c ON c.id=r.id_cabana WHERE r.id=$1 AND c.id=$2`, reservationID, cabinID).Scan(&ledger.NightlyRate, &ledger.TotalPrice); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCabinNotFound
		}
		return nil, err
	}

	chargeRows, err := tx.Query(ctx, `
		SELECT id, concepto, monto, fecha
		FROM cargos_adicionales
		WHERE id_reservacion = $1
		ORDER BY fecha ASC, id ASC`, reservationID)
	if err != nil {
		return nil, err
	}
	for chargeRows.Next() {
		var c ExtraCharge
		if err := chargeRows.Scan(&c.ID, &c.Concept, &c.Amount, &c.CreatedAt); err != nil {
			chargeRows.Close()
			return nil, err
		}
		ledger.ExtraCharges = append(ledger.ExtraCharges, c)
	}
	chargeRows.Close()
	if err := chargeRows.Err(); err != nil {
		return nil, err
	}

	paymentRows, err := tx.Query(ctx, `
		SELECT id, cantidad, tipo_pago, fecha
		FROM pagos
		WHERE id_reservacion = $1 AND tipo_pago <> 'reembolso'
		ORDER BY fecha ASC, id ASC`, reservationID)
	if err != nil {
		return nil, err
	}
	for paymentRows.Next() {
		var p Payment
		if err := paymentRows.Scan(&p.ID, &p.Amount, &p.Type, &p.PaidAt); err != nil {
			paymentRows.Close()
			return nil, err
		}
		ledger.Payments = append(ledger.Payments, p)
	}
	paymentRows.Close()
	if err := paymentRows.Err(); err != nil {
		return nil, err
	}

	refundRows, err := tx.Query(ctx, `
		SELECT id, id_pago, cantidad, motivo, fecha
		FROM reembolsos
		WHERE id_reservacion = $1
		UNION ALL
		SELECT id, NULL::int AS id_pago, cantidad, NULL::text AS motivo, fecha
		FROM pagos
		WHERE id_reservacion = $1 AND tipo_pago = 'reembolso'
		ORDER BY fecha ASC, id ASC`, reservationID)
	if err != nil {
		return nil, err
	}
	for refundRows.Next() {
		var rf Refund
		if err := refundRows.Scan(&rf.ID, &rf.PaymentID, &rf.Amount, &rf.Reason, &rf.RefundedAt); err != nil {
			refundRows.Close()
			return nil, err
		}
		ledger.Refunds = append(ledger.Refunds, rf)
	}
	refundRows.Close()
	if err := refundRows.Err(); err != nil {
		return nil, err
	}

	return ledger, nil
}
