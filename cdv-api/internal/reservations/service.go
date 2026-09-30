package reservations

import (
	"context"
	"errors"
	"regexp"
	"time"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

const dateLayout = "2006-01-02"

var dateFormatRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func parseReservationDate(value, fieldName string) (time.Time, error) {
	if !dateFormatRegex.MatchString(value) {
		return time.Time{}, errors.New("La " + fieldName + " debe tener el formato AAAA-MM-DD")
	}
	parsed, err := time.Parse(dateLayout, value)
	if err != nil {
		return time.Time{}, errors.New("La " + fieldName + " no es una fecha válida")
	}
	return parsed, nil
}

func startOfToday() time.Time {
	now := time.Now()
	year, month, day := now.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, now.Location())
}

func (s *Service) CreateReservation(ctx context.Context, req ReservationRequest) (int, error) {
	if req.UserID == 0 {
		return 0, errors.New("El usuario es requerido")
	}
	if req.CabinID == 0 {
		return 0, errors.New("La cabaña es requerida")
	}
	if req.StartDate == "" {
		return 0, errors.New("La fecha de inicio es requerida")
	}
	if req.EndDate == "" {
		return 0, errors.New("La fecha de fin es requerida")
	}

	startDate, err := parseReservationDate(req.StartDate, "fecha de inicio")
	if err != nil {
		return 0, err
	}
	endDate, err := parseReservationDate(req.EndDate, "fecha de fin")
	if err != nil {
		return 0, err
	}

	if startDate.After(endDate) {
		return 0, errors.New("La fecha de inicio no puede ser posterior a la fecha de fin")
	}
	if startDate.Equal(endDate) {
		return 0, errors.New("La fecha de inicio y la fecha de fin no pueden ser iguales")
	}
	if startDate.Before(startOfToday()) {
		return 0, errors.New("La fecha de inicio no puede ser anterior a la fecha actual")
	}

	overlap, err := s.repository.HasOverlap(ctx, req.CabinID, req.StartDate, req.EndDate)
	if err != nil {
		return 0, err
	}
	if overlap {
		return 0, errors.New("La cabaña no está disponible en el rango de fechas seleccionado")
	}

	return s.repository.CreateReservation(ctx, req)
}
