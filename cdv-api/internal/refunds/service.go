package refunds

import (
	"context"
	"math"
	"time"

	"cdv-api/internal/reservations"
)

type RefundRepository interface {
	GetSummary(ctx context.Context, reservationID int) (totalPaid, totalRefunded float64, err error)
	ListMovements(ctx context.Context, reservationID int) ([]FinancialMovement, error)
	InsertRefund(ctx context.Context, reservationID int, hostID uint, amount float64) (*Refund, error)
}

type Service struct {
	repository RefundRepository
}

func NewService(repository RefundRepository) *Service {
	return &Service{repository: repository}
}

const dateLayout = "2006-01-02"

const refundTimezone = "America/Guatemala"

func refundLocation() *time.Location {
	loc, err := time.LoadLocation(refundTimezone)
	if err != nil {
		return time.FixedZone("GT", -6*60*60)
	}
	return loc
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func RefundPercentForCancellation(startDate string, now time.Time) float64 {
	loc := refundLocation()
	now = now.In(loc)

	start, err := time.ParseInLocation(dateLayout, startDate, loc)
	if err != nil {
		return RefundPercentNone
	}

	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, loc)

	days := int(math.Round(start.Sub(today).Hours() / 24))
	switch {
	case days >= RefundFullDays:
		return RefundPercentFull
	case days >= RefundPartialDays:
		return RefundPercentPartial
	default:
		return RefundPercentNone
	}
}

func (s *Service) RecordRefundOnCancellation(ctx context.Context, reservation *reservations.ReservationDetail) (*reservations.CancelRefund, error) {
	if reservation == nil {
		return nil, nil
	}

	totalPaid, totalRefunded, err := s.repository.GetSummary(ctx, reservation.ID)
	if err != nil {
		return nil, err
	}

	available := round2(totalPaid - totalRefunded)
	if available <= 0 {
		return nil, nil
	}

	percent := RefundPercentForCancellation(reservation.StartDate, time.Now())
	amount := round2(available * percent)
	if amount <= 0 {
		return nil, nil
	}

	if _, err := s.repository.InsertRefund(ctx, reservation.ID, reservation.HostID, amount); err != nil {
		return nil, err
	}

	kind := KindParcial
	if percent >= RefundPercentFull {
		kind = KindTotal
	}

	return &reservations.CancelRefund{
		Amount:  amount,
		Percent: percent,
		Kind:    kind,
	}, nil
}

func (s *Service) CreateRefund(ctx context.Context, reservation *reservations.ReservationDetail, req CreateRefundRequest) (*Refund, error) {
	if reservation == nil {
		return nil, ErrReservationNotFound
	}
	if reservation.Status != reservations.StatusCancelled &&
		reservation.Status != reservations.StatusConfirmed &&
		reservation.Status != reservations.StatusInProgress &&
		reservation.Status != reservations.StatusFinished {
		return nil, ErrReservationNotRefundable
	}

	totalPaid, totalRefunded, err := s.repository.GetSummary(ctx, reservation.ID)
	if err != nil {
		return nil, err
	}

	available := round2(totalPaid - totalRefunded)
	if available <= 0 {
		return nil, ErrNothingRefundable
	}

	var amount float64
	if req.Amount == nil {
		amount = available
	} else {
		amount = round2(*req.Amount)
		if amount <= 0 {
			return nil, ErrRefundInvalidAmount
		}
		if amount > available {
			return nil, ErrRefundExceedsAvailable
		}
	}

	return s.repository.InsertRefund(ctx, reservation.ID, reservation.HostID, amount)
}

func (s *Service) GetHistory(ctx context.Context, reservationID int) (*RefundsHistoryResponse, error) {
	totalPaid, totalRefunded, err := s.repository.GetSummary(ctx, reservationID)
	if err != nil {
		return nil, err
	}

	movements, err := s.repository.ListMovements(ctx, reservationID)
	if err != nil {
		return nil, err
	}
	if movements == nil {
		movements = []FinancialMovement{}
	}

	totalPaid = round2(totalPaid)
	totalRefunded = round2(totalRefunded)

	available := round2(totalPaid - totalRefunded)
	if available < 0 {
		available = 0
	}

	return &RefundsHistoryResponse{
		Movements:           movements,
		TotalPaid:           totalPaid,
		TotalRefunded:       totalRefunded,
		RefundableAvailable: available,
	}, nil
}
