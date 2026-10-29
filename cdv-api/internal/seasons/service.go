package seasons

import (
	"context"
	"math"
	"strings"
	"time"
	"unicode/utf8"
)

type SeasonRepository interface {
	ListSeasons(ctx context.Context, includeInactive bool) ([]Season, error)
	CreateSeason(ctx context.Context, req CreateSeasonRequest) (*Season, error)
	UpdateSeason(ctx context.Context, id int, req UpdateSeasonRequest) error
	SetSeasonActive(ctx context.Context, id int, active bool) error
	ListCabinSeasonPrices(ctx context.Context, cabinID uint) ([]CabinSeasonPrice, error)
	SetCabinSeasonPrice(ctx context.Context, cabinID uint, req SetCabinSeasonPriceRequest) (*CabinSeasonPrice, error)
	DeleteCabinSeasonPrice(ctx context.Context, id int) error
	GetCabinBasePrice(ctx context.Context, cabinID uint) (float64, error)
	GetMatchingCabinSeason(ctx context.Context, cabinID uint, date string) (*CabinSeasonPrice, error)
	GetMatchingGlobalSeason(ctx context.Context, date string) (*Season, error)
	GetPriceQuote(ctx context.Context, req PriceQuoteRequest) (*PriceQuoteResponse, error)
}

type Service struct {
	repo SeasonRepository
}

func NewService(repo SeasonRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListSeasons(ctx context.Context, includeInactive bool) ([]Season, error) {
	return s.repo.ListSeasons(ctx, includeInactive)
}

func (s *Service) CreateSeason(ctx context.Context, req CreateSeasonRequest) (*Season, error) {
	if err := validateSeason(req.Name, req.SuggestedAdjustment, req.StartDate, req.EndDate); err != nil {
		return nil, err
	}
	return s.repo.CreateSeason(ctx, req)
}

func (s *Service) UpdateSeason(ctx context.Context, id int, req UpdateSeasonRequest) error {
	if err := validateSeason(req.Name, req.SuggestedAdjustment, req.StartDate, req.EndDate); err != nil {
		return err
	}
	return s.repo.UpdateSeason(ctx, id, req)
}

// DeactivateSeason desactiva una temporada sin eliminarla, para conservar el
// historial de precios que la referencian.
func (s *Service) DeactivateSeason(ctx context.Context, id int) error {
	return s.repo.SetSeasonActive(ctx, id, false)
}

func (s *Service) ListCabinSeasonPrices(ctx context.Context, cabinID uint) ([]CabinSeasonPrice, error) {
	return s.repo.ListCabinSeasonPrices(ctx, cabinID)
}

func (s *Service) SetCabinSeasonPrice(ctx context.Context, cabinID uint, req SetCabinSeasonPriceRequest) (*CabinSeasonPrice, error) {
	if err := validateDates(req.StartDate, req.EndDate); err != nil {
		return nil, err
	}
	if req.AdjustmentPct <= -100 || math.IsNaN(req.AdjustmentPct) || math.IsInf(req.AdjustmentPct, 0) {
		return nil, ErrInvalidPrice
	}
	if req.FixedPrice != nil && *req.FixedPrice <= 0 {
		return nil, ErrInvalidPrice
	}
	return s.repo.SetCabinSeasonPrice(ctx, cabinID, req)
}

func validateSeason(name string, suggestedAdjustment float64, startDate, endDate string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || utf8.RuneCountInString(trimmed) > 100 {
		return ErrInvalidName
	}
	if math.IsNaN(suggestedAdjustment) || math.IsInf(suggestedAdjustment, 0) ||
		suggestedAdjustment < 0 || suggestedAdjustment > 100 {
		return ErrInvalidAdjustment
	}
	return validateDates(startDate, endDate)
}

func validateDates(startDate, endDate string) error {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return ErrInvalidDates
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil || end.Before(start) {
		return ErrInvalidDates
	}
	return nil
}

func (s *Service) DeleteCabinSeasonPrice(ctx context.Context, id int) error {
	return s.repo.DeleteCabinSeasonPrice(ctx, id)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func (s *Service) GetPriceQuote(ctx context.Context, req PriceQuoteRequest) (*PriceQuoteResponse, error) {
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, ErrInvalidDates
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil || !end.After(start) {
		return nil, ErrInvalidDates
	}
	return s.repo.GetPriceQuote(ctx, req)
}
