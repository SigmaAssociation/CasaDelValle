package seasons

import "testing"

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
	if err := validateSeason("  ", "2026-12-20", "2026-12-21"); err == nil {
		t.Fatal("expected blank name to be rejected")
	}
}
