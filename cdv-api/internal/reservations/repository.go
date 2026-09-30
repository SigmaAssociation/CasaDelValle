package reservations

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) CreateReservation(ctx context.Context, req ReservationRequest) (int, error) {
	var id int
	query := "INSERT INTO reservaciones (id_cabana, id_usuario, fecha_inicio, fecha_fin) VALUES ($1, $2, $3, $4) RETURNING id"
	err := r.pool.QueryRow(ctx, query, req.CabinID, req.UserID, req.StartDate, req.EndDate).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				switch pgErr.ConstraintName {
				case "fk_reservacion_cabana":
					return 0, errors.New("La cabaña especificada no existe")
				case "fk_reservacion_usuario":
					return 0, errors.New("El usuario especificado no existe")
				default:
					return 0, errors.New("La cabaña o el usuario especificado no existe")
				}
			case "23514":
				return 0, errors.New("La fecha de fin no puede ser anterior a la fecha de inicio")
			}
		}
		return 0, err
	}
	return id, nil
}

func (r *Repository) HasOverlap(ctx context.Context, cabinID uint, startDate, endDate string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM reservaciones
			WHERE id_cabana = $1
			  AND fecha_inicio <= $3
			  AND fecha_fin >= $2
		)
	`
	err := r.pool.QueryRow(ctx, query, cabinID, startDate, endDate).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
