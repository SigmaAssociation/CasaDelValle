package images

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const imageColumns = `id, ruta, id_usuario, id_cabana`

func scanImage(scan func(dest ...any) error) (Image, error) {
	var img Image
	err := scan(&img.ID, &img.Ruta, &img.IDUsuario, &img.IDCabana)
	return img, err
}

func (r *Repository) CreateImage(ctx context.Context, img Image) (int, error) {
	var id int
	query := `
		INSERT INTO imagenes (ruta, id_usuario, id_cabana)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	err := r.pool.QueryRow(ctx, query, img.Ruta, img.IDUsuario, img.IDCabana).Scan(&id)
	return id, err
}

func (r *Repository) GetByCabinID(ctx context.Context, cabinID int) ([]Image, error) {
	rows, err := r.pool.Query(
		ctx,
		`SELECT `+imageColumns+` FROM imagenes WHERE id_cabana = $1 ORDER BY id`,
		cabinID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		img, err := scanImage(rows.Scan)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, rows.Err()
}

func (r *Repository) DeleteImage(ctx context.Context, id int) (int, error) {
	result, err := r.pool.Exec(ctx, `DELETE FROM imagenes WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	return int(result.RowsAffected()), nil
}

func (r *Repository) GetImageByID(ctx context.Context, id int) (Image, error) {
	row := r.pool.QueryRow(
		ctx,
		`SELECT `+imageColumns+` FROM imagenes WHERE id = $1`,
		id,
	)
	return scanImage(row.Scan)
}

func (r *Repository) GetCabinHostID(ctx context.Context, cabinID int) (int, error) {
	var hostID int
	err := r.pool.QueryRow(ctx, `SELECT id_anfitrion FROM cabanas WHERE id = $1`, cabinID).Scan(&hostID)
	return hostID, err
}