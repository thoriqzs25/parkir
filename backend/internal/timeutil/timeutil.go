package timeutil

import "time"

// Jakarta is the Asia/Jakarta timezone (WIB, UTC+7).
// Indonesia does not observe daylight saving time.
var Jakarta *time.Location

func init() {
	var err error
	Jakarta, err = time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback to fixed +7 offset if tzdata is unavailable.
		Jakarta = time.FixedZone("Asia/Jakarta", 7*60*60)
	}
}

// FormatJakarta returns the time formatted in Asia/Jakarta timezone.
func FormatJakarta(t time.Time, layout string) string {
	return t.In(Jakarta).Format(layout)
}

// DateJakarta returns the date portion of t in Asia/Jakarta timezone.
func DateJakarta(t time.Time) time.Time {
	j := t.In(Jakarta)
	return time.Date(j.Year(), j.Month(), j.Day(), 0, 0, 0, 0, Jakarta)
}
