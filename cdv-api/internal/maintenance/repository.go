package maintenance

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

type querier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func ownedCabin(ctx context.Context, q querier, id int, actor uint) (string, error) {
	var owner uint
	var name string
	err := q.QueryRow(ctx, `SELECT id_anfitrion,nombre FROM cabanas WHERE id=$1`, id).Scan(&owner, &name)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if owner != actor {
		return "", ErrForbidden
	}
	return name, nil
}

const columns = `id,id_cabana,to_char(fecha_inicio,'YYYY-MM-DD'),to_char(fecha_fin,'YYYY-MM-DD'),motivo,fecha_creacion,fecha_cancelacion`

func scanMaintenance(row pgx.Row) (*Maintenance, error) {
	var m Maintenance
	err := row.Scan(&m.ID, &m.CabinID, &m.StartDate, &m.EndDate, &m.Reason, &m.CreatedAt, &m.CancelledAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &m, err
}
func (r *Repository) List(ctx context.Context, id int, actor uint) ([]Maintenance, error) {
	if _, err := ownedCabin(ctx, r.pool, id, actor); err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `SELECT `+columns+` FROM mantenimientos WHERE id_cabana=$1 ORDER BY fecha_inicio DESC,id DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Maintenance{}
	for rows.Next() {
		m, err := scanMaintenance(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *m)
	}
	return result, rows.Err()
}
func affected(ctx context.Context, q querier, id int, req DateRequest, lock bool) ([]AffectedReservation, error) {
	sql := `SELECT r.id,r.id_usuario,u.nombre,to_char(r.fecha_inicio,'YYYY-MM-DD'),to_char(r.fecha_fin,'YYYY-MM-DD'),r.estado,r.precio_total
        FROM reservaciones r JOIN usuarios u ON u.id=r.id_usuario
        WHERE r.id_cabana=$1 AND r.estado IN ('pendiente','confirmada','en_curso')
        AND r.fecha_inicio<$3::date AND r.fecha_fin>$2::date ORDER BY r.id`
	if lock {
		sql += ` FOR UPDATE OF r`
	}
	rows, err := q.Query(ctx, sql, id, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []AffectedReservation{}
	for rows.Next() {
		var a AffectedReservation
		if err := rows.Scan(&a.ID, &a.GuestID, &a.GuestName, &a.StartDate, &a.EndDate, &a.Status, &a.TotalPrice); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}
func (r *Repository) Preview(ctx context.Context, id int, actor uint, req DateRequest) (*Preview, error) {
	if _, err := ownedCabin(ctx, r.pool, id, actor); err != nil {
		return nil, err
	}
	list, err := affected(ctx, r.pool, id, req, false)
	return &Preview{Reservations: list}, err
}
func (r *Repository) Create(ctx context.Context, id int, actor uint, req CreateRequest) (*Maintenance, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(947,1)`); err != nil {
		return nil, err
	}
	name, err := ownedCabin(ctx, tx, id, actor)
	if err != nil {
		return nil, err
	}
	var overlap bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM mantenimientos WHERE id_cabana=$1 AND fecha_cancelacion IS NULL
        AND fecha_inicio<$3::date AND fecha_fin>$2::date)`, id, req.StartDate, req.EndDate).Scan(&overlap)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, ErrConflict
	}
	list, err := affected(ctx, tx, id, DateRequest{req.StartDate, req.EndDate}, true)
	if err != nil {
		return nil, err
	}
	m, err := scanMaintenance(tx.QueryRow(ctx, `INSERT INTO mantenimientos(id_cabana,fecha_inicio,fecha_fin,motivo) VALUES($1,$2,$3,$4) RETURNING `+columns, id, req.StartDate, req.EndDate, req.Reason))
	if err != nil {
		return nil, err
	}
	for _, a := range list {
		if _, err = tx.Exec(ctx, `INSERT INTO mantenimiento_reservaciones(id_mantenimiento,id_reservacion,accion,estado_anterior)
            VALUES($1,$2,'pendiente',$3)`, m.ID, a.ID, a.Status); err != nil {
			return nil, err
		}
		if _, err = tx.Exec(ctx, `UPDATE reservaciones SET estado='por_mantenimiento' WHERE id=$1`, a.ID); err != nil {
			return nil, err
		}
		message := fmt.Sprintf("Tu reserva #%d en %s fue afectada por mantenimiento del %s al %s. Desde Mis reservaciones puedes elegir nuevas fechas en la misma cabaña, conservando las noches y el total de Q%.2f, o solicitar la devolución completa de lo pagado.", a.ID, name, req.StartDate, req.EndDate, a.TotalPrice)
		if err = notify(ctx, tx, uint(a.GuestID), a.ID, "reserva_mantenimiento", message); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return m, nil
}
func notify(ctx context.Context, tx pgx.Tx, user uint, reservation int, tipo, message string) error {
	_, err := tx.Exec(ctx, `INSERT INTO notificaciones(id_usuario,tipo,mensaje,entidad_tipo,entidad_id) VALUES($1,$2,$3,'reservacion',$4)`, user, tipo, message, reservation)
	return err
}
func pending(ctx context.Context, q querier, id int, actor uint, lock bool) (*PendingDecision, error) {
	sql := `SELECT r.id,r.id_usuario,c.id_anfitrion,r.id_cabana,m.motivo,
        to_char(r.fecha_inicio,'YYYY-MM-DD'),to_char(r.fecha_fin,'YYYY-MM-DD'),r.fecha_fin-r.fecha_inicio,
        r.precio_total,mr.id_mantenimiento,mr.estado_anterior
        FROM reservaciones r JOIN cabanas c ON c.id=r.id_cabana
        JOIN mantenimiento_reservaciones mr ON mr.id_reservacion=r.id AND mr.accion='pendiente'
        JOIN mantenimientos m ON m.id=mr.id_mantenimiento WHERE r.id=$1 AND r.estado='por_mantenimiento'`
	if lock {
		sql += ` FOR UPDATE OF r,mr`
	}
	var d PendingDecision
	err := q.QueryRow(ctx, sql, id).Scan(&d.ReservationID, &d.GuestID, &d.HostID, &d.CabinID, &d.Reason, &d.StartDate, &d.EndDate, &d.Nights, &d.TotalPrice, &d.MaintenanceID, &d.PreviousStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if d.GuestID != actor {
		return nil, ErrForbidden
	}
	err = q.QueryRow(ctx, `SELECT GREATEST(0,COALESCE((SELECT SUM(CASE WHEN tipo_pago='reembolso' THEN -cantidad ELSE cantidad END)
        FROM pagos WHERE id_reservacion=$1),0)-COALESCE((SELECT SUM(cantidad) FROM reembolsos WHERE id_reservacion=$1),0))`, id).Scan(&d.Refundable)
	return &d, err
}
func (r *Repository) Pending(ctx context.Context, id int, actor uint) (*PendingDecision, error) {
	d, err := pending(ctx, r.pool, id, actor, false)
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `SELECT to_char(fecha_inicio,'YYYY-MM-DD'),to_char(fecha_fin,'YYYY-MM-DD') FROM reservaciones
        WHERE id_cabana=$1 AND id<>$2 AND estado IN ('pendiente','confirmada','en_curso','por_mantenimiento')
        UNION ALL SELECT to_char(fecha_inicio,'YYYY-MM-DD'),to_char(fecha_fin,'YYYY-MM-DD') FROM mantenimientos
        WHERE id_cabana=$1 AND fecha_cancelacion IS NULL ORDER BY 1`, d.CabinID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	d.BlockedRanges = []DateRange{}
	for rows.Next() {
		var b DateRange
		if err := rows.Scan(&b.Start, &b.End); err != nil {
			return nil, err
		}
		d.BlockedRanges = append(d.BlockedRanges, b)
	}
	return d, rows.Err()
}
func (r *Repository) Resolve(ctx context.Context, id int, actor uint, req DecisionRequest) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(947,1)`); err != nil {
		return err
	}
	d, err := pending(ctx, tx, id, actor, true)
	if err != nil {
		return err
	}
	var message, tipo string
	refunded := 0.0
	switch req.Action {
	case "reprogramar":
		var nights int
		var unavailable bool
		err = tx.QueryRow(ctx, `SELECT $2::date-$1::date,
            EXISTS(SELECT 1 FROM mantenimientos WHERE id_cabana=$3 AND fecha_cancelacion IS NULL AND fecha_inicio<$2::date AND fecha_fin>$1::date)
            OR EXISTS(SELECT 1 FROM reservaciones WHERE id<>$4 AND estado IN ('pendiente','confirmada','en_curso','por_mantenimiento')
                AND (id_cabana=$3 OR id_usuario=$5) AND fecha_inicio<$2::date AND fecha_fin>$1::date)`, req.StartDate, req.EndDate, d.CabinID, id, actor).Scan(&nights, &unavailable)
		if err != nil {
			return err
		}
		if nights != d.Nights {
			return validationError("El cambio por mantenimiento debe conservar la cantidad de noches original")
		}
		if unavailable {
			return ErrUnavailable
		}
		status := "confirmada"
		if d.PreviousStatus == "pendiente" {
			status = "pendiente"
		}
		_, err = tx.Exec(ctx, `UPDATE reservaciones SET fecha_inicio=$2,fecha_fin=$3,estado=$4,fecha_checkin=NULL,fecha_checkout=NULL WHERE id=$1`, id, req.StartDate, req.EndDate, status)
		tipo = "reserva_actualizada"
		message = fmt.Sprintf("La reserva #%d fue reprogramada por el huésped del %s al %s en la misma cabaña. Conserva el total original de Q%.2f y los pagos registrados.", id, req.StartDate, req.EndDate, d.TotalPrice)
	case "reembolsar":
		refunded = d.Refundable
		if refunded > 0 {
			_, err = tx.Exec(ctx, `INSERT INTO pagos(id_usuario,cantidad,tipo_pago,method,id_reservacion) VALUES($1,$2,'reembolso','transferencia',$3)`, d.HostID, refunded, id)
			if err != nil {
				return err
			}
		}
		_, err = tx.Exec(ctx, `UPDATE reservaciones SET estado='cancelada',fecha_cancelacion=CURRENT_TIMESTAMP WHERE id=$1`, id)
		tipo = "reserva_cancelada"
		message = fmt.Sprintf("El huésped solicitó cancelar la reserva #%d por mantenimiento. Se registró la devolución completa del dinero pagado pendiente de devolución: Q%.2f, sin penalizaciones.", id, refunded)
	default:
		return ErrConflict
	}
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE mantenimiento_reservaciones SET accion=$2,monto_reembolsado=$3,fecha_resolucion=CURRENT_TIMESTAMP WHERE id_reservacion=$1 AND accion='pendiente'`, id, req.Action, refunded); err != nil {
		return err
	}
	if err = notify(ctx, tx, d.GuestID, id, tipo, message); err != nil {
		return err
	}
	if d.HostID != d.GuestID {
		if err = notify(ctx, tx, d.HostID, id, tipo, message); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
func (r *Repository) Cancel(ctx context.Context, cabinID, id int, actor uint) (*Maintenance, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(947,1)`); err != nil {
		return nil, err
	}
	if _, err = ownedCabin(ctx, tx, cabinID, actor); err != nil {
		return nil, err
	}
	m, err := scanMaintenance(tx.QueryRow(ctx, `UPDATE mantenimientos SET fecha_cancelacion=COALESCE(fecha_cancelacion,CURRENT_TIMESTAMP)
        WHERE id=$1 AND id_cabana=$2 RETURNING `+columns, id, cabinID))
	if err != nil {
		return nil, err
	}
	// Restituye las reservas originales aún utilizables. Si las fechas ya
	// pasaron o hay otro mantenimiento, conserva la decisión del huésped.
	rows, err := tx.Query(ctx, `SELECT r.id,r.id_usuario,mr.estado_anterior FROM reservaciones r
        JOIN mantenimiento_reservaciones mr ON mr.id_reservacion=r.id AND mr.accion='pendiente'
        WHERE r.id_cabana=$1 AND r.estado='por_mantenimiento'
        AND r.fecha_inicio >= (now() AT TIME ZONE 'America/Guatemala')::date
        AND NOT EXISTS(SELECT 1 FROM mantenimientos m WHERE m.id_cabana=$1 AND m.fecha_cancelacion IS NULL
            AND m.fecha_inicio<r.fecha_fin AND m.fecha_fin>r.fecha_inicio) FOR UPDATE OF r,mr`, cabinID)
	if err != nil {
		return nil, err
	}
	type restore struct {
		id     int
		guest  uint
		status string
	}
	list := []restore{}
	for rows.Next() {
		var a restore
		if err := rows.Scan(&a.id, &a.guest, &a.status); err != nil {
			rows.Close()
			return nil, err
		}
		list = append(list, a)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, a := range list {
		if _, err = tx.Exec(ctx, `UPDATE reservaciones SET estado=$2 WHERE id=$1`, a.id, a.status); err != nil {
			return nil, err
		}
		if _, err = tx.Exec(ctx, `UPDATE mantenimiento_reservaciones SET accion='sin_efecto',fecha_resolucion=CURRENT_TIMESTAMP WHERE id_reservacion=$1 AND accion='pendiente'`, a.id); err != nil {
			return nil, err
		}
		if err = notify(ctx, tx, a.guest, a.id, "reserva_actualizada", "El mantenimiento fue cancelado. Tu reserva conserva sus fechas, su estado y su total original."); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return m, nil
}
