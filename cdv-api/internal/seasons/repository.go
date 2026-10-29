package seasons

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

func (r *Repository) ListSeasons(ctx context.Context, includeInactive bool) ([]Season, error) {
	query := `
		SELECT id, nombre, porcentaje_ajuste_sugerido, activo, to_char(fecha_inicio,'YYYY-MM-DD'), to_char(fecha_fin,'YYYY-MM-DD')
		FROM temporada
	`
	if !includeInactive {
		query += " WHERE activo = TRUE"
	}
	query += " ORDER BY fecha_inicio"

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seasons := []Season{}
	for rows.Next() {
		var s Season
		if err := rows.Scan(&s.ID, &s.Name, &s.SuggestedAdjustment, &s.Active, &s.StartDate, &s.EndDate); err != nil {
			return nil, err
		}
		seasons = append(seasons, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return seasons, nil
}

func (r *Repository) CreateSeason(ctx context.Context, req CreateSeasonRequest) (*Season, error) {
	var id int
	query := `
		INSERT INTO temporada (nombre, porcentaje_ajuste_sugerido, activo, fecha_inicio, fecha_fin)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id
	`
	active := req.IsActive()
	err := r.pool.QueryRow(ctx, query, req.Name, req.SuggestedAdjustment, active, req.StartDate, req.EndDate).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
			return nil, ErrSeasonOverlap
		}
		return nil, err
	}

	return &Season{
		ID:                  id,
		Name:                req.Name,
		SuggestedAdjustment: req.SuggestedAdjustment,
		Active:              active,
		StartDate:           req.StartDate,
		EndDate:             req.EndDate,
	}, nil
}

func (r *Repository) UpdateSeason(ctx context.Context, id int, req UpdateSeasonRequest) error {
	query := `
		UPDATE temporada
		SET nombre=$1, porcentaje_ajuste_sugerido=$2, activo=$3, fecha_inicio=$4, fecha_fin=$5
		WHERE id=$6
	`
	tag, err := r.pool.Exec(ctx, query, req.Name, req.SuggestedAdjustment, req.Active, req.StartDate, req.EndDate, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
			return ErrSeasonOverlap
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrSeasonNotFound
	}
	return nil
}

func (r *Repository) SetSeasonActive(ctx context.Context, id int, active bool) error {
	tag, err := r.pool.Exec(ctx, `UPDATE temporada SET activo=$1 WHERE id=$2`, active, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
			return ErrSeasonOverlap
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrSeasonNotFound
	}
	return nil
}

func (r *Repository) GetCabinBasePrice(ctx context.Context, cabinID uint) (float64, error) {
	var price float64
	err := r.pool.QueryRow(ctx, `SELECT precio FROM cabanas WHERE id=$1`, cabinID).Scan(&price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errors.New("cabin not found")
		}
		return 0, err
	}
	return price, nil
}

func (r *Repository) ListCabinSeasonPrices(ctx context.Context, cabinID uint) ([]CabinSeasonPrice, error) {
	query := `
		SELECT id, id_cabana, id_temporada, to_char(fecha_inicio,'YYYY-MM-DD'), to_char(fecha_fin,'YYYY-MM-DD'), porcentaje_ajuste, precio_fijo, created_at
		FROM precios_temporada
		WHERE id_cabana=$1
		ORDER BY fecha_inicio
	`
	rows, err := r.pool.Query(ctx, query, cabinID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []CabinSeasonPrice{}
	for rows.Next() {
		var p CabinSeasonPrice
		var seasonID *int
		var fixed *float64
		if err := rows.Scan(&p.ID, &p.CabinID, &seasonID, &p.StartDate, &p.EndDate, &p.AdjustmentPct, &fixed, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.SeasonID = seasonID
		p.FixedPrice = fixed
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) SetCabinSeasonPrice(ctx context.Context, cabinID uint, req SetCabinSeasonPriceRequest) (*CabinSeasonPrice, error) {
	var id int
	var createdAt time.Time
	query := `
		INSERT INTO precios_temporada (id_cabana, id_temporada, fecha_inicio, fecha_fin, porcentaje_ajuste, precio_fijo)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query, cabinID, req.SeasonID, req.StartDate, req.EndDate, req.AdjustmentPct, req.FixedPrice).Scan(&id, &createdAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
			return nil, ErrSeasonOverlap
		}
		return nil, err
	}
	return &CabinSeasonPrice{
		ID:            id,
		CabinID:       cabinID,
		SeasonID:      req.SeasonID,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		AdjustmentPct: req.AdjustmentPct,
		FixedPrice:    req.FixedPrice,
		CreatedAt:     createdAt,
	}, nil
}

func (r *Repository) DeleteCabinSeasonPrice(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM precios_temporada WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCabinSeasonNotFound
	}
	return nil
}

func (r *Repository) GetMatchingGlobalSeason(ctx context.Context, date string) (*Season, error) {
	query := `
		SELECT id, nombre, porcentaje_ajuste_sugerido, activo, to_char(fecha_inicio,'YYYY-MM-DD'), to_char(fecha_fin,'YYYY-MM-DD')
		FROM temporada
		WHERE activo = TRUE AND $1::date BETWEEN fecha_inicio AND fecha_fin
		ORDER BY id
		LIMIT 1
	`
	var s Season
	err := r.pool.QueryRow(ctx, query, date).Scan(&s.ID, &s.Name, &s.SuggestedAdjustment, &s.Active, &s.StartDate, &s.EndDate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *Repository) GetMatchingCabinSeason(ctx context.Context, cabinID uint, date string) (*CabinSeasonPrice, error) {
	query := `
		SELECT id, id_cabana, id_temporada, to_char(fecha_inicio,'YYYY-MM-DD'), to_char(fecha_fin,'YYYY-MM-DD'), porcentaje_ajuste, precio_fijo, created_at
		FROM precios_temporada
		WHERE id_cabana=$1 AND $2::date BETWEEN fecha_inicio AND fecha_fin
		ORDER BY id
		LIMIT 1
	`
	var p CabinSeasonPrice
	var seasonID *int
	var fixed *float64
	err := r.pool.QueryRow(ctx, query, cabinID, date).Scan(&p.ID, &p.CabinID, &seasonID, &p.StartDate, &p.EndDate, &p.AdjustmentPct, &fixed, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	p.SeasonID = seasonID
	p.FixedPrice = fixed
	return &p, nil
}

// GetPriceQuote comparte el cálculo SQL usado al guardar precio_total.
func (r *Repository) GetPriceQuote(ctx context.Context, req PriceQuoteRequest) (*PriceQuoteResponse, error) {
	base, err := r.GetCabinBasePrice(ctx, req.CabinID)
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `SELECT to_char(dia,'YYYY-MM-DD'),base,precio,fuente,temporada_nombre
        FROM cotizar_noches($1,$2,$3)`, req.CabinID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	quote := &PriceQuoteResponse{CabinID: req.CabinID, BaseNightly: base, Breakdown: []PriceBreakdown{}}
	for rows.Next() {
		var b PriceBreakdown
		if err := rows.Scan(&b.StartDate, &b.BaseNightly, &b.AppliedNightly, &b.Source, &b.SeasonName); err != nil {
			return nil, err
		}
		if len(quote.Breakdown) == 0 {
			quote.BaseNightly = b.BaseNightly
		}
		b.EndDate = b.StartDate
		b.Nights = 1
		b.Total = b.AppliedNightly
		quote.TotalPrice += b.Total
		quote.Breakdown = append(quote.Breakdown, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	quote.TotalNights = len(quote.Breakdown)
	if quote.TotalNights == 0 {
		return nil, ErrInvalidDates
	}
	quote.TotalPrice = round2(quote.TotalPrice)
	quote.AverageNightly = round2(quote.TotalPrice / float64(quote.TotalNights))
	return quote, nil
}
