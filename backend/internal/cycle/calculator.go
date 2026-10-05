package cycle

import (
	"time"
)

// DaysInMonth returns the total number of days in the given year and month.
func DaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// MakeDateClamped constructs a UTC date for the given year, month, and day,
// clamping the day to the last valid day of the month if day exceeds days in month.
func MakeDateClamped(year int, month time.Month, day int) time.Time {
	if day < 1 {
		day = 1
	}
	dim := DaysInMonth(year, month)
	if day > dim {
		day = dim
	}
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// CalculateCycle calculates the inclusive [startDate, endDate] of the financial cycle
// that targetDate falls into, based on the user's cycleStartDay (1-31).
func CalculateCycle(startDay int, targetDate time.Time) (time.Time, time.Time) {
	if startDay < 1 {
		startDay = 1
	} else if startDay > 31 {
		startDay = 31
	}

	target := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)

	// Determine the start date of the cycle for the current target month
	startOfTargetMonth := MakeDateClamped(target.Year(), target.Month(), startDay)

	var cycleStart time.Time
	var cycleEnd time.Time

	if !target.Before(startOfTargetMonth) {
		// Target date falls on or after startDay of the target month
		cycleStart = startOfTargetMonth
		nextMonth := time.Date(target.Year(), target.Month()+1, 1, 0, 0, 0, 0, time.UTC)
		nextCycleStart := MakeDateClamped(nextMonth.Year(), nextMonth.Month(), startDay)
		cycleEnd = nextCycleStart.AddDate(0, 0, -1)
	} else {
		// Target date falls before startDay of the target month; belongs to previous month's cycle
		prevMonth := time.Date(target.Year(), target.Month()-1, 1, 0, 0, 0, 0, time.UTC)
		cycleStart = MakeDateClamped(prevMonth.Year(), prevMonth.Month(), startDay)
		cycleEnd = startOfTargetMonth.AddDate(0, 0, -1)
	}

	return cycleStart, cycleEnd
}
