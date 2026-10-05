package cycle

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service defines financial cycle operations.
type Service struct {
	repo Repository
}

// NewService creates a new cycle Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetCurrentCycle calculates the cycle window for a given target date (or today in UTC).
func (s *Service) GetCurrentCycle(ctx context.Context, userID uuid.UUID, targetDateStr string) (*CycleInfo, error) {
	targetDate, err := parseTargetDate(targetDateStr)
	if err != nil {
		return nil, err
	}

	startDay, err := s.repo.GetUserCycleStartDay(ctx, userID)
	if err != nil {
		return nil, err
	}

	cycleStart, cycleEnd := CalculateCycle(startDay, targetDate)
	info := buildCycleInfo(startDay, cycleStart, cycleEnd, targetDate)

	return &info, nil
}

// GetCycleSummary computes financial cycle metrics (income, expenses, net savings, transaction count).
func (s *Service) GetCycleSummary(ctx context.Context, userID uuid.UUID, targetDateStr string) (*CycleSummary, error) {
	cycleInfo, err := s.GetCurrentCycle(ctx, userID, targetDateStr)
	if err != nil {
		return nil, err
	}

	totalIncome, totalExpense, count, err := s.repo.GetCycleAggregates(ctx, userID, cycleInfo.StartDate, cycleInfo.EndDate)
	if err != nil {
		return nil, err
	}

	summary := &CycleSummary{
		Cycle:            *cycleInfo,
		TotalIncome:      totalIncome,
		TotalExpense:     totalExpense,
		NetSavings:       totalIncome - totalExpense,
		TransactionCount: count,
	}

	return summary, nil
}

func parseTargetDate(str string) (time.Time, error) {
	str = strings.TrimSpace(str)
	if str == "" {
		now := time.Now().UTC()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
	}

	t, err := time.Parse("2006-01-02", str)
	if err != nil {
		return time.Time{}, ErrInvalidDateFormat
	}

	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}

func buildCycleInfo(startDay int, cycleStart, cycleEnd, targetDate time.Time) CycleInfo {
	totalDays := int(cycleEnd.Sub(cycleStart).Hours()/24) + 1

	var dayOfCycle, daysRemaining int
	if targetDate.Before(cycleStart) {
		dayOfCycle = 1
		daysRemaining = totalDays
	} else if targetDate.After(cycleEnd) {
		dayOfCycle = totalDays
		daysRemaining = 0
	} else {
		dayOfCycle = int(targetDate.Sub(cycleStart).Hours()/24) + 1
		daysRemaining = int(cycleEnd.Sub(targetDate).Hours()/24)
	}

	return CycleInfo{
		CycleStartDay: startDay,
		StartDate:     cycleStart.Format("2006-01-02"),
		EndDate:       cycleEnd.Format("2006-01-02"),
		TotalDays:     totalDays,
		DayOfCycle:    dayOfCycle,
		DaysRemaining: daysRemaining,
	}
}
