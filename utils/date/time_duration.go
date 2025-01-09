package date

import "time"

// NowUntilEoD calculates the duration until the end of the day.
func NowUntilEoD() time.Duration {
	now := time.Now()
	endOfDay := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
	return endOfDay.Sub(now)
}
