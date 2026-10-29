package maintenance

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"
)

type MaintenanceRepository interface {
	List(context.Context, int, uint) ([]Maintenance, error)
	Preview(context.Context, int, uint, DateRequest) (*Preview, error)
	Create(context.Context, int, uint, CreateRequest) (*Maintenance, error)
	Cancel(context.Context, int, int, uint) (*Maintenance, error)
	Pending(context.Context, int, uint) (*PendingDecision, error)
	Resolve(context.Context, int, uint, DecisionRequest) error
}

type Service struct{ repository MaintenanceRepository }

func NewService(r MaintenanceRepository) *Service { return &Service{repository: r} }

func validateDates(start, end string, now time.Time) error {
	a, err := time.Parse("2006-01-02", start)
	if err != nil || a.Format("2006-01-02") != start {
		return validationError("La fecha de inicio debe ser válida (AAAA-MM-DD)")
	}
	b, err := time.Parse("2006-01-02", end)
	if err != nil || b.Format("2006-01-02") != end {
		return validationError("La fecha de fin debe ser válida (AAAA-MM-DD)")
	}
	if !b.After(a) {
		return validationError("La fecha de fin debe ser posterior a la fecha de inicio")
	}
	loc, err := time.LoadLocation("America/Guatemala")
	if err != nil {
		loc = time.FixedZone("GT", -6*60*60)
	}
	if start < now.In(loc).Format("2006-01-02") {
		return validationError("El mantenimiento no puede iniciar antes de hoy")
	}
	return nil
}
func (s *Service) List(ctx context.Context, cabinID int, actor uint) ([]Maintenance, error) {
	return s.repository.List(ctx, cabinID, actor)
}
func (s *Service) Preview(ctx context.Context, cabinID int, actor uint, req DateRequest) (*Preview, error) {
	if err := validateDates(req.StartDate, req.EndDate, time.Now()); err != nil {
		return nil, err
	}
	return s.repository.Preview(ctx, cabinID, actor, req)
}
func (s *Service) Create(ctx context.Context, cabinID int, actor uint, req CreateRequest) (*Maintenance, error) {
	if err := validateDates(req.StartDate, req.EndDate, time.Now()); err != nil {
		return nil, err
	}
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" || utf8.RuneCountInString(req.Reason) > 500 {
		return nil, validationError("Ingresa un motivo de hasta 500 caracteres")
	}
	return s.repository.Create(ctx, cabinID, actor, req)
}
func (s *Service) Cancel(ctx context.Context, cabinID, id int, actor uint) (*Maintenance, error) {
	return s.repository.Cancel(ctx, cabinID, id, actor)
}

func (s *Service) Pending(ctx context.Context, id int, actor uint) (*PendingDecision, error) {
	return s.repository.Pending(ctx, id, actor)
}
func (s *Service) Resolve(ctx context.Context, id int, actor uint, req DecisionRequest) error {
	switch req.Action {
	case "reprogramar":
		if err := validateDates(req.StartDate, req.EndDate, time.Now()); err != nil {
			return err
		}
	case "reembolsar":
		if req.StartDate != "" || req.EndDate != "" {
			return validationError("Un reembolso no debe incluir nuevas fechas")
		}
	default:
		return validationError("Selecciona cambiar las fechas o solicitar el reembolso total")
	}
	return s.repository.Resolve(ctx, id, actor, req)
}
