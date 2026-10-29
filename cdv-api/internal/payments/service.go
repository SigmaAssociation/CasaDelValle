package payments

import (
	"context"
	"fmt"
	"math"
	"time"

	"cdv-api/internal/notifications"
)

type Service struct {
	repository PaymentRepository
	notifier   Notifier
}

// Notifier registra notificaciones en el buzón personal de los usuarios.
// Interfaz declarada localmente para desacoplar el módulo de notifications y
// poder sustituirla en pruebas.
type Notifier interface {
	Notify(ctx context.Context, input notifications.NotificationInput)
	NotifyAdmins(ctx context.Context, input notifications.NotificationInput)
}

type PaymentRepository interface {
	GetReservationInfo(ctx context.Context, reservationID int) (*ReservationPaymentInfo, error)
	GetPaidAmount(ctx context.Context, reservationID int) (float64, error)
	ListByReservation(ctx context.Context, reservationID int) ([]Payment, error)
	Create(ctx context.Context, reservationID int, payerID uint, amount float64, method string) (*Payment, error)
}

func NewService(repository PaymentRepository, notifier Notifier) *Service {
	return &Service{
		repository: repository,
		notifier:   notifier,
	}
}

const dateLayout = "2006-01-02"

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func calculateNights(startDate, endDate string) (int, error) {
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

// GetReservationInfo expone los datos que el controller necesita para
// autorizar el acceso a los pagos de una reservación.
func (s *Service) GetReservationInfo(ctx context.Context, reservationID int) (*ReservationPaymentInfo, error) {
	if reservationID <= 0 {
		return nil, ErrReservationNotFound
	}
	return s.repository.GetReservationInfo(ctx, reservationID)
}

// GetReservationPayments devuelve el historial de pagos y el estado del saldo.
func (s *Service) GetReservationPayments(ctx context.Context, reservationID int) (*ReservationPayments, error) {
	info, err := s.repository.GetReservationInfo(ctx, reservationID)
	if err != nil {
		return nil, err
	}
	return s.buildSummary(ctx, info)
}

// CreatePayment registra un pago/abono simulado sobre una reservación
// confirmada o en curso. Valida que el monto sea mayor que 0 y que no exceda
// el saldo pendiente (total menos pagos ya registrados).
func (s *Service) CreatePayment(ctx context.Context, reservationID int, actorID uint, req CreatePaymentRequest) (*CreatePaymentResponse, error) {
	if reservationID <= 0 {
		return nil, ErrReservationNotFound
	}

	amount := round2(req.Amount)
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if !payableMethods[req.Method] {
		return nil, ErrInvalidMethod
	}

	info, err := s.repository.GetReservationInfo(ctx, reservationID)
	if err != nil {
		return nil, err
	}

	if info.Status != statusConfirmed && info.Status != statusInProgress {
		return nil, ErrReservationNotPayable
	}

	nights, err := calculateNights(info.StartDate, info.EndDate)
	if err != nil {
		return nil, err
	}
	total := round2(float64(nights) * info.NightlyRate)
	if info.TotalPrice != nil {
		total = round2(*info.TotalPrice)
	}

	paidRaw, err := s.repository.GetPaidAmount(ctx, reservationID)
	if err != nil {
		return nil, err
	}
	paid := round2(paidRaw)
	balance := round2(total - paid)

	if balance <= 0 || amount > balance {
		return nil, ErrAmountExceedsBalance
	}

	payment, err := s.repository.Create(ctx, reservationID, actorID, amount, req.Method)
	if err != nil {
		return nil, err
	}

	s.notifyPaymentRecorded(ctx, info, payment)

	summary, err := s.buildSummary(ctx, info)
	if err != nil {
		// El pago ya quedó registrado; no fallar solo por la relectura.
		return &CreatePaymentResponse{
			Message:       "Pago registrado exitosamente",
			Payment:       payment,
			ReservationID: reservationID,
			Total:         total,
			Paid:          round2(paid + amount),
			Balance:       round2(balance - amount),
			Payments:      []Payment{*payment},
		}, nil
	}

	return &CreatePaymentResponse{
		Message:       "Pago registrado exitosamente",
		Payment:       payment,
		ReservationID: summary.ReservationID,
		Total:         summary.Total,
		Paid:          summary.Paid,
		Balance:       summary.Balance,
		Payments:      summary.Payments,
	}, nil
}

// buildSummary calcula total, pagado y saldo, y adjunta el historial de pagos.
func (s *Service) buildSummary(ctx context.Context, info *ReservationPaymentInfo) (*ReservationPayments, error) {
	nights, err := calculateNights(info.StartDate, info.EndDate)
	if err != nil {
		return nil, err
	}
	total := round2(float64(nights) * info.NightlyRate)
	if info.TotalPrice != nil {
		total = round2(*info.TotalPrice)
	}

	paid, err := s.repository.GetPaidAmount(ctx, info.ReservationID)
	if err != nil {
		return nil, err
	}
	paid = round2(paid)

	list, err := s.repository.ListByReservation(ctx, info.ReservationID)
	if err != nil {
		return nil, err
	}

	return &ReservationPayments{
		ReservationID: info.ReservationID,
		Total:         total,
		Paid:          paid,
		Balance:       round2(total - paid),
		Payments:      list,
	}, nil
}

// notifyPaymentRecorded avisa al huésped, al anfitrión y a los administradores
// del pago registrado.
func (s *Service) notifyPaymentRecorded(ctx context.Context, info *ReservationPaymentInfo, payment *Payment) {
	if info == nil || payment == nil {
		return
	}

	s.notify(ctx, info.UserID, notifications.TipoPagoRegistrado,
		fmt.Sprintf("Se registró un pago de Q%.2f (%s) para tu reserva en %s.", payment.Amount, payment.Method, info.CabinName),
		notifications.EntidadReservacion, info.ReservationID)

	if info.HostID != info.UserID {
		s.notify(ctx, info.HostID, notifications.TipoPagoRegistrado,
			fmt.Sprintf("Se registró un pago de Q%.2f (%s) en tu cabaña %s (reserva #%d).", payment.Amount, payment.Method, info.CabinName, info.ReservationID),
			notifications.EntidadReservacion, info.ReservationID)
	}

	s.notifyAdmins(ctx, notifications.TipoPagoRegistrado,
		fmt.Sprintf("Pago de Q%.2f (%s) registrado en la reserva #%d en %s.", payment.Amount, payment.Method, info.ReservationID, info.CabinName),
		notifications.EntidadReservacion, info.ReservationID)
}

// notify registra una notificación sin interrumpir la operación principal.
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
