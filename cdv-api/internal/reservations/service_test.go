package reservations

import (
	"context"
	"strings"
	"testing"
	"time"
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
	updateReturnErr   error
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

func (m *mockReservationRepository) UpdateReservation(ctx context.Context, id int, req UpdateReservationRequest) error {
	return m.updateReturnErr
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
	service := NewService(repo)

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
	service := NewService(repo)

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
	service := NewService(repo)

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
