package reservations

import (
	"context"
	"errors"
	"testing"
	"time"

	"cdv-api/internal/notifications"
)

func checkoutState(status string, checkedIn bool) *CheckinEligibility {
	state := &CheckinEligibility{
		ReservationID: 70,
		UserID:        1,
		CabinID:       2,
		HostID:        3,
		Status:        status,
		StartDate:     pastDate(2),
		EndDate:       futureDate(1),
		NightlyRate:   100,
		PaidTotal:     300,
	}
	if checkedIn {
		at := time.Now().Add(-48 * time.Hour)
		state.CheckedInAt = &at
	}
	return state
}

func TestCheckoutFinalizesReservationAndNotifies(t *testing.T) {
	repo := &mockReservationRepository{
		checkinElig: checkoutState(StatusInProgress, true),
		getByIDReturn: &ReservationDetail{
			ID: 70, UserID: 1, HostID: 3, CabinID: 2,
			StartDate: pastDate(2), EndDate: futureDate(1), Status: StatusInProgress,
		},
		getCabinName: "Cabaña Bosque",
	}
	notifier := &mockNotifier{}
	service := NewService(repo, notifier)

	detail, err := service.Checkout(context.Background(), 70)
	if err != nil {
		t.Fatalf("Checkout returned error: %v", err)
	}
	if !repo.recordCheckoutCalled {
		t.Fatal("repository.RecordCheckout should be called")
	}
	if detail == nil || detail.Status != StatusFinished {
		t.Fatalf("Checkout should return the finished reservation, got %+v", detail)
	}

	finished := notifier.byTipo(notifications.TipoReservaFinalizada)
	if len(finished) != 2 {
		t.Fatalf("expected 2 finished notifications (guest + host), got %d", len(finished))
	}
	if finished[0].UserID != 1 || finished[1].UserID != 3 {
		t.Fatalf("notifications should target guest 1 and host 3, got %d and %d", finished[0].UserID, finished[1].UserID)
	}
	if len(notifier.adminInputs) != 1 {
		t.Fatalf("expected 1 admin notification, got %d", len(notifier.adminInputs))
	}
}

func TestCheckoutRejects(t *testing.T) {
	testCases := []struct {
		name    string
		id      int
		elig    *CheckinEligibility
		eligErr error
		wantErr error
	}{
		{
			name: "invalid id",
			id:   0,
		},
		{
			name:    "reservation not found",
			id:      71,
			eligErr: ErrReservationNotFound,
			wantErr: ErrReservationNotFound,
		},
		{
			name:    "in progress without check-in is rejected",
			id:      72,
			elig:    checkoutState(StatusInProgress, false),
			wantErr: ErrCheckinRequired,
		},
		{
			name:    "confirmed (no check-in yet) is rejected",
			id:      73,
			elig:    checkoutState(StatusConfirmed, false),
			wantErr: ErrReservationNotCheckoutable,
		},
		{
			name:    "pending is rejected",
			id:      74,
			elig:    checkoutState(StatusPending, false),
			wantErr: ErrReservationNotCheckoutable,
		},
		{
			name:    "already finished is rejected",
			id:      75,
			elig:    checkoutState(StatusFinished, true),
			wantErr: ErrReservationNotCheckoutable,
		},
		{
			name:    "cancelled is rejected",
			id:      76,
			elig:    checkoutState(StatusCancelled, false),
			wantErr: ErrReservationNotCheckoutable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockReservationRepository{
				checkinElig:    tc.elig,
				checkinEligErr: tc.eligErr,
				getByIDReturn:  &ReservationDetail{ID: tc.id, Status: StatusInProgress},
			}
			notifier := &mockNotifier{}
			service := NewService(repo, notifier)

			_, err := service.Checkout(context.Background(), tc.id)
			if err == nil {
				t.Fatal("Checkout() expected error, got nil")
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("Checkout() error = %v, want %v", err, tc.wantErr)
			}
			if repo.recordCheckoutCalled {
				t.Fatal("repository.RecordCheckout should not be called for a rejected check-out")
			}
			if len(notifier.inputs) != 0 || len(notifier.adminInputs) != 0 {
				t.Fatal("no notifications should be sent for a rejected check-out")
			}
		})
	}
}

func TestCheckoutPropagatesRepositoryErrors(t *testing.T) {
	t.Run("record error", func(t *testing.T) {
		dbErr := errors.New("db down")
		repo := &mockReservationRepository{
			checkinElig:       checkoutState(StatusInProgress, true),
			recordCheckoutErr: dbErr,
			getByIDReturn:     &ReservationDetail{ID: 70, Status: StatusInProgress},
		}
		notifier := &mockNotifier{}
		service := NewService(repo, notifier)

		if _, err := service.Checkout(context.Background(), 70); !errors.Is(err, dbErr) {
			t.Fatalf("Checkout() error = %v, want %v", err, dbErr)
		}
		if len(notifier.inputs) != 0 {
			t.Fatal("no notifications should be sent when the check-out fails")
		}
	})

	t.Run("concurrent finalization", func(t *testing.T) {
		repo := &mockReservationRepository{
			checkinElig:       checkoutState(StatusInProgress, true),
			recordCheckoutErr: ErrReservationNotCheckoutable,
			getByIDReturn:     &ReservationDetail{ID: 70, Status: StatusFinished},
		}
		service := NewService(repo, &mockNotifier{})

		if _, err := service.Checkout(context.Background(), 70); !errors.Is(err, ErrReservationNotCheckoutable) {
			t.Fatalf("Checkout() error = %v, want ErrReservationNotCheckoutable", err)
		}
	})
}
