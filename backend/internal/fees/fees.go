package fees

import (
	"math"
	"time"
)

// CalculationInput holds the rate fields and the parking window.
type CalculationInput struct {
	FirstHourRate        float64
	SubsequentHourlyRate float64
	DailyFlatRate        float64
	CheckInAt            time.Time
	CheckOutAt           time.Time
}

// Calculate computes the fee based on the PLAN.md formula:
//
//	duration_hours = CEIL((check_out - check_in) in seconds / 3600)
//	if duration_hours == 0: duration_hours = 1
//	if duration_hours == 1:
//	    raw_fee = first_hour_rate
//	else:
//	    raw_fee = first_hour_rate + (duration_hours - 1) * subsequent_hourly_rate
//	fee = MIN(raw_fee, daily_flat_rate)
func Calculate(input CalculationInput) (fee float64, durationHours int) {
	diff := input.CheckOutAt.Sub(input.CheckInAt)
	durationHours = int(math.Ceil(diff.Hours()))
	if durationHours < 1 {
		durationHours = 1
	}

	var raw float64
	if durationHours == 1 {
		raw = input.FirstHourRate
	} else {
		raw = input.FirstHourRate + float64(durationHours-1)*input.SubsequentHourlyRate
	}

	if raw > input.DailyFlatRate {
		fee = input.DailyFlatRate
	} else {
		fee = raw
	}
	return fee, durationHours
}
