package reservations

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"cdv-api/internal/notifications"
)

type payCall struct {
	reservationID int
	payerID       uint
	amount        float64
	method        string
}

type mockReservationRepository struct {
	createCalled      bool
	createReturnID    int
	createReturnErr   error
	hasOverlapReturn  bool
	hasOverlapErr     error
	hasUserOverlap    bool
	hasUserOverlapErr error
	getByIDReturn     *ReservationDetail
	getByIDErr        error
	cancelReturnErr   error
	getByUserReturn   []ReservationCard
	getByUserErr      error
	getByCabinReturn  []ReservationCard
	getByCabinErr     error
	getCabinHostID    uint
	getCabinHostErr   error
	checkinElig       *CheckinEligibility
	checkinEligErr    error
	recordCheckinErr  error
	recordTsErr       error
	recordTsCalled    bool
	cashCalledWith    *cashCall
	cashReturnErr     error
	updateReturnErr   error
	getCabinName      string
	getCabinNameErr   error
	nightlyRate       float64
	nightlyRateErr    error
	payReturnErr      error
	payCalledWith     *payCall
	advanceReturnIDs  []int
	advanceErr        error
	ledgerReturn      *ReservationLedger
	ledgerErr         error

	recordCheckoutErr    error
	recordCheckoutCalled bool
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

func (m *mockReservationRepository) CreateReservation(ctx context.Context, req ReservationRequest) (int, error) {
	m.createCalled = true
	return m.createReturnID, m.createReturnErr
}

func (m *mockReservationRepository) HasOverlap(ctx context.Context, cabinID uint, startDate, endDate string, excludeID int) (bool, error) {
	return m.hasOverlapReturn, m.hasOverlapErr
}

func (m *mockReservationRepository) HasUserOverlap(ctx context.Context, userID uint, startDate, endDate string, excludeID int) (bool, error) {
	return m.hasUserOverlap, m.hasUserOverlapErr
}

func (m *mockReservationRepository) GetBlockedDateRanges(ctx context.Context, cabinID uint) ([]DateRange, error) {
	return []DateRange{}, nil
}

func (m *mockReservationRepository) GetReservationByID(ctx context.Context, id int) (*ReservationDetail, error) {
	return m.getByIDReturn, m.getByIDErr
}

func (m *mockReservationRepository) CancelReservation(ctx context.Context, id int, cancelledAt time.Time) error {
	return m.cancelReturnErr
}

func (m *mockReservationRepository) GetReservationsByUserID(ctx context.Context, userID uint) ([]ReservationCard, error) {
	return m.getByUserReturn, m.getByUserErr
}

func (m *mockReservationRepository) GetReservationsByCabinID(ctx context.Context, cabinID uint) ([]ReservationCard, error) {
	return m.getByCabinReturn, m.getByCabinErr
}

func (m *mockReservationRepository) GetCabinHostID(ctx context.Context, cabinID uint) (uint, error) {
	return m.getCabinHostID, m.getCabinHostErr
}

func (m *mockReservationRepository) GetCheckinEligibility(ctx context.Context, id int) (*CheckinEligibility, error) {
	return m.checkinElig, m.checkinEligErr
}

func (m *mockReservationRepository) RecordCheckin(ctx context.Context, id int, checkedInAt time.Time) error {
	return m.recordCheckinErr
}

func (m *mockReservationRepository) RecordCheckinTimestamp(ctx context.Context, id int, checkedInAt time.Time) error {
	m.recordTsCalled = true
	return m.recordTsErr
}

func (m *mockReservationRepository) CheckinWithCashPayment(ctx context.Context, id int, payerID uint, cashAmount float64, checkedInAt time.Time) error {
	m.cashCalledWith = &cashCall{reservationID: id, payerID: payerID, amount: cashAmount}
	return m.cashReturnErr
}

func (m *mockReservationRepository) RecordCheckout(ctx context.Context, id int, checkedOutAt time.Time) error {
	m.recordCheckoutCalled = true
	if m.recordCheckoutErr == nil && m.getByIDReturn != nil {
		m.getByIDReturn.Status = StatusFinished
	}
	return m.recordCheckoutErr
}

func (m *mockReservationRepository) UpdateReservation(ctx context.Context, id int, req UpdateReservationRequest) error {
	return m.updateReturnErr
}

func (m *mockReservationRepository) GetCabinName(ctx context.Context, cabinID int) (string, error) {
	return m.getCabinName, m.getCabinNameErr
}

func (m *mockReservationRepository) GetReservationLedger(
	ctx context.Context,
	reservationID int,
	cabinID uint,
) (*ReservationLedger, error) {
	return m.ledgerReturn, m.ledgerErr
}

func (m *mockReservationRepository) GetCabinNightlyRate(ctx context.Context, cabinID uint) (float64, error) {
	return m.nightlyRate, m.nightlyRateErr
}

type cashCall struct {
	reservationID int
	payerID       uint
	amount        float64
}

func (m *mockReservationRepository) PayAnticipo(ctx context.Context, reservationID int, payerID uint, amount float64, method string) error {
	if m.payReturnErr == nil {
		m.payCalledWith = &payCall{reservationID: reservationID, payerID: payerID, amount: amount, method: method}
		if m.getByIDReturn != nil {
			m.getByIDReturn.Status = StatusConfirmed
		}
	}
	return m.payReturnErr
}

func (m *mockReservationRepository) AdvanceReservationStates(ctx context.Context) ([]int, error) {
	return m.advanceReturnIDs, m.advanceErr
}

func futureDate(days int) string {
	return time.Now().AddDate(0, 0, days).Format(dateLayout)
}

func TestReservationDatesOverlap(t *testing.T) {
	testCases := []struct {
		name        string
		existingS   string
		existingE   string
		newS        string
		newE        string
		expectsOver bool
	}{
		{
			name:        "checkout equals next checkin is allowed",
			existingS:   "2026-10-10",
			existingE:   "2026-10-15",
			newS:        "2026-10-15",
			newE:        "2026-10-20",
			expectsOver: false,
		},
		{
			name:        "partial overlap is rejected",
			existingS:   "2026-10-10",
			existingE:   "2026-10-15",
			newS:        "2026-10-14",
			newE:        "2026-10-18",
			expectsOver: true,
		},
		{
			name:        "contained reservation is rejected",
			existingS:   "2026-10-10",
			existingE:   "2026-10-20",
			newS:        "2026-10-12",
			newE:        "2026-10-18",
			expectsOver: true,
		},
		{
			name:        "same range is rejected",
			existingS:   "2026-10-10",
			existingE:   "2026-10-15",
			newS:        "2026-10-10",
			newE:        "2026-10-15",
			expectsOver: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			existingStart, err := time.Parse(dateLayout, tc.existingS)
			if err != nil {
				t.Fatalf("parse existing start: %v", err)
			}
			existingEnd, err := time.Parse(dateLayout, tc.existingE)
			if err != nil {
				t.Fatalf("parse existing end: %v", err)
			}
			newStart, err := time.Parse(dateLayout, tc.newS)
			if err != nil {
				t.Fatalf("parse new start: %v", err)
			}
			newEnd, err := time.Parse(dateLayout, tc.newE)
			if err != nil {
				t.Fatalf("parse new end: %v", err)
			}

			got := datesOverlap(existingStart, existingEnd, newStart, newEnd)
			if got != tc.expectsOver {
				t.Fatalf("datesOverlap() = %v, want %v", got, tc.expectsOver)
			}
		})
	}
}

func TestCalculateReservationTotalPrice(t *testing.T) {
	total, err := calculateReservationTotalPrice("2026-10-10", "2026-10-15", 120)
	if err != nil {
		t.Fatalf("calculateReservationTotalPrice returned error: %v", err)
	}
	if total != 600 {
		t.Fatalf("calculateReservationTotalPrice() = %v, want 600", total)
	}
}

func TestCalculateReservationNights(t *testing.T) {
	nights, err := calculateReservationNights("2026-10-10", "2026-10-15")
	if err != nil {
		t.Fatalf("calculateReservationNights returned error: %v", err)
	}
	if nights != 5 {
		t.Fatalf("calculateReservationNights() = %d, want 5", nights)
	}
}

func TestCreateReservationRejectsOverlap(t *testing.T) {
	repo := &mockReservationRepository{
		hasOverlapReturn: true,
		hasUserOverlap:   false,
		createReturnID:   99,
	}
	service := NewService(repo, &mockNotifier{})

	_, err := service.CreateReservation(context.Background(), ReservationRequest{
		UserID:    1,
		CabinID:   1,
		StartDate: futureDate(10),
		EndDate:   futureDate(15),
	})
	if err == nil {
		t.Fatal("expected overlap validation error")
	}
	if !strings.Contains(err.Error(), "no está disponible") {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createCalled {
		t.Fatal("repository.CreateReservation should not be called when overlap exists")
	}
}

func TestCreateReservationAllowsValidRange(t *testing.T) {
	repo := &mockReservationRepository{
		hasOverlapReturn: false,
		hasUserOverlap:   false,
		createReturnID:   123,
	}
	service := NewService(repo, &mockNotifier{})

	id, err := service.CreateReservation(context.Background(), ReservationRequest{
		UserID:    1,
		CabinID:   1,
		StartDate: futureDate(10),
		EndDate:   futureDate(15),
	})
	if err != nil {
		t.Fatalf("CreateReservation returned error: %v", err)
	}
	if id != 123 {
		t.Fatalf("CreateReservation() = %d, want 123", id)
	}
	if !repo.createCalled {
		t.Fatal("repository.CreateReservation should be called for a valid reservation")
	}
}

func TestCreateReservationRejectsUserOverlap(t *testing.T) {
	repo := &mockReservationRepository{
		hasOverlapReturn: false,
		hasUserOverlap:   true,
		createReturnID:   88,
	}
	service := NewService(repo, &mockNotifier{})

	_, err := service.CreateReservation(context.Background(), ReservationRequest{
		UserID:    1,
		CabinID:   1,
		StartDate: futureDate(10),
		EndDate:   futureDate(15),
	})
	if err == nil {
		t.Fatal("expected user overlap validation error")
	}
	if !strings.Contains(err.Error(), "ya tiene otra reservación") {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createCalled {
		t.Fatal("repository.CreateReservation should not be called when user overlap exists")
	}
}

func TestCreateReservationNotifiesGuestAndHost(t *testing.T) {
	repo := &mockReservationRepository{
		hasOverlapReturn: false,
		hasUserOverlap:   false,
		createReturnID:   55,
		getByIDReturn: &ReservationDetail{
			ID: 55, UserID: 1, HostID: 2, CabinID: 3,
			StartDate: futureDate(10), EndDate: futureDate(15), Status: StatusPending,
		},
		getCabinName: "Cabaña Lago",
	}
	notifier := &mockNotifier{}
	service := NewService(repo, notifier)

	_, err := service.CreateReservation(context.Background(), ReservationRequest{
		UserID:    1,
		CabinID:   3,
		StartDate: futureDate(10),
		EndDate:   futureDate(15),
	})
	if err != nil {
		t.Fatalf("CreateReservation returned error: %v", err)
	}

	created := notifier.byTipo(notifications.TipoReservaCreada)
	if len(created) != 2 {
		t.Fatalf("expected 2 creation notifications (guest + host), got %d", len(created))
	}
	if created[0].UserID != 1 || created[1].UserID != 2 {
		t.Fatalf("notifications should target guest 1 and host 2, got %d and %d", created[0].UserID, created[1].UserID)
	}
	if !strings.Contains(created[0].Mensaje, "Cabaña Lago") {
		t.Fatalf("guest message should include cabin name, got %q", created[0].Mensaje)
	}
}

func TestCreateReservationSkipsDuplicateWhenHostIsGuest(t *testing.T) {
	repo := &mockReservationRepository{
		hasOverlapReturn: false,
		hasUserOverlap:   false,
		createReturnID:   56,
		getByIDReturn: &ReservationDetail{
			ID: 56, UserID: 7, HostID: 7, CabinID: 3,
			StartDate: futureDate(10), EndDate: futureDate(15), Status: StatusPending,
		},
		getCabinName: "Cabaña Sol",
	}
	notifier := &mockNotifier{}
	service := NewService(repo, notifier)

	if _, err := service.CreateReservation(context.Background(), ReservationRequest{
		UserID:    7,
		CabinID:   3,
		StartDate: futureDate(10),
		EndDate:   futureDate(15),
	}); err != nil {
		t.Fatalf("CreateReservation returned error: %v", err)
	}

	created := notifier.byTipo(notifications.TipoReservaCreada)
	if len(created) != 1 {
		t.Fatalf("expected a single notification when host is the guest, got %d", len(created))
	}
}

func TestCancelReservationNotifiesGuestAndHost(t *testing.T) {
	repo := &mockReservationRepository{
		getByIDReturn: &ReservationDetail{
			ID: 77, UserID: 1, HostID: 2, CabinID: 3,
			StartDate: futureDate(10), EndDate: futureDate(15), Status: StatusPending,
		},
		getCabinName: "Cabaña Río",
	}
	notifier := &mockNotifier{}
	service := NewService(repo, notifier)

	if _, err := service.CancelReservation(context.Background(), 77); err != nil {
		t.Fatalf("CancelReservation returned error: %v", err)
	}

	cancelled := notifier.byTipo(notifications.TipoReservaCancelada)
	if len(cancelled) != 2 {
		t.Fatalf("expected 2 cancellation notifications (guest + host), got %d", len(cancelled))
	}
}

type mockRefundRecorder struct {
	calledWith   *ReservationDetail
	returnRefund *CancelRefund
	returnErr    error
}

func (m *mockRefundRecorder) RecordRefundOnCancellation(ctx context.Context, reservation *ReservationDetail) (*CancelRefund, error) {
	m.calledWith = reservation
	return m.returnRefund, m.returnErr
}

func TestCancelReservationWithRefundRecordsAutomaticRefund(t *testing.T) {
	repo := &mockReservationRepository{
		getByIDReturn: &ReservationDetail{
			ID: 77, UserID: 1, HostID: 2, CabinID: 3,
			StartDate: futureDate(10), EndDate: futureDate(15), Status: StatusConfirmed,
		},
		getCabinName: "Cabaña Río",
	}
	recorder := &mockRefundRecorder{
		returnRefund: &CancelRefund{Amount: 90, Percent: 1, Kind: "total"},
	}
	service := NewService(repo, &mockNotifier{})
	service.SetRefundRecorder(recorder)

	_, refund, err := service.CancelReservationWithRefund(context.Background(), 77)
	if err != nil {
		t.Fatalf("CancelReservationWithRefund returned error: %v", err)
	}
	if recorder.calledWith == nil || recorder.calledWith.ID != 77 {
		t.Fatalf("expected the refund recorder to be called with reservation 77, got %+v", recorder.calledWith)
	}
	if refund == nil || refund.Amount != 90 || refund.Kind != "total" {
		t.Fatalf("expected the automatic refund to be returned, got %+v", refund)
	}
}

func TestCancelReservationWithRefundSurvivesRecorderFailure(t *testing.T) {
	repo := &mockReservationRepository{
		getByIDReturn: &ReservationDetail{
			ID: 78, UserID: 1, HostID: 2, CabinID: 3,
			StartDate: futureDate(10), EndDate: futureDate(15), Status: StatusConfirmed,
		},
		getCabinName: "Cabaña Río",
	}
	recorder := &mockRefundRecorder{returnErr: errors.New("fallo al registrar reembolso")}
	service := NewService(repo, &mockNotifier{})
	service.SetRefundRecorder(recorder)

	_, refund, err := service.CancelReservationWithRefund(context.Background(), 78)
	if err != nil {
		t.Fatalf("un fallo al registrar el reembolso no debe revertir la cancelación: %v", err)
	}
	if refund != nil {
		t.Fatalf("no debería devolverse un reembolso cuando el registro falla, got %+v", refund)
	}
}

func TestCancelReservationWithoutRecorderDoesNotFail(t *testing.T) {
	repo := &mockReservationRepository{
		getByIDReturn: &ReservationDetail{
			ID: 79, UserID: 1, HostID: 2, CabinID: 3,
			StartDate: futureDate(10), EndDate: futureDate(15), Status: StatusConfirmed,
		},
		getCabinName: "Cabaña Río",
	}
	service := NewService(repo, &mockNotifier{})

	_, refund, err := service.CancelReservationWithRefund(context.Background(), 79)
	if err != nil {
		t.Fatalf("CancelReservationWithRefund returned error: %v", err)
	}
	if refund != nil {
		t.Fatalf("sin registrador inyectado no debería haber reembolso, got %+v", refund)
	}
}

func TestCancelReservationWithinMinDaysSucceeds(t *testing.T) {
	repo := &mockReservationRepository{
		getByIDReturn: &ReservationDetail{
			ID: 80, UserID: 1, HostID: 2, CabinID: 3,
			StartDate: futureDate(1), EndDate: futureDate(3), Status: StatusConfirmed,
		},
		getCabinName: "Cabaña Río",
	}
	notifier := &mockNotifier{}
	recorder := &mockRefundRecorder{}
	service := NewService(repo, notifier)
	service.SetRefundRecorder(recorder)

	if _, _, err := service.CancelReservationWithRefund(context.Background(), 80); err != nil {
		t.Fatalf("cancelar con menos de %d días debería estar permitido: %v", MinCancellationDays, err)
	}

	cancelled := notifier.byTipo(notifications.TipoReservaCancelada)
	if len(cancelled) != 2 {
		t.Fatalf("expected 2 cancellation notifications (guest + host), got %d", len(cancelled))
	}
	if !strings.Contains(cancelled[0].Mensaje, "no corresponde reembolso") {
		t.Errorf("la notificación al huésped debería explicar por qué no hay reembolso, got %q", cancelled[0].Mensaje)
	}
}

func TestCancelReservationNotifiesRefundAmount(t *testing.T) {
	repo := &mockReservationRepository{
		getByIDReturn: &ReservationDetail{
			ID: 81, UserID: 1, HostID: 2, CabinID: 3,
			StartDate: futureDate(10), EndDate: futureDate(12), Status: StatusConfirmed,
		},
		getCabinName: "Cabaña Río",
	}
	notifier := &mockNotifier{}
	recorder := &mockRefundRecorder{
		returnRefund: &CancelRefund{Amount: 90, Percent: 1, Kind: "total"},
	}
	service := NewService(repo, notifier)
	service.SetRefundRecorder(recorder)

	if _, _, err := service.CancelReservationWithRefund(context.Background(), 81); err != nil {
		t.Fatalf("CancelReservationWithRefund returned error: %v", err)
	}

	cancelled := notifier.byTipo(notifications.TipoReservaCancelada)
	if len(cancelled) != 2 {
		t.Fatalf("expected 2 cancellation notifications (guest + host), got %d", len(cancelled))
	}
	if !strings.Contains(cancelled[0].Mensaje, "reembolso total") || !strings.Contains(cancelled[0].Mensaje, "Q 90.00") {
		t.Errorf("la notificación al huésped debería informar el reembolso, got %q", cancelled[0].Mensaje)
	}
	if !strings.Contains(cancelled[1].Mensaje, "reembolso total") {
		t.Errorf("la notificación al anfitrión debería informar el reembolso, got %q", cancelled[1].Mensaje)
	}
}

func TestAdvanceReservationStatesNotifiesGuestHostAndAdmins(t *testing.T) {
	repo := &mockReservationRepository{
		advanceReturnIDs: []int{42},
		getByIDReturn: &ReservationDetail{
			ID: 42, UserID: 1, HostID: 2, CabinID: 3,
			StartDate: "2026-01-01", EndDate: "2026-01-05", Status: StatusFinished,
		},
		getCabinName: "Cabaña Bosque",
	}
	notifier := &mockNotifier{}
	service := NewService(repo, notifier)

	count, err := service.AdvanceReservationStates(context.Background())
	if err != nil {
		t.Fatalf("AdvanceReservationStates returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("AdvanceReservationStates() = %d, want 1", count)
	}

	finished := notifier.byTipo(notifications.TipoReservaFinalizada)
	if len(finished) != 2 {
		t.Fatalf("expected 2 finished notifications (guest + host), got %d", len(finished))
	}
	if finished[0].UserID != 1 || finished[1].UserID != 2 {
		t.Fatalf("notifications should target guest 1 and host 2, got %d and %d", finished[0].UserID, finished[1].UserID)
	}
	if len(notifier.adminInputs) != 1 {
		t.Fatalf("expected 1 admin notification, got %d", len(notifier.adminInputs))
	}
}

func TestUpdateReservationNotifiesGuestAndHost(t *testing.T) {
	repo := &mockReservationRepository{
		getByIDReturn: &ReservationDetail{
			ID: 88, UserID: 1, HostID: 2, CabinID: 3,
			StartDate: futureDate(10), EndDate: futureDate(15), Status: StatusPending,
		},
		getCabinName: "Cabaña Mar",
	}
	notifier := &mockNotifier{}
	service := NewService(repo, notifier)

	if _, err := service.UpdateReservation(context.Background(), 88, UpdateReservationRequest{
		CabinID:   3,
		StartDate: futureDate(12),
		EndDate:   futureDate(16),
	}); err != nil {
		t.Fatalf("UpdateReservation returned error: %v", err)
	}

	updated := notifier.byTipo(notifications.TipoReservaActualizada)
	if len(updated) != 2 {
		t.Fatalf("expected 2 update notifications (guest + host), got %d", len(updated))
	}
}

func pastDate(days int) string {
	return time.Now().AddDate(0, 0, -days).Format(dateLayout)
}

func checkinEligibility(status string, startDate, endDate string, nightlyRate, paid float64) *CheckinEligibility {
	return &CheckinEligibility{
		ReservationID: 50,
		UserID:        1,
		CabinID:       2,
		HostID:        3,
		Status:        status,
		StartDate:     startDate,
		EndDate:       endDate,
		NightlyRate:   nightlyRate,
		PaidTotal:     paid,
	}
}

func TestCheckinHappyPath(t *testing.T) {
	repo := &mockReservationRepository{
		checkinElig: checkinEligibility(StatusConfirmed, pastDate(1), futureDate(2), 100, 300),
		getByIDReturn: &ReservationDetail{
			ID: 50, UserID: 1, HostID: 3, CabinID: 2,
			StartDate: pastDate(1), EndDate: futureDate(2), Status: StatusInProgress,
		},
		getCabinName: "Cabaña Mar",
	}
	notifier := &mockNotifier{}
	service := NewService(repo, notifier)

	detail, err := service.Checkin(context.Background(), 50, 0)
	if err != nil {
		t.Fatalf("Checkin returned error: %v", err)
	}
	if detail.ID != 50 {
		t.Fatalf("Checkin() reservation ID = %d, want 50", detail.ID)
	}

	notified := notifier.byTipo(notifications.TipoReservaCheckin)
	if len(notified) != 2 {
		t.Fatalf("expected 2 checkin notifications (guest + host), got %d", len(notified))
	}
	if len(notifier.adminInputs) != 1 {
		t.Fatalf("expected 1 admin notification, got %d", len(notifier.adminInputs))
	}
}

func TestCheckinRejects(t *testing.T) {
	checkinTime := time.Now()
	testCases := []struct {
		name      string
		id        int
		cash      float64
		elig      *CheckinEligibility
		eligErr   error
		recordErr error
		wantErr   error
	}{
		{
			name: "invalid id",
			id:   0,
		},
		{
			name:    "reservation not found",
			id:      51,
			eligErr: ErrReservationNotFound,
			wantErr: ErrReservationNotFound,
		},
		{
			name:    "cancelled cannot check in",
			id:      52,
			elig:    checkinEligibility(StatusCancelled, pastDate(1), futureDate(2), 100, 300),
			wantErr: ErrReservationNotCheckinable,
		},
		{
			name:    "pending cannot check in",
			id:      57,
			elig:    checkinEligibility(StatusPending, pastDate(1), futureDate(2), 100, 300),
			wantErr: ErrReservationNotCheckinable,
		},
		{
			name: "already checked in",
			id:   58,
			elig: &CheckinEligibility{
				ReservationID: 58, UserID: 1, CabinID: 2, HostID: 3,
				Status: StatusInProgress, StartDate: pastDate(1), EndDate: futureDate(2),
				CheckedInAt: &checkinTime, NightlyRate: 100, PaidTotal: 300,
			},
			wantErr: ErrReservationNotCheckinable,
		},
		{
			name: "before start date",
			id:   53,
			elig: checkinEligibility(StatusConfirmed, futureDate(5), futureDate(7), 100, 200),
		},
		{
			name: "after end date",
			id:   54,
			elig: checkinEligibility(StatusConfirmed, pastDate(5), pastDate(1), 100, 400),
		},
		{
			name:    "payment incomplete",
			id:      55,
			elig:    checkinEligibility(StatusConfirmed, pastDate(1), futureDate(2), 100, 100),
			wantErr: ErrCheckinPaymentIncomplete,
		},
		{
			name: "cash must cover remaining exactly",
			id:   59,
			cash: 100,
			elig: checkinEligibility(StatusConfirmed, pastDate(1), futureDate(2), 100, 100),
		},
		{
			name: "cash rejected when nothing pending",
			id:   60,
			cash: 50,
			elig: checkinEligibility(StatusConfirmed, pastDate(1), futureDate(2), 100, 300),
		},
		{
			name:      "record error propagates",
			id:        56,
			elig:      checkinEligibility(StatusConfirmed, pastDate(1), futureDate(2), 100, 300),
			recordErr: errors.New("db down"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockReservationRepository{
				checkinElig:      tc.elig,
				checkinEligErr:   tc.eligErr,
				recordCheckinErr: tc.recordErr,
				getByIDReturn:    &ReservationDetail{ID: tc.id, Status: StatusInProgress},
			}
			service := NewService(repo, &mockNotifier{})

			_, err := service.Checkin(context.Background(), tc.id, tc.cash)
			if err == nil {
				t.Fatalf("Checkin() expected error, got nil")
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("Checkin() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestCheckinWithCashCompletesPayment(t *testing.T) {
	repo := &mockReservationRepository{
		checkinElig: checkinEligibility(StatusConfirmed, pastDate(1), futureDate(2), 100, 100),
		getByIDReturn: &ReservationDetail{
			ID: 50, UserID: 1, HostID: 3, CabinID: 2,
			StartDate: pastDate(1), EndDate: futureDate(2), Status: StatusInProgress,
		},
		getCabinName: "Cabaña Mar",
	}
	notifier := &mockNotifier{}
	service := NewService(repo, notifier)

	if _, err := service.Checkin(context.Background(), 50, 200); err != nil {
		t.Fatalf("Checkin with cash returned error: %v", err)
	}
	if repo.cashCalledWith == nil {
		t.Fatal("repository.CheckinWithCashPayment should be called")
	}
	if repo.cashCalledWith.amount != 200 || repo.cashCalledWith.payerID != 1 {
		t.Fatalf("cash call = %+v, want amount 200 payer 1", repo.cashCalledWith)
	}

	notified := notifier.byTipo(notifications.TipoReservaCheckin)
	if len(notified) != 2 {
		t.Fatalf("expected 2 checkin notifications (guest + host), got %d", len(notified))
	}
}

func TestPayAnticipoConfirmsReservationAndNotifies(t *testing.T) {
	repo := &mockReservationRepository{
		getByIDReturn: &ReservationDetail{
			ID: 10, UserID: 1, HostID: 2, CabinID: 3,
			StartDate: futureDate(10), EndDate: futureDate(15), Status: StatusPending,
		},
		nightlyRate:  100,
		getCabinName: "Cabaña Lago",
	}
	notifier := &mockNotifier{}
	service := NewService(repo, notifier)

	updated, amount, err := service.PayAnticipo(context.Background(), 10, 0, "")
	if err != nil {
		t.Fatalf("PayAnticipo returned error: %v", err)
	}
	if updated == nil || updated.Status != StatusConfirmed {
		t.Fatalf("PayAnticipo should return the confirmed reservation, got %+v", updated)
	}
	if amount != 150 {
		t.Fatalf("PayAnticipo amount = %v, want 150", amount)
	}
	if repo.payCalledWith == nil {
		t.Fatal("repository.PayAnticipo should be called")
	}
	if repo.payCalledWith.payerID != 1 {
		t.Fatalf("payerID = %d, want 1 (the guest)", repo.payCalledWith.payerID)
	}
	if repo.payCalledWith.method != PaymentMethodCash {
		t.Fatalf("method = %q, want default CASH", repo.payCalledWith.method)
	}

	confirmed := notifier.byTipo(notifications.TipoReservaConfirmada)
	if len(confirmed) != 2 {
		t.Fatalf("expected 2 confirmation notifications (guest + host), got %d", len(confirmed))
	}
	if confirmed[0].UserID != 1 || confirmed[1].UserID != 2 {
		t.Fatalf("notifications should target guest 1 and host 2, got %d and %d", confirmed[0].UserID, confirmed[1].UserID)
	}
	if len(notifier.adminInputs) != 1 {
		t.Fatalf("expected 1 admin notification, got %d", len(notifier.adminInputs))
	}
}

func TestPayAnticipoAcceptsCustomAmountAndMethod(t *testing.T) {
	newRepo := func() *mockReservationRepository {
		return &mockReservationRepository{
			getByIDReturn: &ReservationDetail{
				ID: 10, UserID: 1, HostID: 2, CabinID: 3,
				StartDate: futureDate(10), EndDate: futureDate(15), Status: StatusPending,
			},
			nightlyRate: 100,
		}
	}

	repo := newRepo()
	service := NewService(repo, &mockNotifier{})
	if _, amount, err := service.PayAnticipo(context.Background(), 10, 300, PaymentMethodCard); err != nil {
		t.Fatalf("PayAnticipo with custom amount returned error: %v", err)
	} else if amount != 300 {
		t.Fatalf("PayAnticipo amount = %v, want 300", amount)
	}
	if repo.payCalledWith == nil || repo.payCalledWith.method != PaymentMethodCard {
		t.Fatalf("expected tarjeta to reach the repository, got %+v", repo.payCalledWith)
	}

	repo = newRepo()
	service = NewService(repo, &mockNotifier{})
	if _, _, err := service.PayAnticipo(context.Background(), 10, 50, PaymentMethodCash); err == nil {
		t.Fatal("PayAnticipo should reject amounts below 30%")
	} else if repo.payCalledWith != nil {
		t.Fatal("repository.PayAnticipo should not be called")
	}

	repo = newRepo()
	service = NewService(repo, &mockNotifier{})
	if _, _, err := service.PayAnticipo(context.Background(), 10, 600, PaymentMethodCash); err == nil {
		t.Fatal("PayAnticipo should reject amounts above the total")
	}

	repo = newRepo()
	service = NewService(repo, &mockNotifier{})
	if _, _, err := service.PayAnticipo(context.Background(), 10, 150, "BITCOIN"); err == nil {
		t.Fatal("PayAnticipo should reject invalid methods")
	}
}

func TestPayAnticipoRejectsNonPendingReservation(t *testing.T) {
	for _, status := range []string{StatusConfirmed, StatusInProgress, StatusFinished, StatusCancelled} {
		repo := &mockReservationRepository{
			getByIDReturn: &ReservationDetail{
				ID: 11, UserID: 1, HostID: 2, CabinID: 3,
				StartDate: futureDate(10), EndDate: futureDate(15), Status: status,
			},
			nightlyRate: 100,
		}
		service := NewService(repo, &mockNotifier{})

		_, _, err := service.PayAnticipo(context.Background(), 11, 0, "")
		if !errors.Is(err, ErrReservationNotPending) {
			t.Fatalf("status %q: expected ErrReservationNotPending, got %v", status, err)
		}
		if repo.payCalledWith != nil {
			t.Fatalf("status %q: repository.PayAnticipo should not be called", status)
		}
	}
}

func TestCancelReservationRejectsFinishedAndCancelled(t *testing.T) {
	for _, status := range []string{StatusFinished, StatusCancelled} {
		repo := &mockReservationRepository{
			getByIDReturn: &ReservationDetail{
				ID: 12, UserID: 1, HostID: 2, CabinID: 3,
				StartDate: futureDate(10), EndDate: futureDate(15), Status: status,
			},
		}
		service := NewService(repo, &mockNotifier{})

		if _, err := service.CancelReservation(context.Background(), 12); !errors.Is(err, ErrReservationNotCancellable) {
			t.Fatalf("status %q: expected ErrReservationNotCancellable, got %v", status, err)
		}
	}
}
