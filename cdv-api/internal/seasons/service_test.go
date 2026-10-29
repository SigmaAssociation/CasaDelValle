package seasons

import (
	"context"
	"errors"
	"math"
	"testing"
)

type mockSeasonRepository struct {
	*Repository

	listIncludeInactive bool
	seasons             []Season
	listErr             error

	createCalls int
	createErr   error
	created     *Season

	updateErr error

	setActiveCalls int
	setActiveID    int
	setActiveValue bool
	setActiveErr   error
}

func (m *mockSeasonRepository) ListSeasons(_ context.Context, includeInactive bool) ([]Season, error) {
	m.listIncludeInactive = includeInactive
	return m.seasons, m.listErr
}

func (m *mockSeasonRepository) CreateSeason(_ context.Context, req CreateSeasonRequest) (*Season, error) {
	m.createCalls++
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.created != nil {
		return m.created, nil
	}
	return &Season{
		ID:                  1,
		Name:                req.Name,
		SuggestedAdjustment: req.SuggestedAdjustment,
		Active:              req.IsActive(),
		StartDate:           req.StartDate,
		EndDate:             req.EndDate,
	}, nil
}

func (m *mockSeasonRepository) UpdateSeason(_ context.Context, _ int, _ UpdateSeasonRequest) error {
	return m.updateErr
}

func (m *mockSeasonRepository) SetSeasonActive(_ context.Context, id int, active bool) error {
	m.setActiveCalls++
	m.setActiveID = id
	m.setActiveValue = active
	return m.setActiveErr
}

func TestValidateDates(t *testing.T) {
	tests := []struct {
		name, start, end string
		wantErr          bool
	}{
		{"same day is valid", "2026-12-20", "2026-12-20", false},
		{"range is valid", "2026-12-20", "2026-12-21", false},
		{"end before start", "2026-12-21", "2026-12-20", true},
		{"invalid format", "20-12-2026", "2026-12-21", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotErr := validateDates(tt.start, tt.end) != nil; gotErr != tt.wantErr {
				t.Fatalf("validateDates() error = %v, want %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestValidateSeasonRequiresName(t *testing.T) {
	if err := validateSeason("  ", 10, "2026-12-20", "2026-12-21"); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}
}

func TestValidateSeasonRejectsLongName(t *testing.T) {
	name := make([]rune, 101)
	for i := range name {
		name[i] = 'a'
	}
	if err := validateSeason(string(name), 10, "2026-12-20", "2026-12-21"); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}
}

func TestValidateSeasonRejectsInvalidAdjustment(t *testing.T) {
	values := []float64{-1, 100.01, math.NaN()}
	for _, value := range values {
		if err := validateSeason("Navidad", value, "2026-12-20", "2026-12-21"); !errors.Is(err, ErrInvalidAdjustment) {
			t.Fatalf("adjustment %v: expected ErrInvalidAdjustment, got %v", value, err)
		}
	}
}

func TestCreateSeasonDefaultsActive(t *testing.T) {
	repo := &mockSeasonRepository{}
	service := NewService(repo)

	season, err := service.CreateSeason(context.Background(), CreateSeasonRequest{
		Name:                "Navidad",
		SuggestedAdjustment: 25,
		StartDate:           "2026-12-20",
		EndDate:             "2026-12-31",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !season.Active {
		t.Fatal("expected season to be active by default")
	}
	if repo.createCalls != 1 {
		t.Fatal("expected repository to be called once")
	}
}

func TestCreateSeasonHonorsExplicitInactive(t *testing.T) {
	repo := &mockSeasonRepository{}
	service := NewService(repo)
	inactive := false

	season, err := service.CreateSeason(context.Background(), CreateSeasonRequest{
		Name:                "Temporada de prueba",
		SuggestedAdjustment: 10,
		Active:              &inactive,
		StartDate:           "2026-12-20",
		EndDate:             "2026-12-21",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if season.Active {
		t.Fatal("expected season to stay inactive")
	}
}

func TestCreateSeasonRejectsInvalidInputWithoutTouchingRepository(t *testing.T) {
	repo := &mockSeasonRepository{}
	service := NewService(repo)

	_, err := service.CreateSeason(context.Background(), CreateSeasonRequest{
		Name:                "Navidad",
		SuggestedAdjustment: 150,
		StartDate:           "2026-12-20",
		EndDate:             "2026-12-31",
	})
	if !errors.Is(err, ErrInvalidAdjustment) {
		t.Fatalf("expected ErrInvalidAdjustment, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatal("repository must not be called for invalid input")
	}
}

func TestCreateSeasonMapsOverlap(t *testing.T) {
	repo := &mockSeasonRepository{createErr: ErrSeasonOverlap}
	service := NewService(repo)

	_, err := service.CreateSeason(context.Background(), CreateSeasonRequest{
		Name:                "Navidad",
		SuggestedAdjustment: 25,
		StartDate:           "2026-12-20",
		EndDate:             "2026-12-31",
	})
	if !errors.Is(err, ErrSeasonOverlap) {
		t.Fatalf("expected ErrSeasonOverlap, got %v", err)
	}
}

func TestUpdateSeasonPropagatesOverlap(t *testing.T) {
	repo := &mockSeasonRepository{updateErr: ErrSeasonOverlap}
	service := NewService(repo)

	err := service.UpdateSeason(context.Background(), 1, UpdateSeasonRequest{
		Name:                "Navidad",
		SuggestedAdjustment: 25,
		Active:              true,
		StartDate:           "2026-12-20",
		EndDate:             "2026-12-31",
	})
	if !errors.Is(err, ErrSeasonOverlap) {
		t.Fatalf("expected ErrSeasonOverlap, got %v", err)
	}
}

func TestDeactivateSeason(t *testing.T) {
	repo := &mockSeasonRepository{}
	service := NewService(repo)

	if err := service.DeactivateSeason(context.Background(), 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.setActiveCalls != 1 || repo.setActiveID != 7 || repo.setActiveValue {
		t.Fatalf("expected SetSeasonActive(7, false), got calls=%d id=%d active=%v",
			repo.setActiveCalls, repo.setActiveID, repo.setActiveValue)
	}
}

func TestDeactivateSeasonNotFound(t *testing.T) {
	repo := &mockSeasonRepository{setActiveErr: ErrSeasonNotFound}
	service := NewService(repo)

	err := service.DeactivateSeason(context.Background(), 99)
	if !errors.Is(err, ErrSeasonNotFound) {
		t.Fatalf("expected ErrSeasonNotFound, got %v", err)
	}
}

func TestListSeasonsPassesIncludeInactive(t *testing.T) {
	repo := &mockSeasonRepository{seasons: []Season{{ID: 1, Name: "Navidad"}}}
	service := NewService(repo)

	got, err := service.ListSeasons(context.Background(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.listIncludeInactive {
		t.Fatal("expected includeInactive to be forwarded to the repository")
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 season, got %d", len(got))
	}
}
