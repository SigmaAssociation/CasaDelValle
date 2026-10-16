package reportreservations

import (
	"context"
	"errors"
)

const noDataMessage = "No reservations found for the selected filters."

var ErrInvalidFilter = errors.New("filters must be positive identifiers")

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GenerateReport(ctx context.Context, filter ReportFilter) (*ReservationsReport, error) {
	if (filter.HostID != nil && *filter.HostID <= 0) ||
		(filter.CabinID != nil && *filter.CabinID <= 0) ||
		(filter.GuestID != nil && *filter.GuestID <= 0) {
		return nil, ErrInvalidFilter
	}

	cabins, err := s.repo.GetReport(ctx, filter)
	if err != nil {
		return nil, err
	}

	report := &ReservationsReport{
		Cabins:      cabins,
		TotalCabins: len(cabins),
	}
	for _, c := range cabins {
		report.TotalReservations += c.TotalReservations
	}

	if report.TotalReservations == 0 {
		report.Message = noDataMessage
	}

	return report, nil
}