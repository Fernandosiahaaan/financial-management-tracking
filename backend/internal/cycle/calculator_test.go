package cycle

import (
	"testing"
	"time"
)

func TestCalculateCycle(t *testing.T) {
	tests := []struct {
		name          string
		startDay      int
		targetDate    string
		expectedStart string
		expectedEnd   string
	}{
		{
			name:          "Calendar month cycle (startDay=1)",
			startDay:      1,
			targetDate:    "2026-08-15",
			expectedStart: "2026-08-01",
			expectedEnd:   "2026-08-31",
		},
		{
			name:          "Cycle start 25 - date after start day (August 30)",
			startDay:      25,
			targetDate:    "2026-08-30",
			expectedStart: "2026-08-25",
			expectedEnd:   "2026-09-24",
		},
		{
			name:          "Cycle start 25 - exact start day (August 25)",
			startDay:      25,
			targetDate:    "2026-08-25",
			expectedStart: "2026-08-25",
			expectedEnd:   "2026-09-24",
		},
		{
			name:          "Cycle start 25 - date before start day (August 10)",
			startDay:      25,
			targetDate:    "2026-08-10",
			expectedStart: "2026-07-25",
			expectedEnd:   "2026-08-24",
		},
		{
			name:          "Cycle start 25 - exact cycle end day (August 24)",
			startDay:      25,
			targetDate:    "2026-08-24",
			expectedStart: "2026-07-25",
			expectedEnd:   "2026-08-24",
		},
		{
			name:          "Cycle start 31 - February non-leap year (2026)",
			startDay:      31,
			targetDate:    "2026-02-15",
			expectedStart: "2026-01-31",
			expectedEnd:   "2026-02-27",
		},
		{
			name:          "Cycle start 31 - February 28 non-leap year",
			startDay:      31,
			targetDate:    "2026-02-28",
			expectedStart: "2026-02-28",
			expectedEnd:   "2026-03-30",
		},
		{
			name:          "Cycle start 31 - February leap year (2024)",
			startDay:      31,
			targetDate:    "2024-02-29",
			expectedStart: "2024-02-29",
			expectedEnd:   "2024-03-30",
		},
		{
			name:          "Cycle start 31 - April 30 (30-day month)",
			startDay:      31,
			targetDate:    "2026-04-30",
			expectedStart: "2026-04-30",
			expectedEnd:   "2026-05-30",
		},
		{
			name:          "Year rollover - December 28 with startDay=25",
			startDay:      25,
			targetDate:    "2026-12-28",
			expectedStart: "2026-12-25",
			expectedEnd:   "2027-01-24",
		},
		{
			name:          "Year rollover - January 10 with startDay=25",
			startDay:      25,
			targetDate:    "2027-01-10",
			expectedStart: "2026-12-25",
			expectedEnd:   "2027-01-24",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := time.Parse("2006-01-02", tt.targetDate)
			if err != nil {
				t.Fatalf("invalid targetDate: %v", err)
			}

			start, end := CalculateCycle(tt.startDay, target)
			gotStart := start.Format("2006-01-02")
			gotEnd := end.Format("2006-01-02")

			if gotStart != tt.expectedStart {
				t.Errorf("start date mismatch: expected %s, got %s", tt.expectedStart, gotStart)
			}
			if gotEnd != tt.expectedEnd {
				t.Errorf("end date mismatch: expected %s, got %s", tt.expectedEnd, gotEnd)
			}
			if !end.After(start) && !end.Equal(start) {
				t.Errorf("end date %s is before start date %s", gotEnd, gotStart)
			}
		})
	}
}

func TestCalculateCycle_Contiguity(t *testing.T) {
	// Verify that throughout an entire year, cycle boundaries have zero gaps and zero overlaps
	startDay := 25
	curr := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endYear := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	for !curr.After(endYear) {
		start, end := CalculateCycle(startDay, curr)
		if curr.Before(start) || curr.After(end) {
			t.Fatalf("date %s does not lie in its calculated cycle [%s, %s]",
				curr.Format("2006-01-02"), start.Format("2006-01-02"), end.Format("2006-01-02"))
		}

		// Next cycle start should be exactly end + 1 day
		nextDay := end.AddDate(0, 0, 1)
		nextStart, _ := CalculateCycle(startDay, nextDay)
		if !nextStart.Equal(nextDay) {
			t.Fatalf("gap detected: end is %s, but next cycle starts at %s",
				end.Format("2006-01-02"), nextStart.Format("2006-01-02"))
		}

		curr = curr.AddDate(0, 0, 1)
	}
}
