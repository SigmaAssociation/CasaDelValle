package reportreservations

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetReport(ctx context.Context, filter ReportFilter) ([]CabinReport, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

const reportQuery = `
SELECT
    c.id,
    c.nombre,
    c.direccion,
    c.precio,
    c.capacidad,
    a.id,
    a.nombre,
    COUNT(r.id) OVER (PARTITION BY c.id) AS total_reservaciones,
    r.id,
    r.fecha_inicio,
    r.fecha_fin,
    r.estado,
    r.fecha_creacion,
    h.id,
    h.nombre
FROM cabanas c
JOIN usuarios a ON a.id = c.id_anfitrion
JOIN reservaciones r ON r.id_cabana = c.id
JOIN usuarios h ON h.id = r.id_usuario
WHERE ($1::int IS NULL OR c.id_anfitrion = $1)
  AND ($2::int IS NULL OR c.id = $2)
  AND ($3::int IS NULL OR r.id_usuario = $3)
  AND ($4::text = '' OR r.estado = $4)
  AND ($5::timestamptz IS NULL OR r.fecha_creacion >= $5)
  AND ($6::timestamptz IS NULL OR r.fecha_creacion < $6)
ORDER BY total_reservaciones DESC, c.nombre ASC, c.id ASC, r.fecha_inicio DESC, r.id DESC
`

func (r *repository) GetReport(ctx context.Context, filter ReportFilter) ([]CabinReport, error) {

	var createdToExclusive *time.Time
	if filter.CreatedTo != nil {
		t := filter.CreatedTo.AddDate(0, 0, 1)
		createdToExclusive = &t
	}

	rows, err := r.pool.Query(ctx, reportQuery,
		filter.HostID,
		filter.CabinID,
		filter.GuestID,
		filter.Status,
		filter.CreatedFrom,
		createdToExclusive,
	)
	if err != nil {
		return nil, fmt.Errorf("querying reservations report: %w", err)
	}
	defer rows.Close()

	cabins := []CabinReport{}
	var currentCabinID int

	for rows.Next() {
		var (
			cabin       CabinReport
			reservation ReservationSummary
		)

		err := rows.Scan(
			&cabin.ID,
			&cabin.Name,
			&cabin.Address,
			&cabin.Price,
			&cabin.Capacity,
			&cabin.Host.ID,
			&cabin.Host.Name,
			&cabin.TotalReservations,
			&reservation.ID,
			&reservation.StartDate,
			&reservation.EndDate,
			&reservation.Status,
			&reservation.CreatedAt,
			&reservation.Guest.ID,
			&reservation.Guest.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning report row: %w", err)
		}

		if len(cabins) == 0 || cabin.ID != currentCabinID {
			cabin.Reservations = []ReservationSummary{}
			cabins = append(cabins, cabin)
			currentCabinID = cabin.ID
		}

		last := &cabins[len(cabins)-1]
		last.Reservations = append(last.Reservations, reservation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating report rows: %w", err)
	}

	return cabins, nil
}