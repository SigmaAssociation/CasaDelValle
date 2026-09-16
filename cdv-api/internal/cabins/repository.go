package cabins

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	id, nombre, direccion, precio, COALESCE(descripcion, ''), capacidad,
	COALESCE(reglas, ''), id_anfitrion, id_comision
`

const cabinSearchColumns = `
	c.id, c.nombre, c.direccion, c.precio, c.capacidad,
	COALESCE(u.nombre, '')
`

const cabinSearchFrom = `
	FROM cabanas c
	JOIN usuarios u ON u.id = c.id_anfitrion
`
func scanCabin(scan func(dest ...any) error) (Cabin, error) {
	var c Cabin

	err := scan(
		&c.ID,
		&c.Name,
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
		INSERT INTO cabanas (nombre, direccion, precio, descripcion, capacidad, reglas, id_anfitrion, id_comision)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		req.Nombre,
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

func scanCabinSearchResult(scan func(dest ...any) error) (CabinCardResponse, error) {
	var card CabinCardResponse

	err := scan(
		&card.ID,
		&card.Name,
		&card.Address,
		&card.Price,
		&card.Capacity,
		&card.HostName,
	)
	return card, err
}

// SearchCabins retorna cabañas aplicando cualquier combinación de filtros
// definidos en params. Un filtro nil se omite del WHERE.
func (r *Repository) SearchCabins(ctx context.Context, params CabinSearchParams) ([]CabinCardResponse, error) {
	query := `SELECT ` + cabinSearchColumns + cabinSearchFrom

	var (
		conditions []string
		args       []any
	)

	addCondition := func(cond string, val any) {
		args = append(args, val)
		conditions = append(conditions, fmt.Sprintf(cond, len(args)))
	}

	if params.HostID != nil {
		addCondition("c.id_anfitrion = $%d", *params.HostID)
	}
	if params.MinCapacity != nil {
		addCondition("c.capacidad >= $%d", *params.MinCapacity)
	}
	if params.MaxCapacity != nil {
		addCondition("c.capacidad <= $%d", *params.MaxCapacity)
	}
	if params.MinPrice != nil {
		addCondition("c.precio >= $%d", *params.MinPrice)
	}
	if params.MaxPrice != nil {
		addCondition("c.precio <= $%d", *params.MaxPrice)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY c.id"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cabins := []CabinCardResponse{}

	for rows.Next() {
		c, err := scanCabinSearchResult(rows.Scan)
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