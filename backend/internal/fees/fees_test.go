package fees

import (
	"testing"
	"time"
)

func TestCalculate(t *testing.T) {
	base := time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC)
	rate := CalculationInput{
		FirstHourRate:        5000,
		SubsequentHourlyRate: 3000,
		DailyFlatRate:        50000,
	}

	tests := []struct {
		name     string
		duration time.Duration
		wantFee  float64
		wantDur  int
	}{
		{"1 hour", 1 * time.Hour, 5000, 1},
		{"3 hours", 3 * time.Hour, 11000, 3},
		{"12 hours", 12 * time.Hour, 38000, 12},
		{"24 hours exactly", 24 * time.Hour, 50000, 24},
		{"25 hours (new block starts)", 25 * time.Hour, 55000, 25},
		{"36 hours (user confirmed)", 36 * time.Hour, 88000, 36},
		{"48 hours (two full days)", 48 * time.Hour, 100000, 48},
		{"49 hours (two days + 1 hour)", 49 * time.Hour, 105000, 49},
		{"0 duration (edge case)", 0, 5000, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, dur := Calculate(CalculationInput{
				FirstHourRate:        rate.FirstHourRate,
				SubsequentHourlyRate: rate.SubsequentHourlyRate,
				DailyFlatRate:        rate.DailyFlatRate,
				CheckInAt:            base,
				CheckOutAt:           base.Add(tt.duration),
			})
			if got != tt.wantFee {
				t.Errorf("fee = %.0f, want %.0f", got, tt.wantFee)
			}
			if dur != tt.wantDur {
				t.Errorf("durationHours = %d, want %d", dur, tt.wantDur)
			}
		})
	}
}

func TestCalculate_RateCapsAtDaily(t *testing.T) {
	base := time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC)

	// Low daily cap should cap each block
	got, dur := Calculate(CalculationInput{
		FirstHourRate:        20000,
		SubsequentHourlyRate: 10000,
		DailyFlatRate:        25000,
		CheckInAt:            base,
		CheckOutAt:           base.Add(24 * time.Hour),
	})
	if got != 25000 {
		t.Errorf("single block capped = %.0f, want 25000", got)
	}
	if dur != 24 {
		t.Errorf("durationHours = %d, want 24", dur)
	}
}
