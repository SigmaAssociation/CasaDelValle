package reservations

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"cdv-api/internal/notifications"
)

type Service struct {
	repository ReservationRepository
	notifier   Notifier
}

// Notifier registra notificaciones en el buzón personal de los usuarios.
// Interfaz declarada en este package para desacoplarlo del módulo de
// notifications y poder sustituirla en pruebas.
type Notifier interface {
	Notify(ctx context.Context, input notifications.NotificationInput)
	NotifyAdmins(ctx context.Context, input notifications.NotificationInput)
}

type ReservationRepository interface {
	CreateReservation(ctx context.Context, req ReservationRequest) (int, error)
	HasOverlap(ctx context.Context, cabinID uint, startDate, endDate string, excludeID int) (bool, error)
	HasUserOverlap(ctx context.Context, userID uint, startDate, endDate string, excludeID int) (bool, error)
	GetReservationByID(ctx context.Context, id int) (*ReservationDetail, error)
	CancelReservation(ctx context.Context, id int, cancelledAt time.Time) error
	GetReservationsByUserID(ctx context.Context, userID uint) ([]ReservationCard, error)
	UpdateReservation(ctx context.Context, id int, req UpdateReservationRequest) error
	GetCabinName(ctx context.Context, cabinID int) (string, error)
	FinalizePastReservations(ctx context.Context) ([]int, error)
}

func NewService(repository ReservationRepository, notifier Notifier) *Service {
	return &Service{
		repository: repository,
		notifier:   notifier,
	}
}

// notify registra una notificación en el buzón del usuario sin interrumpir
// la operación principal: los errores quedan en el log del notifier.
func (s *Service) notify(ctx context.Context, userID uint, tipo, mensaje, entidadTipo string, entidadID int) {
	if s.notifier == nil || userID == 0 {
		return
	}
	s.notifier.Notify(ctx, notifications.NotificationInput{
		UserID:      userID,
		Tipo:        tipo,
		Mensaje:     mensaje,
		EntidadTipo: entidadTipo,
		EntidadID:   entidadID,
	})
}

// notifyAdmins replica un evento del sistema en el buzón de los administradores.
func (s *Service) notifyAdmins(ctx context.Context, tipo, mensaje, entidadTipo string, entidadID int) {
	if s.notifier == nil {
		return
	}
	s.notifier.NotifyAdmins(ctx, notifications.NotificationInput{
		Tipo:        tipo,
		Mensaje:     mensaje,
		EntidadTipo: entidadTipo,
		EntidadID:   entidadID,
	})
}

// cabinDisplayName devuelve el nombre de la cabaña para los mensajes del
// buzón; si no se puede obtener, se usa una referencia genérica por ID.
func (s *Service) cabinDisplayName(ctx context.Context, cabinID uint) string {
	name, err := s.repository.GetCabinName(ctx, int(cabinID))
	if err != nil || strings.TrimSpace(name) == "" {
		return fmt.Sprintf("la cabaña #%d", cabinID)
	}
	return name
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

	reservationID, err := s.repository.CreateReservation(ctx, req)
	if err != nil {
		return 0, err
	}

	s.notifyReservationCreated(ctx, reservationID, req)

	return reservationID, nil
}

// notifyReservationCreated avisa al huésped y al anfitrión de la nueva
// reservación. Se ejecuta después del éxito de la operación; los fallos al
// notificar nunca afectan la reservación.
func (s *Service) notifyReservationCreated(ctx context.Context, reservationID int, req ReservationRequest) {
	detail, err := s.repository.GetReservationByID(ctx, reservationID)
	if err != nil || detail == nil {
		return
	}

	cabinName := s.cabinDisplayName(ctx, req.CabinID)
	period := fmt.Sprintf("del %s al %s", req.StartDate, req.EndDate)

	s.notify(ctx, detail.UserID, notifications.TipoReservaCreada,
		fmt.Sprintf("Tu reserva en %s %s fue registrada exitosamente.", cabinName, period),
		notifications.EntidadReservacion, detail.ID)

	if detail.HostID != detail.UserID {
		s.notify(ctx, detail.HostID, notifications.TipoReservaCreada,
			fmt.Sprintf("Tu cabaña %s recibió una nueva reserva %s.", cabinName, period),
			notifications.EntidadReservacion, detail.ID)
	}

	s.notifyAdmins(ctx, notifications.TipoReservaCreada,
		fmt.Sprintf("Nueva reserva registrada en %s %s (reserva #%d).", cabinName, period, detail.ID),
		notifications.EntidadReservacion, detail.ID)
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

	s.notifyReservationCancelled(ctx, reservation)

	return s.repository.GetReservationByID(ctx, id)
}

// notifyReservationCancelled avisa al huésped y al anfitrión de la
// cancelación de la reservación.
func (s *Service) notifyReservationCancelled(ctx context.Context, reservation *ReservationDetail) {
	if reservation == nil {
		return
	}

	cabinName := s.cabinDisplayName(ctx, reservation.CabinID)
	period := fmt.Sprintf("del %s al %s", reservation.StartDate, reservation.EndDate)

	s.notify(ctx, reservation.UserID, notifications.TipoReservaCancelada,
		fmt.Sprintf("Tu reserva en %s %s fue cancelada.", cabinName, period),
		notifications.EntidadReservacion, reservation.ID)

	if reservation.HostID != reservation.UserID {
		s.notify(ctx, reservation.HostID, notifications.TipoReservaCancelada,
			fmt.Sprintf("La reserva de tu cabaña %s %s fue cancelada.", cabinName, period),
			notifications.EntidadReservacion, reservation.ID)
	}

	s.notifyAdmins(ctx, notifications.TipoReservaCancelada,
		fmt.Sprintf("Se canceló la reserva #%d en %s %s.", reservation.ID, cabinName, period),
		notifications.EntidadReservacion, reservation.ID)
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

	updated, err := s.repository.GetReservationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.notifyReservationUpdated(ctx, updated)

	return updated, nil
}

// notifyReservationUpdated avisa al huésped y al anfitrión de que la
// reservación fue modificada.
func (s *Service) notifyReservationUpdated(ctx context.Context, reservation *ReservationDetail) {
	if reservation == nil {
		return
	}

	cabinName := s.cabinDisplayName(ctx, reservation.CabinID)
	period := fmt.Sprintf("del %s al %s", reservation.StartDate, reservation.EndDate)

	s.notify(ctx, reservation.UserID, notifications.TipoReservaActualizada,
		fmt.Sprintf("Tu reserva en %s %s fue actualizada.", cabinName, period),
		notifications.EntidadReservacion, reservation.ID)

	if reservation.HostID != reservation.UserID {
		s.notify(ctx, reservation.HostID, notifications.TipoReservaActualizada,
			fmt.Sprintf("La reserva de tu cabaña %s %s fue actualizada.", cabinName, period),
			notifications.EntidadReservacion, reservation.ID)
	}

	s.notifyAdmins(ctx, notifications.TipoReservaActualizada,
		fmt.Sprintf("Se actualizó la reserva #%d en %s %s.", reservation.ID, cabinName, period),
		notifications.EntidadReservacion, reservation.ID)
}

// FinalizePastReservations cierra las reservaciones cuya fecha de fin ya pasó
// y avisa al huésped y al anfitrión. Está pensada para ejecutarse de forma
// periódica desde un job en segundo plano; devuelve cuántas finalizó.
func (s *Service) FinalizePastReservations(ctx context.Context) (int, error) {
	ids, err := s.repository.FinalizePastReservations(ctx)
	if err != nil {
		return 0, err
	}

	for _, id := range ids {
		detail, err := s.repository.GetReservationByID(ctx, id)
		if err != nil || detail == nil {
			continue
		}
		s.notifyReservationFinished(ctx, detail)
	}

	return len(ids), nil
}

// notifyReservationFinished avisa al huésped y al anfitrión de que la
// reservación concluyó.
func (s *Service) notifyReservationFinished(ctx context.Context, reservation *ReservationDetail) {
	if reservation == nil {
		return
	}

	cabinName := s.cabinDisplayName(ctx, reservation.CabinID)
	period := fmt.Sprintf("del %s al %s", reservation.StartDate, reservation.EndDate)

	s.notify(ctx, reservation.UserID, notifications.TipoReservaFinalizada,
		fmt.Sprintf("Tu reserva en %s %s ha finalizado. ¡Gracias por tu visita!", cabinName, period),
		notifications.EntidadReservacion, reservation.ID)

	if reservation.HostID != reservation.UserID {
		s.notify(ctx, reservation.HostID, notifications.TipoReservaFinalizada,
			fmt.Sprintf("La reserva de tu cabaña %s %s ha finalizado.", cabinName, period),
			notifications.EntidadReservacion, reservation.ID)
	}

	s.notifyAdmins(ctx, notifications.TipoReservaFinalizada,
		fmt.Sprintf("Finalizó la reserva #%d en %s %s.", reservation.ID, cabinName, period),
		notifications.EntidadReservacion, reservation.ID)
}
