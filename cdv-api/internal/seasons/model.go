package seasons

import (
	"errors"
	"time"
)

type Season struct {
	ID                  int     `json:"id"`
	Name                string  `json:"name"`
	SuggestedAdjustment float64 `json:"suggestedAdjustmentPercent"`
	Active              bool    `json:"active"`
	StartDate           string  `json:"startDate"`
	EndDate             string  `json:"endDate"`
}

type CabinSeasonPrice struct {
	ID            int       `json:"id"`
	CabinID       uint      `json:"cabinId"`
	SeasonID      *int      `json:"seasonId,omitempty"`
	StartDate     string    `json:"startDate"`
	EndDate       string    `json:"endDate"`
	AdjustmentPct float64   `json:"adjustmentPercent"`
	FixedPrice    *float64  `json:"fixedPrice,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

type CreateSeasonRequest struct {
	Name                string  `json:"name"`
	SuggestedAdjustment float64 `json:"suggestedAdjustmentPercent"`
	Active              *bool   `json:"active,omitempty"`
	StartDate           string  `json:"startDate"`
	EndDate             string  `json:"endDate"`
}

// IsActive devuelve el estado solicitado; si se omite, la temporada nace activa.
func (r CreateSeasonRequest) IsActive() bool {
	if r.Active == nil {
		return true
	}
	return *r.Active
}

type UpdateSeasonRequest struct {
	Name                string  `json:"name"`
	SuggestedAdjustment float64 `json:"suggestedAdjustmentPercent"`
	Active              bool    `json:"active"`
	StartDate           string  `json:"startDate"`
	EndDate             string  `json:"endDate"`
}

type SetCabinSeasonPriceRequest struct {
	SeasonID      *int     `json:"seasonId,omitempty"`
	StartDate     string   `json:"startDate"`
	EndDate       string   `json:"endDate"`
	AdjustmentPct float64  `json:"adjustmentPercent"`
	FixedPrice    *float64 `json:"fixedPrice,omitempty"`
}

type PriceQuoteRequest struct {
	CabinID   uint   `json:"cabinId"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type PriceBreakdown struct {
	StartDate      string  `json:"startDate"`
	EndDate        string  `json:"endDate"`
	Nights         int     `json:"nights"`
	BaseNightly    float64 `json:"baseNightly"`
	AppliedNightly float64 `json:"appliedNightly"`
	Total          float64 `json:"total"`
	Source         string  `json:"source"`
	SeasonName     *string `json:"seasonName,omitempty"`
}

type PriceQuoteResponse struct {
	CabinID        uint             `json:"cabinId"`
	BaseNightly    float64          `json:"baseNightly"`
	TotalNights    int              `json:"totalNights"`
	TotalPrice     float64          `json:"totalPrice"`
	AverageNightly float64          `json:"averageNightly"`
	Breakdown      []PriceBreakdown `json:"breakdown"`
}

var (
	ErrSeasonNotFound      = errors.New("season not found")
	ErrSeasonOverlap       = errors.New("season dates overlap")
	ErrCabinSeasonNotFound = errors.New("cabin season price not found")
	ErrInvalidDates        = errors.New("invalid dates")
	ErrInvalidPrice        = errors.New("invalid seasonal price")
	ErrInvalidName         = errors.New("invalid season name")
	ErrInvalidAdjustment   = errors.New("invalid suggested adjustment")
)

type ErrorResponse struct {
	Message string `json:"message"`
}
