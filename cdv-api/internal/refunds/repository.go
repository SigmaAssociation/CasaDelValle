package refunds

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetSummary(ctx context.Context, reservationID int) (float64, float64, error) {
	query := `
		SELECT
			COALESCE(SUM(cantidad) FILTER (WHERE tipo_pago IN ('anticipo', 'pago')), 0)::float8,
			COALESCE(SUM(cantidad) FILTER (WHERE tipo_pago = 'reembolso'), 0)::float8
		FROM pagos
		WHERE id_reservacion = $1
	`

	var totalPaid, totalRefunded float64
	if err := r.pool.QueryRow(ctx, query, reservationID).Scan(&totalPaid, &totalRefunded); err != nil {
		return 0, 0, err
	}
	return totalPaid, totalRefunded, nil
}

func (r *Repository) ListMovements(ctx context.Context, reservationID int) ([]FinancialMovement, error) {
	query := `
		SELECT id, tipo_pago, cantidad, method, fecha
		FROM pagos
		WHERE id_reservacion = $1
		ORDER BY fecha ASC, id ASC
	`

	rows, err := r.pool.Query(ctx, query, reservationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movements := []FinancialMovement{}
	for rows.Next() {
		var m FinancialMovement
		if err := rows.Scan(&m.ID, &m.Subtype, &m.Amount, &m.Method, &m.CreatedAt); err != nil {
			return nil, err
		}
		if m.Subtype == RefundType {
			m.Type = RefundType
		} else {
			m.Type = PaymentTypePago
		}
		movements = append(movements, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return movements, nil
}

func (r *Repository) InsertRefund(ctx context.Context, reservationID int, hostID uint, amount float64) (*Refund, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var totalPaid, totalRefunded float64
	summary := `
		SELECT
			COALESCE(SUM(cantidad) FILTER (WHERE tipo_pago IN ('anticipo', 'pago')), 0)::float8,
			COALESCE(SUM(cantidad) FILTER (WHERE tipo_pago = 'reembolso'), 0)::float8
		FROM pagos
		WHERE id_reservacion = $1
	`
	if err := tx.QueryRow(ctx, summary, reservationID).Scan(&totalPaid, &totalRefunded); err != nil {
		return nil, err
	}

	if round2(totalRefunded+amount) > round2(totalPaid) {
		return nil, ErrRefundExceedsAvailable
	}

	insert := `
		INSERT INTO pagos (id_usuario, cantidad, tipo_pago, method, id_reservacion)
		VALUES ($1, $2, 'reembolso', $4, $3)
		RETURNING id, fecha
	`

	var refund Refund
	if err := tx.QueryRow(ctx, insert, hostID, amount, reservationID, RefundMethod).Scan(&refund.ID, &refund.CreatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return nil, ErrReservationNotFound
			case "23514":
				return nil, fmt.Errorf("no se pudo registrar el reembolso (restricción %s): %w", pgErr.ConstraintName, err)
			}
		}
		return nil, err
	}

	refund.ReservationID = reservationID
	refund.HostID = hostID
	refund.Amount = amount
	refund.Method = RefundMethod

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &refund, nil
}
