package seasonprices

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetByID(ctx context.Context, id int) (*SeasonPrice, error) {
	query := `SELECT id, id_cabana FROM precios_temporada WHERE id = $1`

	var sp SeasonPrice
	err := r.pool.QueryRow(ctx, query, id).Scan(&sp.ID, &sp.CabinID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSeasonPriceNotFound
		}
		return nil, err
	}
	return &sp, nil
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM precios_temporada WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrSeasonPriceNotFound
	}
	return nil
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