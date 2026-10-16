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

func (r *Repository) GetCabinName(ctx context.Context, cabinID int) (string, error) {
	var name string
	query := "SELECT nombre FROM cabanas WHERE id = $1"
	if err := r.pool.QueryRow(ctx, query, cabinID).Scan(&name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("Cabaña no encontrada")
		}
		return "", err
	}
	return name, nil
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
		WHERE id = $1 AND estado IN ('pendiente', 'confirmada')
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
		SELECT r.id, r.id_cabana, c.id_anfitrion, c.nombre,
		       img.ruta,
		       r.id_usuario, u.nombre,
		       to_char(r.fecha_inicio, 'YYYY-MM-DD'),
		       to_char(r.fecha_fin, 'YYYY-MM-DD'),
		       c.precio,
		       COALESCE((SELECT SUM(p.cantidad) FROM pagos p WHERE p.id_reservacion = r.id AND p.tipo_pago <> 'reembolso'), 0),
		       r.estado, r.fecha_creacion, r.fecha_cancelacion
		FROM reservaciones r
		JOIN cabanas c ON c.id = r.id_cabana
		JOIN usuarios u ON u.id = r.id_usuario
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
			&c.ID, &c.CabinID, &c.HostID, &c.CabinName, &c.CabinImageURL,
			&c.GuestID, &c.GuestName,
			&c.StartDate, &c.EndDate, &nightlyRate, &c.PaidTotal,
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

func (r *Repository) GetReservationsByCabinID(ctx context.Context, cabinID uint) ([]ReservationCard, error) {
	query := `
		SELECT r.id, r.id_cabana, c.id_anfitrion, c.nombre,
		       img.ruta,
		       r.id_usuario, u.nombre,
		       to_char(r.fecha_inicio, 'YYYY-MM-DD'),
		       to_char(r.fecha_fin, 'YYYY-MM-DD'),
		       c.precio,
		       COALESCE((SELECT SUM(p.cantidad) FROM pagos p WHERE p.id_reservacion = r.id AND p.tipo_pago <> 'reembolso'), 0),
		       r.estado, r.fecha_creacion, r.fecha_cancelacion
		FROM reservaciones r
		JOIN cabanas c ON c.id = r.id_cabana
		JOIN usuarios u ON u.id = r.id_usuario
		LEFT JOIN LATERAL (
			SELECT i.ruta
			FROM imagenes i
			WHERE i.id_cabana = c.id
			ORDER BY i.fecha_creacion DESC, i.id DESC
			LIMIT 1
		) img ON TRUE
		WHERE r.id_cabana = $1
		ORDER BY r.fecha_inicio DESC, r.id DESC
	`

	rows, err := r.pool.Query(ctx, query, cabinID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cards := []ReservationCard{}
	for rows.Next() {
		var c ReservationCard
		var nightlyRate float64
		if err := rows.Scan(
			&c.ID, &c.CabinID, &c.HostID, &c.CabinName, &c.CabinImageURL,
			&c.GuestID, &c.GuestName,
			&c.StartDate, &c.EndDate, &nightlyRate, &c.PaidTotal,
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

func (r *Repository) GetCheckinEligibility(ctx context.Context, id int) (*CheckinEligibility, error) {
	var e CheckinEligibility
	query := `
	SELECT r.id, r.id_usuario, r.id_cabana, c.id_anfitrion,
	       r.estado,
	       to_char(r.fecha_inicio, 'YYYY-MM-DD'),
	       to_char(r.fecha_fin, 'YYYY-MM-DD'),
	       r.fecha_checkin,
	       c.precio,
	       COALESCE((SELECT SUM(p.cantidad) FROM pagos p WHERE p.id_reservacion = r.id AND p.tipo_pago <> 'reembolso'), 0)
		FROM reservaciones r
		JOIN cabanas c ON c.id = r.id_cabana
		WHERE r.id = $1
	`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&e.ReservationID, &e.UserID, &e.CabinID, &e.HostID,
		&e.Status, &e.StartDate, &e.EndDate, &e.CheckedInAt,
		&e.NightlyRate, &e.PaidTotal,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReservationNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *Repository) RecordCheckin(ctx context.Context, id int, checkedInAt time.Time) error {
	query := `
		UPDATE reservaciones
		SET estado = 'en_curso', fecha_checkin = $2
		WHERE id = $1 AND estado = 'confirmada'
	`
	tag, err := r.pool.Exec(ctx, query, id, checkedInAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrReservationNotCheckinable
	}
	return nil
}

func (r *Repository) RecordCheckinTimestamp(ctx context.Context, id int, checkedInAt time.Time) error {
	query := `
		UPDATE reservaciones
		SET fecha_checkin = $2
		WHERE id = $1 AND estado = 'en_curso' AND fecha_checkin IS NULL
	`
	tag, err := r.pool.Exec(ctx, query, id, checkedInAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrReservationNotCheckinable
	}
	return nil
}

func (r *Repository) CheckinWithCashPayment(ctx context.Context, id int, payerID uint, cashAmount float64, checkedInAt time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	insert := `
		INSERT INTO pagos (id_usuario, cantidad, tipo_pago, method, id_reservacion)
		VALUES ($1, $2, 'pago', 'efectivo', $3)
	`
	if _, err := tx.Exec(ctx, insert, payerID, cashAmount, id); err != nil {
		return err
	}

	update := `
		UPDATE reservaciones
		SET estado = 'en_curso', fecha_checkin = COALESCE(fecha_checkin, $2)
		WHERE id = $1 AND (estado = 'confirmada'
			OR (estado = 'en_curso' AND fecha_checkin IS NULL))
	`
	tag, err := tx.Exec(ctx, update, id, checkedInAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrReservationNotCheckinable
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetCabinHostID(ctx context.Context, cabinID uint) (uint, error) {
	var hostID uint
	err := r.pool.QueryRow(ctx, `SELECT id_anfitrion FROM cabanas WHERE id = $1`, cabinID).Scan(&hostID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrCabinNotFound
		}
		return 0, err
	}
	return hostID, nil
}

func (r *Repository) HasUserOverlap(ctx context.Context, userID uint, startDate, endDate string, excludeID int) (bool, error) {
	query := `
		SELECT fecha_inicio, fecha_fin
		FROM reservaciones
		WHERE id_usuario = $1
		  AND id <> $2
	  AND estado IN ('pendiente', 'confirmada', 'en_curso')
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

func (r *Repository) UpdateReservation(ctx context.Context, id int, req UpdateReservationRequest) error {
	query := `
		UPDATE reservaciones
		SET id_cabana = $1, fecha_inicio = $2, fecha_fin = $3
		WHERE id = $4 AND estado IN ('pendiente', 'confirmada')
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

func (r *Repository) AdvanceReservationStates(ctx context.Context) ([]int, error) {
	startQuery := `
		UPDATE reservaciones
		SET estado = 'en_curso'
		WHERE estado = 'confirmada'
		  AND fecha_inicio <= CURRENT_DATE
		  AND fecha_fin >= CURRENT_DATE
	`
	if _, err := r.pool.Exec(ctx, startQuery); err != nil {
		return nil, err
	}

	finishQuery := `
		UPDATE reservaciones
		SET estado = 'finalizada',
		    fecha_checkin = COALESCE(fecha_checkin, fecha_inicio::timestamp AT TIME ZONE 'America/Guatemala'),
		    fecha_checkout = COALESCE(fecha_checkout, fecha_fin::timestamp AT TIME ZONE 'America/Guatemala')
		WHERE estado IN ('pendiente', 'confirmada', 'en_curso')
		  AND fecha_fin < CURRENT_DATE
		RETURNING id
	`

	rows, err := r.pool.Query(ctx, finishQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := []int{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *Repository) GetCabinNightlyRate(ctx context.Context, cabinID uint) (float64, error) {
	var rate float64
	err := r.pool.QueryRow(ctx, `SELECT precio FROM cabanas WHERE id = $1`, cabinID).Scan(&rate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrCabinNotFound
		}
		return 0, err
	}
	return rate, nil
}

func (r *Repository) PayAnticipo(ctx context.Context, reservationID int, payerID uint, amount float64, method string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	insert := `
		INSERT INTO pagos (id_usuario, cantidad, tipo_pago, method, id_reservacion)
		VALUES ($1, $2, 'anticipo', $3, $4)
	`
	if _, err := tx.Exec(ctx, insert, payerID, amount, method, reservationID); err != nil {
		return err
	}

	update := `
		UPDATE reservaciones
		SET estado = 'confirmada'
		WHERE id = $1 AND estado = 'pendiente'
	`
	tag, err := tx.Exec(ctx, update, reservationID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrReservationNotPending
	}

	return tx.Commit(ctx)
}
