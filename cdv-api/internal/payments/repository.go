package payments

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetReservationInfo(ctx context.Context, reservationID int) (*ReservationPaymentInfo, error) {
	query := `
		SELECT r.id, r.id_usuario, c.id_anfitrion, r.id_cabana, c.nombre,
		       r.estado, r.fecha_inicio::text, r.fecha_fin::text, r.precio_total/(r.fecha_fin-r.fecha_inicio),r.precio_total
		FROM reservaciones r
		JOIN cabanas c ON c.id = r.id_cabana
		WHERE r.id = $1
	`

	var info ReservationPaymentInfo
	err := r.pool.QueryRow(ctx, query, reservationID).Scan(
		&info.ReservationID, &info.UserID, &info.HostID, &info.CabinID, &info.CabinName,
		&info.Status, &info.StartDate, &info.EndDate, &info.NightlyRate, &info.TotalPrice,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReservationNotFound
		}
		return nil, err
	}
	return &info, nil
}

func (r *Repository) GetPaidAmount(ctx context.Context, reservationID int) (float64, error) {
	var paid float64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(cantidad), 0) FROM pagos WHERE id_reservacion = $1 AND tipo_pago <> 'reembolso'`,
		reservationID,
	).Scan(&paid)
	if err != nil {
		return 0, err
	}
	return paid, nil
}

func (r *Repository) ListByReservation(ctx context.Context, reservationID int) ([]Payment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, id_reservacion, id_usuario, cantidad, tipo_pago, method, fecha
		FROM pagos
		WHERE id_reservacion = $1 AND tipo_pago <> 'reembolso'
		ORDER BY fecha ASC, id ASC
	`, reservationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payments := []Payment{}
	for rows.Next() {
		var p Payment
		var tipoPago, methodCode string
		if err := rows.Scan(&p.ID, &p.ReservationID, &p.UserID, &p.Amount, &tipoPago, &methodCode, &p.Date); err != nil {
			return nil, err
		}
		p.Method = apiMethod(tipoPago, methodCode)
		payments = append(payments, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *Repository) Create(ctx context.Context, reservationID int, payerID uint, amount float64, method string) (*Payment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(947,1)`); err != nil {
		return nil, err
	}
	var status string
	var remaining float64
	if err := tx.QueryRow(ctx, `SELECT estado,precio_total-COALESCE((SELECT SUM(cantidad) FROM pagos WHERE id_reservacion=$1 AND tipo_pago<>'reembolso'),0)
        FROM reservaciones WHERE id=$1 FOR UPDATE`, reservationID).Scan(&status, &remaining); err != nil {
		return nil, err
	}
	if status != "confirmada" && status != "en_curso" {
		return nil, ErrReservationNotPayable
	}
	if amount > remaining {
		return nil, ErrAmountExceedsBalance
	}
	payment := &Payment{
		ReservationID: reservationID,
		UserID:        payerID,
		Amount:        amount,
		Method:        method,
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO pagos (id_usuario, cantidad, tipo_pago, method, id_reservacion)
		VALUES ($1, $2, 'pago', $3, $4)
		RETURNING id, fecha
	`, payerID, amount, method, reservationID).Scan(&payment.ID, &payment.Date)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return payment, nil
}
