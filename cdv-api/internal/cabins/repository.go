package cabins

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

func (r *Repository) CreateCabin(ctx context.Context, req CreateCabinRequest) (int, error) {
	var id int

	query := `
		INSERT INTO cabanas (direccion, precio, descripcion, capacidad, reglas, id_anfitrion, id_comision)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		req.Direccion,
		req.Precio,
		req.Descripcion,
		req.Capacidad,
		req.Reglas,
		req.IDAnfitrion,
		req.IDComision,
	).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return 0, errors.New("El anfitrión o la comisión especificada no existe")
		}
		return 0, err
	}

	return id, nil
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

func (r *Repository) GetHostID(ctx context.Context, cabinID int) (int, error) {
	var hostID int
	err := r.pool.QueryRow(ctx, `SELECT id_anfitrion FROM cabanas WHERE id = $1`, cabinID).Scan(&hostID)
	return hostID, err
}

func (r *Repository) Update(ctx context.Context, req UpdateCabinRequest) error {
	query := `
		UPDATE cabanas 
		SET direccion = $1, precio = $2, descripcion = $3, capacidad = $4, reglas = $5
		WHERE id = $6
	`
	cmdTag, err := r.pool.Exec(ctx, query,
		req.Address,
		req.Price,
		req.Description,
		req.Capacity,
		req.Rules,
		req.ID,
	)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("no se encontró la cabaña para actualizar")
	}
	return nil
}
