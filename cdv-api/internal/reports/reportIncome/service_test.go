package reportincome

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRepository struct {
	movements  []IncomeMovement
	err        error
	lastFilter ReportFilter
	calls      int
}

func (m *mockRepository) GetReport(ctx context.Context, filter ReportFilter) ([]IncomeMovement, error) {
	m.calls++
	m.lastFilter = filter
	if m.err != nil {
		return nil, m.err
	}
	return m.movements, nil
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestGenerateReportCalculatesNetIncome(t *testing.T) {
	repo := &mockRepository{
		movements: []IncomeMovement{
			{ID: 1, Type: "anticipo", Amount: 500},
			{ID: 2, Type: "pago", Amount: 300},
			{ID: 3, Type: "reembolso", Amount: -100},
		},
	}
	service := NewService(repo)

	report, err := service.GenerateReport(context.Background(), ReportFilter{
		StartDate: date(2026, time.October, 1),
		EndDate:   date(2026, time.October, 31),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.TotalPayments != 800 {
		t.Errorf("TotalPayments = %v, want 800", report.TotalPayments)
	}
	if report.TotalRefunds != 100 {
		t.Errorf("TotalRefunds = %v, want 100", report.TotalRefunds)
	}
	if report.TotalIncome != 700 {
		t.Errorf("TotalIncome = %v, want 700", report.TotalIncome)
	}
	if report.TotalMovements != 3 {
		t.Errorf("TotalMovements = %v, want 3", report.TotalMovements)
	}
	if report.Message != "" {
		t.Errorf("Message = %q, want empty", report.Message)
	}
	if report.StartDate != "2026-10-01" || report.EndDate != "2026-10-31" {
		t.Errorf("date range = %s - %s", report.StartDate, report.EndDate)
	}
}

func TestGenerateReportEmptyResult(t *testing.T) {
	repo := &mockRepository{movements: []IncomeMovement{}}
	service := NewService(repo)

	report, err := service.GenerateReport(context.Background(), ReportFilter{
		StartDate: date(2026, time.January, 1),
		EndDate:   date(2026, time.January, 31),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.TotalIncome != 0 || report.TotalPayments != 0 || report.TotalRefunds != 0 {
		t.Errorf("expected zero totals, got %+v", report)
	}
	if report.TotalMovements != 0 {
		t.Errorf("TotalMovements = %v, want 0", report.TotalMovements)
	}
	if report.Message != noDataMessage {
		t.Errorf("Message = %q, want %q", report.Message, noDataMessage)
	}
	if report.Movements == nil {
		t.Error("Movements should not be nil")
	}
}

func TestGenerateReportRejectsInvertedRange(t *testing.T) {
	repo := &mockRepository{}
	service := NewService(repo)

	_, err := service.GenerateReport(context.Background(), ReportFilter{
		StartDate: date(2026, time.October, 20),
		EndDate:   date(2026, time.October, 10),
	})
	if !errors.Is(err, ErrInvalidDateRange) {
		t.Fatalf("error = %v, want ErrInvalidDateRange", err)
	}
	if repo.calls != 0 {
		t.Errorf("repository calls = %d, want 0", repo.calls)
	}
}

func TestGenerateReportRejectsMissingDates(t *testing.T) {
	repo := &mockRepository{}
	service := NewService(repo)

	_, err := service.GenerateReport(context.Background(), ReportFilter{
		EndDate: date(2026, time.October, 10),
	})
	if !errors.Is(err, ErrInvalidDateRange) {
		t.Fatalf("error = %v, want ErrInvalidDateRange", err)
	}
}

func TestGenerateReportRejectsInvalidCabin(t *testing.T) {
	repo := &mockRepository{}
	service := NewService(repo)
	cabinID := 0

	_, err := service.GenerateReport(context.Background(), ReportFilter{
		StartDate: date(2026, time.October, 1),
		EndDate:   date(2026, time.October, 10),
		CabinID:   &cabinID,
	})
	if !errors.Is(err, ErrInvalidFilter) {
		t.Fatalf("error = %v, want ErrInvalidFilter", err)
	}
	if repo.calls != 0 {
		t.Errorf("repository calls = %d, want 0", repo.calls)
	}
}

func TestGenerateReportPassesHostFilterAndNormalizesDates(t *testing.T) {
	repo := &mockRepository{movements: []IncomeMovement{}}
	service := NewService(repo)
	hostID := uint(7)
	cabinID := 3
	start := time.Date(2026, time.October, 1, 18, 30, 0, 0, time.UTC)
	end := time.Date(2026, time.October, 31, 23, 59, 0, 0, time.UTC)

	_, err := service.GenerateReport(context.Background(), ReportFilter{
		StartDate: start,
		EndDate:   end,
		CabinID:   &cabinID,
		HostID:    &hostID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.lastFilter.HostID == nil || *repo.lastFilter.HostID != hostID {
		t.Errorf("HostID = %v, want %v", repo.lastFilter.HostID, hostID)
	}
	if repo.lastFilter.CabinID == nil || *repo.lastFilter.CabinID != cabinID {
		t.Errorf("CabinID = %v, want %v", repo.lastFilter.CabinID, cabinID)
	}
	if !repo.lastFilter.StartDate.Equal(date(2026, time.October, 1)) {
		t.Errorf("StartDate = %v, want midnight", repo.lastFilter.StartDate)
	}
	if !repo.lastFilter.EndDate.Equal(date(2026, time.October, 31)) {
		t.Errorf("EndDate = %v, want midnight", repo.lastFilter.EndDate)
	}
}

func TestGenerateReportPropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("boom")
	repo := &mockRepository{err: repoErr}
	service := NewService(repo)

	_, err := service.GenerateReport(context.Background(), ReportFilter{
		StartDate: date(2026, time.October, 1),
		EndDate:   date(2026, time.October, 10),
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("error = %v, want %v", err, repoErr)
	}
}
