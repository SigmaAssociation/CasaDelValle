package reservations

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"cdv-api/internal/notifications"
)

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
	updateReturnErr   error
	getCabinName      string
	getCabinNameErr   error
	finalizeReturnIDs []int
	finalizeErr       error
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

func (m *mockReservationRepository) UpdateReservation(ctx context.Context, id int, req UpdateReservationRequest) error {
	return m.updateReturnErr
}

func (m *mockReservationRepository) GetCabinName(ctx context.Context, cabinID int) (string, error) {
	return m.getCabinName, m.getCabinNameErr
}

func (m *mockReservationRepository) FinalizePastReservations(ctx context.Context) ([]int, error) {
	return m.finalizeReturnIDs, m.finalizeErr
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
			StartDate: futureDate(10), EndDate: futureDate(15), Status: StatusActive,
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
			StartDate: futureDate(10), EndDate: futureDate(15), Status: StatusActive,
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
			StartDate: futureDate(10), EndDate: futureDate(15), Status: StatusActive,
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

func TestFinalizePastReservationsNotifiesGuestHostAndAdmins(t *testing.T) {
	repo := &mockReservationRepository{
		finalizeReturnIDs: []int{42},
		getByIDReturn: &ReservationDetail{
			ID: 42, UserID: 1, HostID: 2, CabinID: 3,
			StartDate: "2026-01-01", EndDate: "2026-01-05", Status: StatusFinished,
		},
		getCabinName: "Cabaña Bosque",
	}
	notifier := &mockNotifier{}
	service := NewService(repo, notifier)

	count, err := service.FinalizePastReservations(context.Background())
	if err != nil {
		t.Fatalf("FinalizePastReservations returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("FinalizePastReservations() = %d, want 1", count)
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
			StartDate: futureDate(10), EndDate: futureDate(15), Status: StatusActive,
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
		checkinElig: checkinEligibility(StatusActive, pastDate(1), futureDate(2), 100, 300),
		getByIDReturn: &ReservationDetail{
			ID: 50, UserID: 1, HostID: 3, CabinID: 2,
			StartDate: pastDate(1), EndDate: futureDate(2), Status: StatusCheckedIn,
		},
		getCabinName: "Cabaña Mar",
	}
	notifier := &mockNotifier{}
	service := NewService(repo, notifier)

	detail, err := service.Checkin(context.Background(), 50)
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
	testCases := []struct {
		name      string
		id        int
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
			name:    "not active",
			id:      52,
			elig:    checkinEligibility(StatusCancelled, pastDate(1), futureDate(2), 100, 300),
			wantErr: ErrReservationNotCheckinable,
		},
		{
			name: "before start date",
			id:   53,
			elig: checkinEligibility(StatusActive, futureDate(5), futureDate(7), 100, 200),
		},
		{
			name: "after end date",
			id:   54,
			elig: checkinEligibility(StatusActive, pastDate(5), pastDate(1), 100, 400),
		},
		{
			name:    "payment incomplete",
			id:      55,
			elig:    checkinEligibility(StatusActive, pastDate(1), futureDate(2), 100, 100),
			wantErr: ErrCheckinPaymentIncomplete,
		},
		{
			name:      "record error propagates",
			id:        56,
			elig:      checkinEligibility(StatusActive, pastDate(1), futureDate(2), 100, 300),
			recordErr: errors.New("db down"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockReservationRepository{
				checkinElig:      tc.elig,
				checkinEligErr:   tc.eligErr,
				recordCheckinErr: tc.recordErr,
				getByIDReturn:    &ReservationDetail{ID: tc.id, Status: StatusCheckedIn},
			}
			service := NewService(repo, &mockNotifier{})

			_, err := service.Checkin(context.Background(), tc.id)
			if err == nil {
				t.Fatalf("Checkin() expected error, got nil")
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("Checkin() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}
