package images

import (
	"context"

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

const imageColumns = `
	id, ruta, id_usuario, id_cabana, fecha_creacion
`

// Orden de las imágenes: de la más reciente a la más antigua.
const imageOrder = ` ORDER BY fecha_creacion DESC, id DESC `

func scanImage(scan func(dest ...any) error) (Image, error) {
	var i Image

	err := scan(
		&i.ID,
		&i.Path,
		&i.UserID,
		&i.CabinID,
		&i.CreatedAt,
	)

	return i, err
}

// GetByCabinID retorna todas las imágenes que pertenecen a una cabaña.
func (r *Repository) GetByCabinID(ctx context.Context, cabinID int) ([]Image, error) {
	rows, err := r.pool.Query(
		ctx,
		`SELECT `+imageColumns+` FROM imagenes WHERE id_cabana = $1`+imageOrder,
		cabinID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	images := []Image{}

	for rows.Next() {
		i, err := scanImage(rows.Scan)
		if err != nil {
			return nil, err
		}

		images = append(images, i)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return images, nil
}

// GetMainByCabinID retorna la imagen más reciente de una cabaña.
func (r *Repository) GetMainByCabinID(ctx context.Context, cabinID int) (Image, error) {
	row := r.pool.QueryRow(
		ctx,
		`SELECT `+imageColumns+` FROM imagenes WHERE id_cabana = $1`+imageOrder+` LIMIT 1`,
		cabinID,
	)

	return scanImage(row.Scan)
}

// GetByID retorna una imagen por su identificador.
func (r *Repository) GetByID(ctx context.Context, id int) (Image, error) {
	row := r.pool.QueryRow(
		ctx,
		`SELECT `+imageColumns+` FROM imagenes WHERE id = $1`,
		id,
	)

	return scanImage(row.Scan)
}
