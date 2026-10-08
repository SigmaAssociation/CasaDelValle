package reservations

import (
	"context"
	"errors"
	"regexp"
	"time"
)

type Service struct {
	repository ReservationRepository
}

type ReservationRepository interface {
	CreateReservation(ctx context.Context, req ReservationRequest) (int, error)
	HasOverlap(ctx context.Context, cabinID uint, startDate, endDate string, excludeID int) (bool, error)
	HasUserOverlap(ctx context.Context, userID uint, startDate, endDate string, excludeID int) (bool, error)
	GetReservationByID(ctx context.Context, id int) (*ReservationDetail, error)
	CancelReservation(ctx context.Context, id int, cancelledAt time.Time) error
	GetReservationsByUserID(ctx context.Context, userID uint) ([]ReservationCard, error)
	UpdateReservation(ctx context.Context, id int, req UpdateReservationRequest) error
}

func NewService(repository ReservationRepository) *Service {
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

func calculateReservationNights(startDate, endDate string) (int, error) {
	start, err := time.ParseInLocation(dateLayout, startDate, time.UTC)
	if err != nil {
		return 0, err
	}

	end, err := time.ParseInLocation(dateLayout, endDate, time.UTC)
	if err != nil {
		return 0, err
	}

	return int(end.Sub(start).Hours() / 24), nil
}

func calculateReservationTotalPrice(startDate, endDate string, nightlyRate float64) (float64, error) {
	nights, err := calculateReservationNights(startDate, endDate)
	if err != nil {
		return 0, err
	}

	return float64(nights) * nightlyRate, nil
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

	overlap, err := s.repository.HasOverlap(ctx, req.CabinID, req.StartDate, req.EndDate, 0)
	if err != nil {
		return 0, err
	}
	if overlap {
		return 0, errors.New("La cabaña no está disponible en el rango de fechas seleccionado")
	}

	userOverlap, err := s.repository.HasUserOverlap(ctx, req.UserID, req.StartDate, req.EndDate, 0)
	if err != nil {
		return 0, err
	}
	if userOverlap {
		return 0, errors.New("El usuario ya tiene otra reservación en el rango de fechas seleccionado")
	}

	return s.repository.CreateReservation(ctx, req)
}

const cancellationTimezone = "America/Guatemala"

func cancellationLocation() *time.Location {
	loc, err := time.LoadLocation(cancellationTimezone)
	if err != nil {
		return time.FixedZone("GT", -6*60*60)
	}
	return loc
}

func (s *Service) GetReservationByID(ctx context.Context, id int) (*ReservationDetail, error) {
	return s.repository.GetReservationByID(ctx, id)
}

func (s *Service) CancelReservation(ctx context.Context, id int) (*ReservationDetail, error) {
	reservation, err := s.repository.GetReservationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if reservation.Status != StatusActive {
		return nil, ErrReservationNotCancellable
	}

	loc := cancellationLocation()
	now := time.Now().In(loc)

	startDate, err := time.ParseInLocation(dateLayout, reservation.StartDate, loc)
	if err != nil {
		return nil, err
	}

	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, loc)
	limit := today.AddDate(0, 0, MinCancellationDays)

	if startDate.Before(limit) {
		return nil, ErrCancellationTooLate
	}

	if err := s.repository.CancelReservation(ctx, id, now); err != nil {
		return nil, err
	}

	return s.repository.GetReservationByID(ctx, id)
}

func (s *Service) GetReservationsByUserID(ctx context.Context, userID uint) ([]ReservationCard, error) {
	if userID == 0 {
		return nil, errors.New("El usuario es requerido")
	}

	return s.repository.GetReservationsByUserID(ctx, userID)
}
func (s *Service) UpdateReservation(ctx context.Context, id int, req UpdateReservationRequest) (*ReservationDetail, error) {
	if id <= 0 {
		return nil, errors.New("ID de reservación inválido")
	}
	if req.CabinID == 0 {
		return nil, errors.New("La cabaña es requerida")
	}
	if req.StartDate == "" {
		return nil, errors.New("La fecha de inicio es requerida")
	}
	if req.EndDate == "" {
		return nil, errors.New("La fecha de fin es requerida")
	}

	startDate, err := parseReservationDate(req.StartDate, "fecha de inicio")
	if err != nil {
		return nil, err
	}
	endDate, err := parseReservationDate(req.EndDate, "fecha de fin")
	if err != nil {
		return nil, err
	}

	if startDate.After(endDate) {
		return nil, errors.New("La fecha de inicio no puede ser posterior a la fecha de fin")
	}
	if startDate.Equal(endDate) {
		return nil, errors.New("La fecha de inicio y la fecha de fin no pueden ser iguales")
	}
	if startDate.Before(startOfToday()) {
		return nil, errors.New("La fecha de inicio no puede ser anterior a la fecha actual")
	}

	existing, err := s.repository.GetReservationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.Status != StatusActive {
		return nil, ErrReservationNotEditable
	}

	overlap, err := s.repository.HasOverlap(ctx, req.CabinID, req.StartDate, req.EndDate, id)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, errors.New("La cabaña no está disponible en el rango de fechas seleccionado")
	}

	userOverlap, err := s.repository.HasUserOverlap(ctx, existing.UserID, req.StartDate, req.EndDate, id)
	if err != nil {
		return nil, err
	}
	if userOverlap {
		return nil, errors.New("El usuario ya tiene otra reservación en el rango de fechas seleccionado")
	}

	if err := s.repository.UpdateReservation(ctx, id, req); err != nil {
		return nil, err
	}

	return s.repository.GetReservationByID(ctx, id)
}
