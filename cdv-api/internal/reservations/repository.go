package reservations

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
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

func datesOverlap(existingStart, existingEnd, newStart, newEnd time.Time) bool {
	return existingStart.Before(newEnd) && existingEnd.After(newStart)
}

// HasOverlap indica si ya existe otra reservación activa para la cabaña
// cuyo rango de fechas se cruza con [startDate, endDate].
// excludeID permite ignorar una reservación (útil al editar).
func (r *Repository) HasOverlap(ctx context.Context, cabinID uint, startDate, endDate string, excludeID int) (bool, error) {
	query := `
		SELECT fecha_inicio, fecha_fin
		FROM reservaciones
		WHERE id_cabana = $1
		  AND id <> $2
		  AND estado <> 'cancelada'
	`

	rows, err := r.pool.Query(ctx, query, cabinID, excludeID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	newStart, err := parseReservationDate(startDate, "fecha de inicio")
	if err != nil {
		return false, err
	}
	newEnd, err := parseReservationDate(endDate, "fecha de fin")
	if err != nil {
		return false, err
	}

	for rows.Next() {
		var existingStart, existingEnd time.Time
		if err := rows.Scan(&existingStart, &existingEnd); err != nil {
			return false, err
		}
		if datesOverlap(existingStart, existingEnd, newStart, newEnd) {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}

	return false, nil
}

func (r *Repository) GetReservationByID(ctx context.Context, id int) (*ReservationDetail, error) {
	query := `
		SELECT r.id, r.id_usuario, r.id_cabana, c.id_anfitrion,
		       to_char(r.fecha_inicio, 'YYYY-MM-DD'),
		       to_char(r.fecha_fin, 'YYYY-MM-DD'),
		       r.estado, r.fecha_cancelacion
		FROM reservaciones r
		JOIN cabanas c ON c.id = r.id_cabana
		WHERE r.id = $1
	`

	var res ReservationDetail
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&res.ID, &res.UserID, &res.CabinID, &res.HostID,
		&res.StartDate, &res.EndDate, &res.Status, &res.CancelledAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReservationNotFound
		}
		return nil, err
	}
	return &res, nil
}

func (r *Repository) CancelReservation(ctx context.Context, id int, cancelledAt time.Time) error {
	query := `
		UPDATE reservaciones
		SET estado = 'cancelada', fecha_cancelacion = $2
		WHERE id = $1 AND estado = 'activa'
	`
	tag, err := r.pool.Exec(ctx, query, id, cancelledAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrReservationNotCancellable
	}
	return nil
}

func (r *Repository) GetReservationsByUserID(ctx context.Context, userID uint) ([]ReservationCard, error) {
	query := `
		SELECT r.id, r.id_cabana, c.nombre,
		       img.ruta,
		       to_char(r.fecha_inicio, 'YYYY-MM-DD'),
		       to_char(r.fecha_fin, 'YYYY-MM-DD'),
		       c.precio,
		       r.estado, r.fecha_creacion, r.fecha_cancelacion
		FROM reservaciones r
		JOIN cabanas c ON c.id = r.id_cabana
		LEFT JOIN LATERAL (
			SELECT i.ruta
			FROM imagenes i
			WHERE i.id_cabana = c.id
			ORDER BY i.fecha_creacion DESC, i.id DESC
			LIMIT 1
		) img ON TRUE
		WHERE r.id_usuario = $1
		ORDER BY r.fecha_inicio DESC, r.id DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cards := []ReservationCard{}
	for rows.Next() {
		var c ReservationCard
		var nightlyRate float64
		if err := rows.Scan(
			&c.ID, &c.CabinID, &c.CabinName, &c.CabinImageURL,
			&c.StartDate, &c.EndDate, &nightlyRate,
			&c.Status, &c.CreatedAt, &c.CancelledAt,
		); err != nil {
			return nil, err
		}

		totalPrice, err := calculateReservationTotalPrice(c.StartDate, c.EndDate, nightlyRate)
		if err != nil {
			return nil, err
		}
		c.TotalPrice = totalPrice
		cards = append(cards, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cards, nil
}

// HasUserOverlap indica si el usuario ya tiene otra reservación activa
// (en cualquier cabaña) que se cruza con [startDate, endDate].
// excludeID permite ignorar una reservación (útil al editar).
func (r *Repository) HasUserOverlap(ctx context.Context, userID uint, startDate, endDate string, excludeID int) (bool, error) {
	query := `
		SELECT fecha_inicio, fecha_fin
		FROM reservaciones
		WHERE id_usuario = $1
		  AND id <> $2
		  AND estado = 'activa'
	`

	rows, err := r.pool.Query(ctx, query, userID, excludeID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	newStart, err := parseReservationDate(startDate, "fecha de inicio")
	if err != nil {
		return false, err
	}
	newEnd, err := parseReservationDate(endDate, "fecha de fin")
	if err != nil {
		return false, err
	}

	for rows.Next() {
		var existingStart, existingEnd time.Time
		if err := rows.Scan(&existingStart, &existingEnd); err != nil {
			return false, err
		}
		if datesOverlap(existingStart, existingEnd, newStart, newEnd) {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}

	return false, nil
}

// UpdateReservation modifica cabaña y fechas de una reservación activa.
func (r *Repository) UpdateReservation(ctx context.Context, id int, req UpdateReservationRequest) error {
	query := `
		UPDATE reservaciones
		SET id_cabana = $1, fecha_inicio = $2, fecha_fin = $3
		WHERE id = $4 AND estado = 'activa'
	`
	tag, err := r.pool.Exec(ctx, query, req.CabinID, req.StartDate, req.EndDate, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				if pgErr.ConstraintName == "fk_reservacion_cabana" {
					return errors.New("La cabaña especificada no existe")
				}
				return errors.New("La cabaña o el usuario especificado no existe")
			case "23514":
				return errors.New("La fecha de fin no puede ser anterior a la fecha de inicio")
			}
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrReservationNotEditable
	}
	return nil
}
