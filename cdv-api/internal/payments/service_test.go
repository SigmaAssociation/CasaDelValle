package payments

import (
	"context"
	"errors"
	"testing"
	"time"

	"cdv-api/internal/notifications"
)

type createCall struct {
	reservationID int
	payerID       uint
	amount        float64
	method        string
}

type mockPaymentRepository struct {
	info      *ReservationPaymentInfo
	infoErr   error
	paid      float64
	paidErr   error
	list      []Payment
	listErr   error
	created   *Payment
	createErr error
	calls     []createCall
}

func (m *mockPaymentRepository) GetReservationInfo(ctx context.Context, reservationID int) (*ReservationPaymentInfo, error) {
	return m.info, m.infoErr
}

func (m *mockPaymentRepository) GetPaidAmount(ctx context.Context, reservationID int) (float64, error) {
	return m.paid, m.paidErr
}

func (m *mockPaymentRepository) ListByReservation(ctx context.Context, reservationID int) ([]Payment, error) {
	return m.list, m.listErr
}

func (m *mockPaymentRepository) Create(ctx context.Context, reservationID int, payerID uint, amount float64, method string) (*Payment, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	m.calls = append(m.calls, createCall{reservationID: reservationID, payerID: payerID, amount: amount, method: method})

	payment := m.created
	if payment == nil {
		payment = &Payment{ID: 1, ReservationID: reservationID, UserID: payerID, Amount: amount, Method: method, Date: time.Now()}
	}
	// Refleja el pago en el saldo para la relectura del resumen.
	m.paid += amount
	m.list = append(m.list, *payment)
	return payment, nil
}

type mockNotifier struct {
	inputs      []notifications.NotificationInput
	adminInputs []notifications.NotificationInput
}

func (m *mockNotifier) Notify(ctx context.Context, input notifications.NotificationInput) {
	m.inputs = append(m.inputs, input)
}

func (m *mockNotifier) NotifyAdmins(ctx context.Context, input notifications.NotificationInput) {
	m.adminInputs = append(m.adminInputs, input)
}

func (m *mockNotifier) byTipo(tipo string) []notifications.NotificationInput {
	var out []notifications.NotificationInput
	for _, in := range m.inputs {
		if in.Tipo == tipo {
			out = append(out, in)
		}
	}
	return out
}

func confirmedReservation() *ReservationPaymentInfo {
	return &ReservationPaymentInfo{
		ReservationID: 10,
		UserID:        5,
		HostID:        9,
		CabinID:       3,
		CabinName:     "Cabaña Luna",
		Status:        statusConfirmed,
		StartDate:     "2026-10-10",
		EndDate:       "2026-10-15", // 5 noches
		NightlyRate:   100,          // total 500
	}
}

func TestCreatePaymentRecordsValidAmount(t *testing.T) {
	repo := &mockPaymentRepository{info: confirmedReservation(), paid: 150}
	notifier := &mockNotifier{}
	service := NewService(repo, notifier)

	res, err := service.CreatePayment(context.Background(), 10, 5, CreatePaymentRequest{Amount: 100, Method: MethodTransfer})
	if err != nil {
		t.Fatalf("CreatePayment returned error: %v", err)
	}
	if len(repo.calls) != 1 {
		t.Fatalf("expected 1 repository.Create call, got %d", len(repo.calls))
	}
	if repo.calls[0].amount != 100 || repo.calls[0].method != MethodTransfer || repo.calls[0].payerID != 5 {
		t.Fatalf("unexpected create call: %+v", repo.calls[0])
	}
	if res.Total != 500 {
		t.Fatalf("total = %v, want 500", res.Total)
	}
	if res.Paid != 250 {
		t.Fatalf("paid = %v, want 250 (150 anticipo + 100 nuevo)", res.Paid)
	}
	if res.Balance != 250 {
		t.Fatalf("balance = %v, want 250", res.Balance)
	}
}

func TestCreatePaymentRejectsInvalidAmount(t *testing.T) {
	for _, amount := range []float64{0, -10} {
		repo := &mockPaymentRepository{info: confirmedReservation(), paid: 150}
		service := NewService(repo, &mockNotifier{})

		_, err := service.CreatePayment(context.Background(), 10, 5, CreatePaymentRequest{Amount: amount, Method: MethodCard})
		if !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("amount %v: err = %v, want ErrInvalidAmount", amount, err)
		}
		if len(repo.calls) != 0 {
			t.Fatalf("amount %v: repository.Create should not be called", amount)
		}
	}
}

func TestCreatePaymentRejectsInvalidMethod(t *testing.T) {
	repo := &mockPaymentRepository{info: confirmedReservation(), paid: 150}
	service := NewService(repo, &mockNotifier{})

	_, err := service.CreatePayment(context.Background(), 10, 5, CreatePaymentRequest{Amount: 100, Method: "bitcoin"})
	if !errors.Is(err, ErrInvalidMethod) {
		t.Fatalf("err = %v, want ErrInvalidMethod", err)
	}
	if len(repo.calls) != 0 {
		t.Fatal("repository.Create should not be called for an invalid method")
	}
}

func TestCreatePaymentRejectsAmountExceedingBalance(t *testing.T) {
	// total 500, pagado 150 anticipo -> saldo 350. Intentar 400.
	repo := &mockPaymentRepository{info: confirmedReservation(), paid: 150}
	service := NewService(repo, &mockNotifier{})

	_, err := service.CreatePayment(context.Background(), 10, 5, CreatePaymentRequest{Amount: 400, Method: MethodCash})
	if !errors.Is(err, ErrAmountExceedsBalance) {
		t.Fatalf("err = %v, want ErrAmountExceedsBalance", err)
	}
	if len(repo.calls) != 0 {
		t.Fatal("repository.Create should not be called when the amount exceeds the balance")
	}
}

func TestCreatePaymentRejectsWhenBalanceIsZero(t *testing.T) {
	repo := &mockPaymentRepository{info: confirmedReservation(), paid: 500}
	service := NewService(repo, &mockNotifier{})

	_, err := service.CreatePayment(context.Background(), 10, 5, CreatePaymentRequest{Amount: 50, Method: MethodCash})
	if !errors.Is(err, ErrAmountExceedsBalance) {
		t.Fatalf("err = %v, want ErrAmountExceedsBalance", err)
	}
}

func TestCreatePaymentRejectsNonPayableState(t *testing.T) {
	for _, status := range []string{"pendiente", "finalizada", "cancelada"} {
		info := confirmedReservation()
		info.Status = status
		repo := &mockPaymentRepository{info: info, paid: 0}
		service := NewService(repo, &mockNotifier{})

		_, err := service.CreatePayment(context.Background(), 10, 5, CreatePaymentRequest{Amount: 50, Method: MethodCash})
		if !errors.Is(err, ErrReservationNotPayable) {
			t.Fatalf("status %q: err = %v, want ErrReservationNotPayable", status, err)
		}
		if len(repo.calls) != 0 {
			t.Fatalf("status %q: repository.Create should not be called", status)
		}
	}
}

func TestCreatePaymentAllowsInProgressReservation(t *testing.T) {
	info := confirmedReservation()
	info.Status = statusInProgress
	repo := &mockPaymentRepository{info: info, paid: 150}
	service := NewService(repo, &mockNotifier{})

	if _, err := service.CreatePayment(context.Background(), 10, 5, CreatePaymentRequest{Amount: 100, Method: MethodCash}); err != nil {
		t.Fatalf("in-progress reservation should accept payments: %v", err)
	}
}

func TestCreatePaymentNotifiesGuestHostAndAdmins(t *testing.T) {
	repo := &mockPaymentRepository{info: confirmedReservation(), paid: 150}
	notifier := &mockNotifier{}
	service := NewService(repo, notifier)

	if _, err := service.CreatePayment(context.Background(), 10, 5, CreatePaymentRequest{Amount: 100, Method: MethodCard}); err != nil {
		t.Fatalf("CreatePayment returned error: %v", err)
	}

	if got := len(notifier.byTipo(notifications.TipoPagoRegistrado)); got != 2 {
		t.Fatalf("guest+host notifications = %d, want 2", got)
	}
	if got := len(notifier.adminInputs); got != 1 {
		t.Fatalf("admin notifications = %d, want 1", got)
	}
	if notifier.adminInputs[0].Tipo != notifications.TipoPagoRegistrado {
		t.Fatalf("admin notification tipo = %q", notifier.adminInputs[0].Tipo)
	}
}

func TestGetReservationPaymentsComputesBalance(t *testing.T) {
	repo := &mockPaymentRepository{
		info: confirmedReservation(),
		paid: 150,
		list: []Payment{
			{ID: 1, ReservationID: 10, Amount: 150, Method: MethodAnticipo},
		},
	}
	service := NewService(repo, &mockNotifier{})

	summary, err := service.GetReservationPayments(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetReservationPayments returned error: %v", err)
	}
	if summary.Total != 500 || summary.Paid != 150 || summary.Balance != 350 {
		t.Fatalf("summary = %+v, want total 500, paid 150, balance 350", summary)
	}
	if len(summary.Payments) != 1 {
		t.Fatalf("payments = %d, want 1", len(summary.Payments))
	}
}

func TestCreatePaymentRejectsInvalidReservationID(t *testing.T) {
	repo := &mockPaymentRepository{info: confirmedReservation(), paid: 150}
	service := NewService(repo, &mockNotifier{})

	_, err := service.CreatePayment(context.Background(), 0, 5, CreatePaymentRequest{Amount: 100, Method: MethodCard})
	if !errors.Is(err, ErrReservationNotFound) {
		t.Fatalf("err = %v, want ErrReservationNotFound", err)
	}
}

func TestPaymentMethodTranslation(t *testing.T) {
	cases := []struct {
		api string
		db  string
	}{
		{MethodCard, dbMethodCard},
		{MethodTransfer, dbMethodTransfer},
		{MethodCash, dbMethodCash},
	}
	for _, c := range cases {
		got, ok := dbMethodCode(c.api)
		if !ok || got != c.db {
			t.Fatalf("dbMethodCode(%q) = %q, %v; want %q", c.api, got, ok, c.db)
		}
		if back := apiMethod("pago", got); back != c.api {
			t.Fatalf("apiMethod(pago, %q) = %q; want %q", got, back, c.api)
		}
	}
	if apiMethod(MethodAnticipo, dbMethodCash) != MethodAnticipo {
		t.Fatal("los pagos de tipo anticipo deben exponerse como 'anticipo'")
	}
}
