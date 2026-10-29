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
	base, err := s.repo.GetCabinBasePrice(ctx, req.CabinID)
	if err != nil {
		return nil, err
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, ErrInvalidDates
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, ErrInvalidDates
	}
	if !end.After(start) {
		return nil, ErrInvalidDates
	}

	breakdown := []PriceBreakdown{}
	total := 0.0
	cur := start
	for cur.Before(end) {
		next := cur.AddDate(0, 0, 1)
		day := cur.Format("2006-01-02")
		applied := base
		source := "normal"
		var seasonName *string

		if cabinPrice, err := s.repo.GetMatchingCabinSeason(ctx, req.CabinID, day); err == nil && cabinPrice != nil {
			if cabinPrice.FixedPrice != nil && *cabinPrice.FixedPrice > 0 {
				applied = *cabinPrice.FixedPrice
				source = "cabin_fixed"
			} else if cabinPrice.AdjustmentPct != 0 {
				applied = base * (1 + cabinPrice.AdjustmentPct/100)
				source = "cabin_percent"
			}
			if cabinPrice.SeasonID != nil {
				if seasons, err := s.repo.ListSeasons(ctx, true); err == nil {
					for _, se := range seasons {
						if se.ID == *cabinPrice.SeasonID {
							n := se.Name
							seasonName = &n
							break
						}
					}
				}
			}
		} else if global, err := s.repo.GetMatchingGlobalSeason(ctx, day); err == nil && global != nil && global.Active {
			applied = base * (1 + global.SuggestedAdjustment/100)
			source = "global_season"
			n := global.Name
			seasonName = &n
		}

		applied = round2(applied)
		breakdown = append(breakdown, PriceBreakdown{
			StartDate:      day,
			EndDate:        day,
			Nights:         1,
			BaseNightly:    round2(base),
			AppliedNightly: applied,
			Total:          applied,
			Source:         source,
			SeasonName:     seasonName,
		})
		total += applied
		cur = next
	}

	totalNights := int(end.Sub(start).Hours() / 24)
	return &PriceQuoteResponse{
		CabinID:        req.CabinID,
		BaseNightly:    round2(base),
		TotalNights:    totalNights,
		TotalPrice:     round2(total),
		AverageNightly: round2(total / float64(totalNights)),
		Breakdown:      breakdown,
	}, nil
}
