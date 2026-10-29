package reservations

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
	"strings"
	"time"

	"cdv-api/internal/notifications"
)

type Service struct {
	repository     ReservationRepository
	notifier       Notifier
	refundRecorder RefundRecorder
}

type Notifier interface {
	Notify(ctx context.Context, input notifications.NotificationInput)
	NotifyAdmins(ctx context.Context, input notifications.NotificationInput)
}

type RefundRecorder interface {
	RecordRefundOnCancellation(ctx context.Context, reservation *ReservationDetail) (*CancelRefund, error)
}

type ReservationRepository interface {
	CreateReservation(ctx context.Context, req ReservationRequest) (int, error)
	HasOverlap(ctx context.Context, cabinID uint, startDate, endDate string, excludeID int) (bool, error)
	HasUserOverlap(ctx context.Context, userID uint, startDate, endDate string, excludeID int) (bool, error)
	GetBlockedDateRanges(ctx context.Context, cabinID uint) ([]DateRange, error)
	GetReservationByID(ctx context.Context, id int) (*ReservationDetail, error)
	CancelReservation(ctx context.Context, id int, cancelledAt time.Time) error
	GetReservationsByUserID(ctx context.Context, userID uint) ([]ReservationCard, error)
	GetReservationsByCabinID(ctx context.Context, cabinID uint) ([]ReservationCard, error)
	UpdateReservation(ctx context.Context, id int, req UpdateReservationRequest) error
	GetCabinName(ctx context.Context, cabinID int) (string, error)
	GetCabinHostID(ctx context.Context, cabinID uint) (uint, error)
	GetReservationLedger(ctx context.Context, reservationID int, cabinID uint) (*ReservationLedger, error)
	GetCabinNightlyRate(ctx context.Context, cabinID uint) (float64, error)
	PayAnticipo(ctx context.Context, reservationID int, payerID uint, amount float64, method string) error
	GetCheckinEligibility(ctx context.Context, id int) (*CheckinEligibility, error)
	RecordCheckin(ctx context.Context, id int, checkedInAt time.Time) error
	RecordCheckinTimestamp(ctx context.Context, id int, checkedInAt time.Time) error
	CheckinWithCashPayment(ctx context.Context, id int, payerID uint, cashAmount float64, checkedInAt time.Time) error
	RecordCheckout(ctx context.Context, id int, checkedOutAt time.Time) error
	AdvanceReservationStates(ctx context.Context) ([]int, error)
}

func NewService(repository ReservationRepository, notifier Notifier) *Service {
	return &Service{
		repository: repository,
		notifier:   notifier,
	}
}

func (s *Service) SetRefundRecorder(recorder RefundRecorder) {
	s.refundRecorder = recorder
}

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

func (s *Service) notifyReservationCreated(ctx context.Context, reservationID int, req ReservationRequest) {
	detail, err := s.repository.GetReservationByID(ctx, reservationID)
	if err != nil || detail == nil {
		return
	}

	cabinName := s.cabinDisplayName(ctx, req.CabinID)
	period := fmt.Sprintf("del %s al %s", req.StartDate, req.EndDate)

	s.notify(ctx, detail.UserID, notifications.TipoReservaCreada,
		fmt.Sprintf("Tu reserva en %s %s fue registrada. Paga el anticipo para confirmarla.", cabinName, period),
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

func daysUntilStartDate(startDate string, loc *time.Location, now time.Time) (int, bool) {
	start, err := time.ParseInLocation(dateLayout, startDate, loc)
	if err != nil {
		return 0, false
	}

	y, m, d := now.In(loc).Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, loc)
	return int(math.Round(start.Sub(today).Hours() / 24)), true
}

func formatMoney(value float64) string {
	return fmt.Sprintf("Q %.2f", value)
}

func (s *Service) GetReservationByID(ctx context.Context, id int) (*ReservationDetail, error) {
	return s.repository.GetReservationByID(ctx, id)
}

func (s *Service) CancelReservation(ctx context.Context, id int) (*ReservationDetail, error) {
	detail, _, err := s.CancelReservationWithRefund(ctx, id)
	return detail, err
}

func (s *Service) CancelReservationWithRefund(ctx context.Context, id int) (*ReservationDetail, *CancelRefund, error) {
	reservation, err := s.repository.GetReservationByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	if reservation.Status != StatusPending && reservation.Status != StatusConfirmed {
		return nil, nil, ErrReservationNotCancellable
	}

	loc := cancellationLocation()
	now := time.Now().In(loc)

	if err := s.repository.CancelReservation(ctx, id, now); err != nil {
		return nil, nil, err
	}

	refund, err := s.recordRefundOnCancellation(ctx, reservation)
	if err != nil {
		log.Printf("reservations: no se pudo registrar el reembolso de la reservación %d: %v", id, err)
	}

	daysLeft, _ := daysUntilStartDate(reservation.StartDate, loc, now)
	s.notifyReservationCancelled(ctx, reservation, refund, daysLeft < MinCancellationDays)

	updated, err := s.repository.GetReservationByID(ctx, id)
	return updated, refund, err
}

func (s *Service) Checkin(ctx context.Context, id int, cashAmount float64) (*ReservationDetail, error) {
	if id <= 0 {
		return nil, errors.New("ID de reservación inválido")
	}
	if cashAmount < 0 {
		return nil, errors.New("El monto en efectivo no puede ser negativo")
	}
	cashAmount = math.Round(cashAmount*100) / 100

	eligibility, err := s.repository.GetCheckinEligibility(ctx, id)
	if err != nil {
		return nil, err
	}

	recordTimestampOnly := false
	switch {
	case eligibility.Status == StatusConfirmed:
	case eligibility.Status == StatusInProgress && eligibility.CheckedInAt == nil:
		recordTimestampOnly = true
	default:
		return nil, ErrReservationNotCheckinable
	}

	startDate, err := parseReservationDate(eligibility.StartDate, "fecha de inicio")
	if err != nil {
		return nil, err
	}
	endDate, err := parseReservationDate(eligibility.EndDate, "fecha de fin")
	if err != nil {
		return nil, err
	}

	today := startOfToday()
	if today.Before(startDate) {
		return nil, errors.New("El check-in solo puede registrarse a partir de la fecha de inicio de la reserva")
	}
	if today.After(endDate) {
		return nil, errors.New("El periodo de la reserva ya finalizó")
	}

	total, err := calculateReservationTotalPrice(eligibility.StartDate, eligibility.EndDate, eligibility.NightlyRate)
	if err != nil {
		return nil, err
	}
	if eligibility.TotalPrice != nil {
		total = *eligibility.TotalPrice
	}
	total = math.Round(total*100) / 100
	remaining := math.Round((total-eligibility.PaidTotal)*100) / 100

	if cashAmount > 0 {
		if cashAmount != remaining {
			return nil, errors.New("El monto en efectivo debe cubrir exactamente el saldo pendiente")
		}
		if err := s.repository.CheckinWithCashPayment(ctx, id, eligibility.UserID, cashAmount, time.Now()); err != nil {
			return nil, err
		}
	} else {
		if remaining > 0 {
			return nil, ErrCheckinPaymentIncomplete
		}
		if recordTimestampOnly {
			if err := s.repository.RecordCheckinTimestamp(ctx, id, time.Now()); err != nil {
				return nil, err
			}
		} else if err := s.repository.RecordCheckin(ctx, id, time.Now()); err != nil {
			return nil, err
		}
	}

	updated, err := s.repository.GetReservationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.notifyCheckin(ctx, updated)

	return updated, nil
}

func (s *Service) notifyCheckin(ctx context.Context, reservation *ReservationDetail) {
	if reservation == nil {
		return
	}

	cabinName := s.cabinDisplayName(ctx, reservation.CabinID)
	period := fmt.Sprintf("del %s al %s", reservation.StartDate, reservation.EndDate)

	s.notify(ctx, reservation.UserID, notifications.TipoReservaCheckin,
		fmt.Sprintf("Tu estadía en %s %s ha iniciado. ¡Buen viaje!", cabinName, period),
		notifications.EntidadReservacion, reservation.ID)

	if reservation.HostID != reservation.UserID {
		s.notify(ctx, reservation.HostID, notifications.TipoReservaCheckin,
			fmt.Sprintf("Se registró el check-in de tu cabaña %s %s.", cabinName, period),
			notifications.EntidadReservacion, reservation.ID)
	}

	s.notifyAdmins(ctx, notifications.TipoReservaCheckin,
		fmt.Sprintf("Check-in registrado en %s %s (reserva #%d).", cabinName, period, reservation.ID),
		notifications.EntidadReservacion, reservation.ID)
}

func (s *Service) recordRefundOnCancellation(ctx context.Context, reservation *ReservationDetail) (*CancelRefund, error) {
	if s.refundRecorder == nil {
		return nil, nil
	}
	return s.refundRecorder.RecordRefundOnCancellation(ctx, reservation)
}

func (s *Service) notifyReservationCancelled(ctx context.Context, reservation *ReservationDetail, refund *CancelRefund, noRefundByDeadline bool) {
	if reservation == nil {
		return
	}

	cabinName := s.cabinDisplayName(ctx, reservation.CabinID)
	period := fmt.Sprintf("del %s al %s", reservation.StartDate, reservation.EndDate)

	notice := ""
	switch {
	case refund != nil:
		notice = fmt.Sprintf(" Se registró un reembolso %s de %s.", refund.Kind, formatMoney(refund.Amount))
	case noRefundByDeadline:
		notice = fmt.Sprintf(" Al cancelarse con menos de %d días de anticipación no corresponde reembolso.", MinCancellationDays)
	}

	s.notify(ctx, reservation.UserID, notifications.TipoReservaCancelada,
		fmt.Sprintf("Tu reserva en %s %s fue cancelada.%s", cabinName, period, notice),
		notifications.EntidadReservacion, reservation.ID)

	if reservation.HostID != reservation.UserID {
		s.notify(ctx, reservation.HostID, notifications.TipoReservaCancelada,
			fmt.Sprintf("La reserva de tu cabaña %s %s fue cancelada.%s", cabinName, period, notice),
			notifications.EntidadReservacion, reservation.ID)
	}

	s.notifyAdmins(ctx, notifications.TipoReservaCancelada,
		fmt.Sprintf("Se canceló la reserva #%d en %s %s.%s", reservation.ID, cabinName, period, notice),
		notifications.EntidadReservacion, reservation.ID)
}

func (s *Service) PayAnticipo(ctx context.Context, id int, amount float64, method string) (*ReservationDetail, float64, error) {
	if id <= 0 {
		return nil, 0, errors.New("ID de reservación inválido")
	}

	reservation, err := s.repository.GetReservationByID(ctx, id)
	if err != nil {
		return nil, 0, err
	}

	if reservation.Status != StatusPending {
		return nil, 0, ErrReservationNotPending
	}

	var rate float64
	if reservation.NightlyRate != nil {
		rate = *reservation.NightlyRate
	} else {
		rate, err = s.repository.GetCabinNightlyRate(ctx, reservation.CabinID)
		if err != nil {
			return nil, 0, err
		}
	}

	nights, err := calculateReservationNights(reservation.StartDate, reservation.EndDate)
	if err != nil {
		return nil, 0, err
	}

	total := float64(nights) * rate
	if reservation.TotalPrice != nil {
		total = *reservation.TotalPrice
	}
	minimum := math.Round(total*AnticipoPercent*100) / 100

	if method == "" {
		method = PaymentMethodCash
	}
	if !IsValidPaymentMethod(method) {
		return nil, 0, errors.New("Método de pago inválido: usa tarjeta, transferencia o efectivo")
	}

	if amount <= 0 {
		amount = minimum
	}
	amount = math.Round(amount*100) / 100
	if amount < minimum {
		return nil, 0, errors.New("El anticipo debe ser al menos el 30% del total")
	}
	if amount > total {
		return nil, 0, errors.New("El monto no puede superar el total de la reservación")
	}

	if err := s.repository.PayAnticipo(ctx, id, reservation.UserID, amount, method); err != nil {
		if err.Error() == "La cabaña ya fue confirmada para el rango de fechas seleccionado" {
			return nil, 0, errors.New("La cabaña no está disponible en el rango de fechas seleccionado")
		}
		return nil, 0, err
	}

	s.notifyReservationConfirmed(ctx, reservation)

	updated, err := s.repository.GetReservationByID(ctx, id)
	if err != nil {
		return reservation, amount, nil
	}

	return updated, amount, nil
}

func (s *Service) notifyReservationConfirmed(ctx context.Context, reservation *ReservationDetail) {
	if reservation == nil {
		return
	}

	cabinName := s.cabinDisplayName(ctx, reservation.CabinID)
	period := fmt.Sprintf("del %s al %s", reservation.StartDate, reservation.EndDate)

	s.notify(ctx, reservation.UserID, notifications.TipoReservaConfirmada,
		fmt.Sprintf("Tu reserva en %s %s fue confirmada con el pago del anticipo.", cabinName, period),
		notifications.EntidadReservacion, reservation.ID)

	if reservation.HostID != reservation.UserID {
		s.notify(ctx, reservation.HostID, notifications.TipoReservaConfirmada,
			fmt.Sprintf("La reserva de tu cabaña %s %s fue confirmada con el pago del anticipo.", cabinName, period),
			notifications.EntidadReservacion, reservation.ID)
	}

	s.notifyAdmins(ctx, notifications.TipoReservaConfirmada,
		fmt.Sprintf("Se confirmó con anticipo la reserva #%d en %s %s.", reservation.ID, cabinName, period),
		notifications.EntidadReservacion, reservation.ID)
}

func (s *Service) GetReservationsByUserID(ctx context.Context, userID uint) ([]ReservationCard, error) {
	if userID == 0 {
		return nil, errors.New("El usuario es requerido")
	}

	return s.repository.GetReservationsByUserID(ctx, userID)
}

func (s *Service) GetReservationsByCabinID(ctx context.Context, cabinID uint) ([]ReservationCard, error) {
	if cabinID == 0 {
		return nil, errors.New("La cabaña es requerida")
	}

	return s.repository.GetReservationsByCabinID(ctx, cabinID)
}

func (s *Service) GetCabinHostID(ctx context.Context, cabinID uint) (uint, error) {
	if cabinID == 0 {
		return 0, errors.New("La cabaña es requerida")
	}

	return s.repository.GetCabinHostID(ctx, cabinID)
}

func (s *Service) GetCabinAvailability(ctx context.Context, cabinID uint) (*CabinAvailabilityResponse, error) {
	if cabinID == 0 {
		return nil, errors.New("La cabaña es requerida")
	}

	ranges, err := s.repository.GetBlockedDateRanges(ctx, cabinID)
	if err != nil {
		return nil, err
	}
	if ranges == nil {
		ranges = []DateRange{}
	}

	return &CabinAvailabilityResponse{
		CabinID:       cabinID,
		BlockedRanges: ranges,
	}, nil
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
	if existing.Status != StatusPending && existing.Status != StatusConfirmed {
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

func (s *Service) AdvanceReservationStates(ctx context.Context) (int, error) {
	ids, err := s.repository.AdvanceReservationStates(ctx)
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
