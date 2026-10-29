package reportincome

import (
	"context"
	"errors"
	"time"
)

const (
	dateLayout    = "2006-01-02"
	refundType    = "reembolso"
	noDataMessage = "No se encontraron ingresos en el rango seleccionado."
)

var (
	ErrInvalidFilter    = errors.New("cabin_id debe ser un identificador positivo")
	ErrInvalidDateRange = errors.New("start_date no puede ser posterior a end_date")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GenerateReport(ctx context.Context, filter ReportFilter) (*IncomeReport, error) {
	if filter.StartDate.IsZero() || filter.EndDate.IsZero() {
		return nil, ErrInvalidDateRange
	}

	if filter.CabinID != nil && *filter.CabinID <= 0 {
		return nil, ErrInvalidFilter
	}

	start := dateOnly(filter.StartDate)
	end := dateOnly(filter.EndDate)
	if start.After(end) {
		return nil, ErrInvalidDateRange
	}

	filter.StartDate = start
	filter.EndDate = end

	movements, err := s.repo.GetReport(ctx, filter)
	if err != nil {
		return nil, err
	}

	report := &IncomeReport{
		Movements:      movements,
		TotalMovements: len(movements),
		StartDate:      start.Format(dateLayout),
		EndDate:        end.Format(dateLayout),
	}

	for _, movement := range movements {
		if movement.Type == refundType {
			report.TotalRefunds += -movement.Amount
			continue
		}
		report.TotalPayments += movement.Amount
	}

	report.TotalIncome = report.TotalPayments - report.TotalRefunds

	if report.TotalMovements == 0 {
		report.Message = noDataMessage
	}

	return report, nil
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}
