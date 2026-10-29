package maintenance

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubRepository struct{ called bool }

func (r *stubRepository) List(context.Context, int, uint) ([]Maintenance, error) { return nil, nil }
func (r *stubRepository) Preview(context.Context, int, uint, DateRequest) (*Preview, error) {
	r.called = true
	return &Preview{}, nil
}
func (r *stubRepository) Create(context.Context, int, uint, CreateRequest) (*Maintenance, error) {
	r.called = true
	return &Maintenance{ID: 1}, nil
}
func (r *stubRepository) Cancel(context.Context, int, int, uint) (*Maintenance, error) {
	r.called = true
	return &Maintenance{ID: 1}, nil
}

func TestMaintenanceDatesUseGuatemalaAndExclusiveEnd(t *testing.T) {
	now := time.Date(2030, 2, 2, 2, 0, 0, 0, time.UTC) // Todavía es 1 de febrero en Guatemala.
	for _, tt := range []struct {
		start, end string
		valid      bool
	}{
		{"2030-02-01", "2030-02-02", true},
		{"2030-01-31", "2030-02-02", false},
		{"2030-02-01", "2030-02-01", false},
		{"2030-02-03", "2030-02-02", false},
		{"2030-02-30", "2030-03-02", false},
		{"2030-2-01", "2030-02-02", false},
	} {
		if err := validateDates(tt.start, tt.end, now); (err == nil) != tt.valid {
			t.Errorf("%s-%s: %v", tt.start, tt.end, err)
		}
	}
}
func (r *stubRepository) Pending(context.Context, int, uint) (*PendingDecision, error) {
	return nil, nil
}
func (r *stubRepository) Resolve(context.Context, int, uint, DecisionRequest) error {
	r.called = true
	return nil
}
func TestRejectsInvalidGuestDecisions(t *testing.T) {
	for _, req := range []DecisionRequest{{Action: "reasignar"}, {Action: "reprogramar", StartDate: "2099-01-05", EndDate: "2099-01-03"}, {Action: "reembolsar", StartDate: "2099-01-01"}} {
		repo := &stubRepository{}
		err := NewService(repo).Resolve(context.Background(), 1, 20, req)
		if err == nil || repo.called {
			t.Fatalf("decisión inválida aceptada: %+v", req)
		}
	}
}
func TestReasonRequiredAndCancellationHasNoDateRestriction(t *testing.T) {
	repo := &stubRepository{}
	_, err := NewService(repo).Create(context.Background(), 1, 3, CreateRequest{"2099-01-01", "2099-01-03", "  "})
	var invalid validationError
	if !errors.As(err, &invalid) || repo.called {
		t.Fatal("aceptó motivo vacío")
	}
	if _, err := NewService(repo).Cancel(context.Background(), 1, 1, 3); err != nil || !repo.called {
		t.Fatal("no permitió cancelar")
	}
}
