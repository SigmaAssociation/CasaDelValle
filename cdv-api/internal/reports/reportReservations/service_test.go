package reportreservations

import (
	"context"
	"testing"
)

type reportRepositoryStub struct{ cabins []CabinReport }

func (s reportRepositoryStub) GetReport(context.Context, ReportFilter) ([]CabinReport, error) {
	return s.cabins, nil
}

func TestGenerateReportCalculatesCancellationPercentage(t *testing.T) {
	service := NewService(reportRepositoryStub{cabins: []CabinReport{{TotalReservations: 3, Reservations: []ReservationSummary{{Status: "cancelada"}, {Status: "confirmada"}, {Status: "cancelada"}}}}})
	report, err := service.GenerateReport(context.Background(), ReportFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if report.CancelledReservations != 2 || report.CancellationPercentage != 66.67 {
		t.Fatalf("got %d and %.2f", report.CancelledReservations, report.CancellationPercentage)
	}
}
