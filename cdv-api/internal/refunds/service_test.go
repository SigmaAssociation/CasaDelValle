package refunds

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"cdv-api/internal/reservations"
)

type insertCall struct {
	reservationID int
	hostID        uint
	amount        float64
}

type mockRepository struct {
	summaryPaid      float64
	summaryRefunded  float64
	summaryErr       error
	movements        []FinancialMovement
	movementsErr     error
	insertErr        error
	insertCalledWith *insertCall
}

func (m *mockRepository) GetSummary(ctx context.Context, reservationID int) (float64, float64, error) {
	return m.summaryPaid, m.summaryRefunded, m.summaryErr
}

func (m *mockRepository) ListMovements(ctx context.Context, reservationID int) ([]FinancialMovement, error) {
	return m.movements, m.movementsErr
}

func (m *mockRepository) InsertRefund(ctx context.Context, reservationID int, hostID uint, amount float64) (*Refund, error) {
	if m.insertErr != nil {
		return nil, m.insertErr
	}
	m.insertCalledWith = &insertCall{reservationID: reservationID, hostID: hostID, amount: amount}
	return &Refund{
		ID:            1,
		ReservationID: reservationID,
		HostID:        hostID,
		Amount:        amount,
		Method:        RefundMethod,
		CreatedAt:     time.Now(),
	}, nil
}

func futureDate(days int) string {
	return time.Now().AddDate(0, 0, days).Format("2006-01-02")
}

func cancelledReservation(startsInDays int) *reservations.ReservationDetail {
	return &reservations.ReservationDetail{
		ID:        5,
		UserID:    1,
		HostID:    2,
		CabinID:   3,
		StartDate: futureDate(startsInDays),
		EndDate:   futureDate(startsInDays + 2),
		Status:    reservations.StatusCancelled,
	}
}

func TestRefundPercentForCancellation(t *testing.T) {
	now := time.Now()
	cases := []struct {
		startsInDays int
		wantPercent  float64
	}{
		{startsInDays: 30, wantPercent: RefundPercentFull},
		{startsInDays: 7, wantPercent: RefundPercentFull},
		{startsInDays: 6, wantPercent: RefundPercentPartial},
		{startsInDays: 3, wantPercent: RefundPercentPartial},
		{startsInDays: 2, wantPercent: RefundPercentNone},
		{startsInDays: 0, wantPercent: RefundPercentNone},
	}

	for _, tc := range cases {
		got := RefundPercentForCancellation(futureDate(tc.startsInDays), now)
		if got != tc.wantPercent {
			t.Errorf("inicio en %d días: porcentaje = %v, se esperaba %v", tc.startsInDays, got, tc.wantPercent)
		}
	}
}

func TestRecordRefundOnCancellationFullRefund(t *testing.T) {
	repo := &mockRepository{summaryPaid: 300}
	service := NewService(repo)

	refund, err := service.RecordRefundOnCancellation(context.Background(), cancelledReservation(10))
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if refund == nil {
		t.Fatal("se esperaba un reembolso registrado")
	}
	if refund.Amount != 300 {
		t.Errorf("monto = %v, se esperaba 300", refund.Amount)
	}
	if refund.Kind != KindTotal {
		t.Errorf("tipo = %q, se esperaba %q", refund.Kind, KindTotal)
	}
	if repo.insertCalledWith == nil || repo.insertCalledWith.amount != 300 {
		t.Errorf("no se insertó el reembolso por 300")
	}
	if repo.insertCalledWith.hostID != 2 {
		t.Errorf("el reembolso debe registrarse a nombre del anfitrión (hostID=2), se obtuvo %d", repo.insertCalledWith.hostID)
	}
}

func TestRecordRefundOnCancellationPartialRefund(t *testing.T) {
	repo := &mockRepository{summaryPaid: 300}
	service := NewService(repo)

	refund, err := service.RecordRefundOnCancellation(context.Background(), cancelledReservation(5))
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if refund == nil {
		t.Fatal("se esperaba un reembolso registrado")
	}
	if refund.Amount != 210 {
		t.Errorf("monto = %v, se esperaba 210 (70%% de 300)", refund.Amount)
	}
	if refund.Kind != KindParcial {
		t.Errorf("tipo = %q, se esperaba %q", refund.Kind, KindParcial)
	}
	if refund.Percent != RefundPercentPartial {
		t.Errorf("porcentaje = %v, se esperaba %v", refund.Percent, RefundPercentPartial)
	}
}

func TestRecordRefundOnCancellationRespectsAlreadyRefunded(t *testing.T) {
	repo := &mockRepository{summaryPaid: 300, summaryRefunded: 100}
	service := NewService(repo)

	refund, err := service.RecordRefundOnCancellation(context.Background(), cancelledReservation(10))
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if refund == nil || refund.Amount != 200 {
		t.Errorf("monto = %v, se esperaba 200 (saldo disponible)", refund)
	}
}

func TestRecordRefundOnCancellationSkipsWhenNothingPaid(t *testing.T) {
	repo := &mockRepository{}
	service := NewService(repo)

	refund, err := service.RecordRefundOnCancellation(context.Background(), cancelledReservation(10))
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if refund != nil {
		t.Errorf("no debería registrarse reembolso sin pagos, se obtuvo %v", refund)
	}
	if repo.insertCalledWith != nil {
		t.Errorf("no debería insertarse ningún movimiento")
	}
}

func TestRecordRefundOnCancellationRoundsToCents(t *testing.T) {
	repo := &mockRepository{summaryPaid: 99.99}
	service := NewService(repo)

	refund, err := service.RecordRefundOnCancellation(context.Background(), cancelledReservation(4))
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if refund == nil || math.Abs(refund.Amount-69.99) > 0.001 {
		t.Errorf("monto = %v, se esperaba 69.99", refund)
	}
}

func TestRecordRefundOnCancellationSkipsWhenLessThanMinDays(t *testing.T) {
	repo := &mockRepository{summaryPaid: 300}
	service := NewService(repo)

	refund, err := service.RecordRefundOnCancellation(context.Background(), cancelledReservation(1))
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if refund != nil {
		t.Errorf("no debería registrarse reembolso con menos de %d días, se obtuvo %v", RefundPartialDays, refund)
	}
	if repo.insertCalledWith != nil {
		t.Errorf("no debería insertarse ningún movimiento")
	}
}

func TestCreateRefundRejectsNonCancelledReservation(t *testing.T) {
	service := NewService(&mockRepository{summaryPaid: 300})

	reservation := cancelledReservation(5)
	reservation.Status = reservations.StatusConfirmed

	amount := 100.0
	_, err := service.CreateRefund(context.Background(), reservation, CreateRefundRequest{Amount: &amount})
	if !errors.Is(err, ErrReservationNotRefundable) {
		t.Errorf("error = %v, se esperaba ErrReservationNotRefundable", err)
	}
}

func TestCreateRefundRejectsExceedingAvailable(t *testing.T) {
	repo := &mockRepository{summaryPaid: 300, summaryRefunded: 100}
	service := NewService(repo)

	amount := 200.01
	_, err := service.CreateRefund(context.Background(), cancelledReservation(5), CreateRefundRequest{Amount: &amount})
	if !errors.Is(err, ErrRefundExceedsAvailable) {
		t.Errorf("error = %v, se esperaba ErrRefundExceedsAvailable", err)
	}
	if repo.insertCalledWith != nil {
		t.Errorf("no debería insertarse ningún movimiento")
	}
}

func TestCreateRefundAcceptsPartialWithinAvailable(t *testing.T) {
	repo := &mockRepository{summaryPaid: 300, summaryRefunded: 100}
	service := NewService(repo)

	amount := 150.55
	refund, err := service.CreateRefund(context.Background(), cancelledReservation(5), CreateRefundRequest{Amount: &amount})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if refund.Amount != 150.55 {
		t.Errorf("monto = %v, se esperaba 150.55", refund.Amount)
	}
}

func TestCreateRefundRejectsNonPositiveAmount(t *testing.T) {
	service := NewService(&mockRepository{summaryPaid: 300})

	for _, amount := range []float64{0, -10} {
		value := amount
		_, err := service.CreateRefund(context.Background(), cancelledReservation(5), CreateRefundRequest{Amount: &value})
		if !errors.Is(err, ErrRefundInvalidAmount) {
			t.Errorf("monto %v: error = %v, se esperaba ErrRefundInvalidAmount", amount, err)
		}
	}
}

func TestCreateRefundWithoutAmountRefundsFullAvailable(t *testing.T) {
	repo := &mockRepository{summaryPaid: 300, summaryRefunded: 120}
	service := NewService(repo)

	refund, err := service.CreateRefund(context.Background(), cancelledReservation(5), CreateRefundRequest{})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if refund.Amount != 180 {
		t.Errorf("monto = %v, se esperaba 180 (total disponible)", refund.Amount)
	}
}

func TestCreateRefundRejectsWhenNothingRefundable(t *testing.T) {
	service := NewService(&mockRepository{summaryPaid: 300, summaryRefunded: 300})

	_, err := service.CreateRefund(context.Background(), cancelledReservation(5), CreateRefundRequest{})
	if !errors.Is(err, ErrNothingRefundable) {
		t.Errorf("error = %v, se esperaba ErrNothingRefundable", err)
	}
}

func TestGetHistoryReturnsTotalsAndMovements(t *testing.T) {
	repo := &mockRepository{
		summaryPaid:     300,
		summaryRefunded: 75,
		movements: []FinancialMovement{
			{ID: 1, Type: PaymentTypePago, Subtype: PaymentTypeAnticipo, Amount: 90, Method: "CASH"},
			{ID: 2, Type: RefundType, Subtype: RefundType, Amount: 75, Method: RefundMethod},
		},
	}
	service := NewService(repo)

	history, err := service.GetHistory(context.Background(), 5)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(history.Movements) != 2 {
		t.Errorf("movimientos = %d, se esperaban 2", len(history.Movements))
	}
	if history.TotalPaid != 300 {
		t.Errorf("total pagado = %v, se esperaba 300", history.TotalPaid)
	}
	if history.TotalRefunded != 75 {
		t.Errorf("total reembolsado = %v, se esperaba 75", history.TotalRefunded)
	}
	if history.RefundableAvailable != 225 {
		t.Errorf("disponible = %v, se esperaba 225", history.RefundableAvailable)
	}
}

func TestGetHistoryReturnsEmptyMovementList(t *testing.T) {
	service := NewService(&mockRepository{})

	history, err := service.GetHistory(context.Background(), 5)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if history.Movements == nil || len(history.Movements) != 0 {
		t.Errorf("se esperaba una lista de movimientos vacía, no nil")
	}
}
