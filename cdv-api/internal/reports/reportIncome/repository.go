package reportincome

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetReport(ctx context.Context, filter ReportFilter) ([]IncomeMovement, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

const reportQuery = `
SELECT
    p.id,
    p.id_reservacion,
    p.cantidad,
    p.tipo_pago,
    p.method,
    p.fecha,
    r.id_cabana,
    c.nombre,
    u.nombre
FROM pagos p
JOIN reservaciones r ON r.id = p.id_reservacion
JOIN cabanas c ON c.id = r.id_cabana
JOIN usuarios u ON u.id = r.id_usuario
WHERE p.fecha >= $1
  AND p.fecha < $2
  AND ($3::int IS NULL OR c.id = $3)
  AND ($4::int IS NULL OR c.id_anfitrion = $4)
ORDER BY p.fecha ASC, p.id ASC
`

func (r *repository) GetReport(ctx context.Context, filter ReportFilter) ([]IncomeMovement, error) {
	endExclusive := filter.EndDate.AddDate(0, 0, 1)

	var hostID *int
	if filter.HostID != nil {
		value := int(*filter.HostID)
		hostID = &value
	}

	rows, err := r.pool.Query(ctx, reportQuery,
		filter.StartDate,
		endExclusive,
		filter.CabinID,
		hostID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying income report: %w", err)
	}
	defer rows.Close()

	movements := []IncomeMovement{}
	for rows.Next() {
		var (
			movement     IncomeMovement
			movementType string
			amount       float64
		)

		if err := rows.Scan(
			&movement.ID,
			&movement.ReservationID,
			&amount,
			&movementType,
			&movement.Method,
			&movement.Date,
			&movement.CabinID,
			&movement.CabinName,
			&movement.GuestName,
		); err != nil {
			return nil, fmt.Errorf("scanning income report row: %w", err)
		}

		movement.Type = movementType
		if movementType == refundType {
			movement.Amount = -amount
		} else {
			movement.Amount = amount
		}

		movements = append(movements, movement)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating income report rows: %w", err)
	}

	return movements, nil
}
