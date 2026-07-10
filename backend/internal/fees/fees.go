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

// Calculate computes the fee using a recurring 24-hour block model:
//
//	duration_hours = CEIL((check_out - check_in) in seconds / 3600)
//	if duration_hours < 1: duration_hours = 1
//
//	For each 24-hour block:
//	  block_hours = min(remaining, 24)
//	  if block_hours == 1:
//	      block_fee = first_hour_rate
//	  else:
//	      block_fee = first_hour_rate + (block_hours - 1) * subsequent_hourly_rate
//	  block_fee = min(block_fee, daily_flat_rate)
//	  total += block_fee
//
//	Loop repeats for subsequent 24-hour periods.
func Calculate(input CalculationInput) (fee float64, durationHours int) {
	diff := input.CheckOutAt.Sub(input.CheckInAt)
	durationHours = int(math.Ceil(diff.Hours()))
	if durationHours < 1 {
		durationHours = 1
	}

	var total float64
	remaining := durationHours

	for remaining > 0 {
		blockHours := remaining
		if blockHours > 24 {
			blockHours = 24
		}

		var raw float64
		if blockHours == 1 {
			raw = input.FirstHourRate
		} else {
			raw = input.FirstHourRate + float64(blockHours-1)*input.SubsequentHourlyRate
		}

		if raw > input.DailyFlatRate {
			total += input.DailyFlatRate
		} else {
			total += raw
		}

		remaining -= blockHours
	}

	return total, durationHours
}
