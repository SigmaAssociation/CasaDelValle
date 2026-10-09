package notifications

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
	return &Repository{pool: pool}
}

const notificationColumns = `
	id, id_usuario, tipo, mensaje,
	COALESCE(entidad_tipo, ''), entidad_id,
	leida, fecha_creacion
`

func (r *Repository) Create(ctx context.Context, input NotificationInput) (int, error) {
	var id int
	query := `
		INSERT INTO notificaciones (id_usuario, tipo, mensaje, entidad_tipo, entidad_id)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5)
		RETURNING id
	`
	err := r.pool.QueryRow(
		ctx,
		query,
		input.UserID, input.Tipo, input.Mensaje, input.EntidadTipo, input.EntidadID,
	).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return 0, errors.New("El usuario especificado no existe")
			case "23514":
				return 0, errors.New("Datos de notificación inválidos")
			}
		}
		return 0, err
	}
	return id, nil
}

// GetByUser retorna una página del buzón del usuario, siempre en orden
// cronológico descendente.
func (r *Repository) GetByUser(ctx context.Context, userID uint, limit, offset int) ([]Notification, error) {
	query := `
		SELECT ` + notificationColumns + `
		FROM notificaciones
		WHERE id_usuario = $1
		ORDER BY fecha_creacion DESC, id DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Notification{}
	for rows.Next() {
		var n Notification
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.Tipo, &n.Mensaje,
			&n.EntidadTipo, &n.EntidadID,
			&n.Leida, &n.FechaCreacion,
		); err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// CountByUser devuelve el total de notificaciones del buzón del usuario.
func (r *Repository) CountByUser(ctx context.Context, userID uint) (int, error) {
	query := "SELECT COUNT(*) FROM notificaciones WHERE id_usuario = $1"

	var count int
	if err := r.pool.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) GetUnreadCount(ctx context.Context, userID uint) (int, error) {
	query := "SELECT COUNT(*) FROM notificaciones WHERE id_usuario = $1 AND leida = FALSE"

	var count int
	if err := r.pool.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) GetAdminUserIDs(ctx context.Context) ([]uint, error) {
	query := `
		SELECT u.id
		FROM usuarios u
		JOIN roles r ON r.id = u.id_rol
		WHERE r.tipo = 'Administrador'
		ORDER BY u.id
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := []uint{}
	for rows.Next() {
		var id uint
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

func (r *Repository) MarkAllRead(ctx context.Context, userID uint) (int, error) {
	query := "UPDATE notificaciones SET leida = TRUE WHERE id_usuario = $1 AND leida = FALSE"

	tag, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (r *Repository) MarkRead(ctx context.Context, id int, userID uint) error {
	query := "UPDATE notificaciones SET leida = TRUE WHERE id = $1 AND id_usuario = $2"

	tag, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotificationNotFound
	}
	return nil
}
