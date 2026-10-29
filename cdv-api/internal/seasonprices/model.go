package seasonprices

import "errors"

type SeasonPrice struct {
	ID      int  `json:"id"`
	CabinID uint `json:"cabin_id"`
}

type DeleteSeasonPriceResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

var (
	ErrSeasonPriceNotFound = errors.New("season price not found")
	ErrCabinNotFound       = errors.New("cabin not found")
)