package cabins

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

const cabinColumns = `
	id, direccion, precio, COALESCE(descripcion, ''), capacidad,
	COALESCE(reglas, ''), id_anfitrion, id_comision
`

func scanCabin(scan func(dest ...any) error) (Cabin, error) {
	var c Cabin

	err := scan(
		&c.ID,
		&c.Address,
		&c.Price,
		&c.Description,
		&c.Capacity,
		&c.Rules,
		&c.HostID,
		&c.CommissionID,
	)

	return c, err
}

func (r *Repository) GetAll(ctx context.Context) ([]Cabin, error) {
	rows, err := r.pool.Query(
		ctx,
		`SELECT `+cabinColumns+` FROM cabanas ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cabins := []Cabin{}

	for rows.Next() {
		c, err := scanCabin(rows.Scan)
		if err != nil {
			return nil, err
		}

		cabins = append(cabins, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cabins, nil
}

func (r *Repository) GetByID(ctx context.Context, id int) (Cabin, error) {
	row := r.pool.QueryRow(
		ctx,
		`SELECT `+cabinColumns+` FROM cabanas WHERE id = $1`,
		id,
	)

	return scanCabin(row.Scan)
}

func (r *Repository) GetByHostID(ctx context.Context, hostID int) ([]Cabin, error) {
	rows, err := r.pool.Query(
		ctx,
		`SELECT `+cabinColumns+` FROM cabanas WHERE id_anfitrion = $1 ORDER BY id`,
		hostID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cabins := []Cabin{}

	for rows.Next() {
		c, err := scanCabin(rows.Scan)
		if err != nil {
			return nil, err
		}

		cabins = append(cabins, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cabins, nil
}
