package service

import (
	"context"
	"fmt"
	"time"

	"massrouter.ai/backend/internal/model"
	"massrouter.ai/backend/internal/repository"
)

type quotaService struct {
	quotaRepo   repository.UserQuotaRepository
	usageRepo   repository.UserUsageRepository
	monthlyRepo repository.MonthlyUsageRepository
}

func NewQuotaService(
	quotaRepo repository.UserQuotaRepository,
	usageRepo repository.UserUsageRepository,
	monthlyRepo repository.MonthlyUsageRepository,
) QuotaService {
	return &quotaService{
		quotaRepo:   quotaRepo,
		usageRepo:   usageRepo,
		monthlyRepo: monthlyRepo,
	}
}

func (s *quotaService) GetUserQuota(ctx context.Context, userID string) (*model.UserQuota, error) {
	quota, err := s.quotaRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user quota: %w", err)
	}
	return quota, nil
}

func (s *quotaService) UpdateUserQuota(ctx context.Context, userID string, req *UpdateQuotaRequest) error {
	// Get existing quota
	quota, err := s.quotaRepo.FindByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get existing quota: %w", err)
	}

	// Update fields if provided
	if req.DailyRequestLimit != nil {
		quota.DailyRequestLimit = *req.DailyRequestLimit
	}
	if req.DailyTokenLimit != nil {
		quota.DailyTokenLimit = *req.DailyTokenLimit
	}
	if req.DailyCostLimit != nil {
		quota.DailyCostLimit = *req.DailyCostLimit
	}
	if req.MonthlyRequestLimit != nil {
		quota.MonthlyRequestLimit = *req.MonthlyRequestLimit
	}
	if req.MonthlyTokenLimit != nil {
		quota.MonthlyTokenLimit = *req.MonthlyTokenLimit
	}
	if req.MonthlyCostLimit != nil {
		quota.MonthlyCostLimit = *req.MonthlyCostLimit
	}
	if req.PerMinuteRateLimit != nil {
		quota.PerMinuteRateLimit = *req.PerMinuteRateLimit
	}
	if req.PerHourRateLimit != nil {
		quota.PerHourRateLimit = *req.PerHourRateLimit
	}
	if req.ResetDay != nil {
		quota.ResetDay = *req.ResetDay
	}
	if req.Timezone != "" {
		quota.Timezone = req.Timezone
	}
	if req.ModelLimits != nil {
		quota.ModelLimits = req.ModelLimits
	}
	if req.IsActive != nil {
		quota.IsActive = *req.IsActive
	}

	// Save updated quota
	if err := s.quotaRepo.UpdateQuota(ctx, userID, quota); err != nil {
		return fmt.Errorf("failed to update quota: %w", err)
	}
	return nil
}

func (s *quotaService) CheckQuota(ctx context.Context, userID string, tokens int, cost float64) (*QuotaCheckResult, error) {
	// Get user quota
	quota, err := s.quotaRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	// Check if quota is active
	if !quota.IsActive {
		return &QuotaCheckResult{
			Allowed: false,
			Reason:  "quota is not active",
		}, nil
	}

	// Get today's date and current month
	now := time.Now()
	yearMonth := now.Format("2006-01")

	// Get daily usage
	dailyUsage, err := s.usageRepo.FindByUserAndDate(ctx, userID, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily usage: %w", err)
	}

	// Get monthly usage
	monthlyUsage, err := s.monthlyRepo.FindByUserAndMonth(ctx, userID, yearMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly usage: %w", err)
	}

	// Calculate projected usage
	projectedDailyRequests := dailyUsage.RequestCount + 1
	projectedDailyTokens := dailyUsage.TokenCount + tokens
	projectedDailyCost := dailyUsage.TotalCost + cost

	projectedMonthlyRequests := monthlyUsage.RequestCount + 1
	projectedMonthlyTokens := monthlyUsage.TokenCount + tokens
	projectedMonthlyCost := monthlyUsage.TotalCost + cost

	// Check daily limits
	if projectedDailyRequests > quota.DailyRequestLimit {
		return &QuotaCheckResult{
			Allowed:              false,
			Reason:               "daily request limit exceeded",
			DailyRequests:        dailyUsage.RequestCount,
			DailyRequestsLimit:   quota.DailyRequestLimit,
			DailyTokens:          dailyUsage.TokenCount,
			DailyTokensLimit:     quota.DailyTokenLimit,
			DailyCost:            dailyUsage.TotalCost,
			DailyCostLimit:       quota.DailyCostLimit,
			MonthlyRequests:      monthlyUsage.RequestCount,
			MonthlyRequestsLimit: quota.MonthlyRequestLimit,
			MonthlyTokens:        monthlyUsage.TokenCount,
			MonthlyTokensLimit:   quota.MonthlyTokenLimit,
			MonthlyCost:          monthlyUsage.TotalCost,
			MonthlyCostLimit:     quota.MonthlyCostLimit,
			NextReset:            s.getNextDailyReset(now, quota.Timezone),
		}, nil
	}

	if projectedDailyTokens > quota.DailyTokenLimit {
		return &QuotaCheckResult{
			Allowed:              false,
			Reason:               "daily token limit exceeded",
			DailyRequests:        dailyUsage.RequestCount,
			DailyRequestsLimit:   quota.DailyRequestLimit,
			DailyTokens:          dailyUsage.TokenCount,
			DailyTokensLimit:     quota.DailyTokenLimit,
			DailyCost:            dailyUsage.TotalCost,
			DailyCostLimit:       quota.DailyCostLimit,
			MonthlyRequests:      monthlyUsage.RequestCount,
			MonthlyRequestsLimit: quota.MonthlyRequestLimit,
			MonthlyTokens:        monthlyUsage.TokenCount,
			MonthlyTokensLimit:   quota.MonthlyTokenLimit,
			MonthlyCost:          monthlyUsage.TotalCost,
			MonthlyCostLimit:     quota.MonthlyCostLimit,
			NextReset:            s.getNextDailyReset(now, quota.Timezone),
		}, nil
	}

	if projectedDailyCost > quota.DailyCostLimit {
		return &QuotaCheckResult{
			Allowed:              false,
			Reason:               "daily cost limit exceeded",
			DailyRequests:        dailyUsage.RequestCount,
			DailyRequestsLimit:   quota.DailyRequestLimit,
			DailyTokens:          dailyUsage.TokenCount,
			DailyTokensLimit:     quota.DailyTokenLimit,
			DailyCost:            dailyUsage.TotalCost,
			DailyCostLimit:       quota.DailyCostLimit,
			MonthlyRequests:      monthlyUsage.RequestCount,
			MonthlyRequestsLimit: quota.MonthlyRequestLimit,
			MonthlyTokens:        monthlyUsage.TokenCount,
			MonthlyTokensLimit:   quota.MonthlyTokenLimit,
			MonthlyCost:          monthlyUsage.TotalCost,
			MonthlyCostLimit:     quota.MonthlyCostLimit,
			NextReset:            s.getNextDailyReset(now, quota.Timezone),
		}, nil
	}

	// Check monthly limits
	if projectedMonthlyRequests > quota.MonthlyRequestLimit {
		return &QuotaCheckResult{
			Allowed:              false,
			Reason:               "monthly request limit exceeded",
			DailyRequests:        dailyUsage.RequestCount,
			DailyRequestsLimit:   quota.DailyRequestLimit,
			DailyTokens:          dailyUsage.TokenCount,
			DailyTokensLimit:     quota.DailyTokenLimit,
			DailyCost:            dailyUsage.TotalCost,
			DailyCostLimit:       quota.DailyCostLimit,
			MonthlyRequests:      monthlyUsage.RequestCount,
			MonthlyRequestsLimit: quota.MonthlyRequestLimit,
			MonthlyTokens:        monthlyUsage.TokenCount,
			MonthlyTokensLimit:   quota.MonthlyTokenLimit,
			MonthlyCost:          monthlyUsage.TotalCost,
			MonthlyCostLimit:     quota.MonthlyCostLimit,
			NextReset:            s.getNextMonthlyReset(now, quota.ResetDay, quota.Timezone),
		}, nil
	}

	if projectedMonthlyTokens > quota.MonthlyTokenLimit {
		return &QuotaCheckResult{
			Allowed:              false,
			Reason:               "monthly token limit exceeded",
			DailyRequests:        dailyUsage.RequestCount,
			DailyRequestsLimit:   quota.DailyRequestLimit,
			DailyTokens:          dailyUsage.TokenCount,
			DailyTokensLimit:     quota.DailyTokenLimit,
			DailyCost:            dailyUsage.TotalCost,
			DailyCostLimit:       quota.DailyCostLimit,
			MonthlyRequests:      monthlyUsage.RequestCount,
			MonthlyRequestsLimit: quota.MonthlyRequestLimit,
			MonthlyTokens:        monthlyUsage.TokenCount,
			MonthlyTokensLimit:   quota.MonthlyTokenLimit,
			MonthlyCost:          monthlyUsage.TotalCost,
			MonthlyCostLimit:     quota.MonthlyCostLimit,
			NextReset:            s.getNextMonthlyReset(now, quota.ResetDay, quota.Timezone),
		}, nil
	}

	if projectedMonthlyCost > quota.MonthlyCostLimit {
		return &QuotaCheckResult{
			Allowed:              false,
			Reason:               "monthly cost limit exceeded",
			DailyRequests:        dailyUsage.RequestCount,
			DailyRequestsLimit:   quota.DailyRequestLimit,
			DailyTokens:          dailyUsage.TokenCount,
			DailyTokensLimit:     quota.DailyTokenLimit,
			DailyCost:            dailyUsage.TotalCost,
			DailyCostLimit:       quota.DailyCostLimit,
			MonthlyRequests:      monthlyUsage.RequestCount,
			MonthlyRequestsLimit: quota.MonthlyRequestLimit,
			MonthlyTokens:        monthlyUsage.TokenCount,
			MonthlyTokensLimit:   quota.MonthlyTokenLimit,
			MonthlyCost:          monthlyUsage.TotalCost,
			MonthlyCostLimit:     quota.MonthlyCostLimit,
			NextReset:            s.getNextMonthlyReset(now, quota.ResetDay, quota.Timezone),
		}, nil
	}

	// All checks passed
	return &QuotaCheckResult{
		Allowed:              true,
		DailyRequests:        dailyUsage.RequestCount,
		DailyRequestsLimit:   quota.DailyRequestLimit,
		DailyTokens:          dailyUsage.TokenCount,
		DailyTokensLimit:     quota.DailyTokenLimit,
		DailyCost:            dailyUsage.TotalCost,
		DailyCostLimit:       quota.DailyCostLimit,
		MonthlyRequests:      monthlyUsage.RequestCount,
		MonthlyRequestsLimit: quota.MonthlyRequestLimit,
		MonthlyTokens:        monthlyUsage.TokenCount,
		MonthlyTokensLimit:   quota.MonthlyTokenLimit,
		MonthlyCost:          monthlyUsage.TotalCost,
		MonthlyCostLimit:     quota.MonthlyCostLimit,
		NextReset:            s.getNextDailyReset(now, quota.Timezone),
	}, nil
}

func (s *quotaService) RecordUsage(ctx context.Context, userID string, modelID string, tokens int, cost float64) error {
	now := time.Now()
	yearMonth := now.Format("2006-01")

	// Prepare model usage data
	modelUsage := model.JSONB{
		modelID: model.JSONB{
			"requests": 1,
			"tokens":   tokens,
			"cost":     cost,
		},
	}

	// Update daily usage
	if err := s.usageRepo.IncrementUsage(ctx, userID, now, 1, tokens, cost, modelUsage); err != nil {
		return fmt.Errorf("failed to update daily usage: %w", err)
	}

	// Update monthly usage
	if err := s.monthlyRepo.IncrementUsage(ctx, userID, yearMonth, 1, tokens, cost, modelUsage); err != nil {
		return fmt.Errorf("failed to update monthly usage: %w", err)
	}

	// Check if limits are exceeded and update is_exceeded flag
	quota, err := s.quotaRepo.FindByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get quota for exceeded check: %w", err)
	}

	dailyUsage, err := s.usageRepo.FindByUserAndDate(ctx, userID, now)
	if err != nil {
		return fmt.Errorf("failed to get daily usage for exceeded check: %w", err)
	}

	monthlyUsage, err := s.monthlyRepo.FindByUserAndMonth(ctx, userID, yearMonth)
	if err != nil {
		return fmt.Errorf("failed to get monthly usage for exceeded check: %w", err)
	}

	// Check if any limit is exceeded
	isDailyExceeded := dailyUsage.RequestCount > quota.DailyRequestLimit ||
		dailyUsage.TokenCount > quota.DailyTokenLimit ||
		dailyUsage.TotalCost > quota.DailyCostLimit

	isMonthlyExceeded := monthlyUsage.RequestCount > quota.MonthlyRequestLimit ||
		monthlyUsage.TokenCount > quota.MonthlyTokenLimit ||
		monthlyUsage.TotalCost > quota.MonthlyCostLimit

	// Update exceeded flags if needed
	if isDailyExceeded && !dailyUsage.IsExceeded {
		dailyUsage.IsExceeded = true
		if err := s.usageRepo.Update(ctx, dailyUsage); err != nil {
			return fmt.Errorf("failed to update daily exceeded flag: %w", err)
		}
	}

	if isMonthlyExceeded && !monthlyUsage.IsExceeded {
		monthlyUsage.IsExceeded = true
		if err := s.monthlyRepo.Update(ctx, monthlyUsage); err != nil {
			return fmt.Errorf("failed to update monthly exceeded flag: %w", err)
		}
	}

	return nil
}

func (s *quotaService) GetDailyUsage(ctx context.Context, userID string, date time.Time) (*model.UserUsage, error) {
	usage, err := s.usageRepo.FindByUserAndDate(ctx, userID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily usage: %w", err)
	}
	return usage, nil
}

func (s *quotaService) GetMonthlyUsage(ctx context.Context, userID string, yearMonth string) (*model.MonthlyUsage, error) {
	usage, err := s.monthlyRepo.FindByUserAndMonth(ctx, userID, yearMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly usage: %w", err)
	}
	return usage, nil
}

func (s *quotaService) GetUsageHistory(ctx context.Context, userID string, startDate, endDate time.Time) ([]*model.UserUsage, error) {
	usage, err := s.usageRepo.GetUsageForPeriod(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage history: %w", err)
	}
	return usage, nil
}

func (s *quotaService) ResetDailyQuota(ctx context.Context, userID string) error {
	// Reset today's usage
	now := time.Now()

	// Get today's usage record
	dailyUsage, err := s.usageRepo.FindByUserAndDate(ctx, userID, now)
	if err != nil {
		return fmt.Errorf("failed to get daily usage: %w", err)
	}

	// Reset the usage
	dailyUsage.RequestCount = 0
	dailyUsage.TokenCount = 0
	dailyUsage.TotalCost = 0
	dailyUsage.ModelUsage = model.JSONB{}
	dailyUsage.IsExceeded = false

	if err := s.usageRepo.Update(ctx, dailyUsage); err != nil {
		return fmt.Errorf("failed to reset daily quota: %w", err)
	}
	return nil
}

func (s *quotaService) ResetMonthlyQuota(ctx context.Context, userID string) error {
	// Reset current month's usage
	now := time.Now()
	yearMonth := now.Format("2006-01")

	// Get current month's usage record
	monthlyUsage, err := s.monthlyRepo.FindByUserAndMonth(ctx, userID, yearMonth)
	if err != nil {
		return fmt.Errorf("failed to get monthly usage: %w", err)
	}

	// Reset the usage
	monthlyUsage.RequestCount = 0
	monthlyUsage.TokenCount = 0
	monthlyUsage.TotalCost = 0
	monthlyUsage.ModelUsage = model.JSONB{}
	monthlyUsage.IsExceeded = false

	if err := s.monthlyRepo.Update(ctx, monthlyUsage); err != nil {
		return fmt.Errorf("failed to reset monthly quota: %w", err)
	}
	return nil
}

func (s *quotaService) getNextDailyReset(now time.Time, timezone string) time.Time {
	// For simplicity, assume reset at midnight in the specified timezone
	// In production, you'd use time.LoadLocation to handle timezone correctly
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		// Default to UTC if timezone is invalid
		loc = time.UTC
	}

	// Get tomorrow's date at midnight
	tomorrow := now.In(loc).Add(24 * time.Hour)
	nextReset := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, loc)
	return nextReset
}

func (s *quotaService) getNextMonthlyReset(now time.Time, resetDay int, timezone string) time.Time {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	current := now.In(loc)
	year := current.Year()
	month := current.Month()

	// Adjust reset day if it's beyond the last day of the month
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
	actualResetDay := resetDay
	if actualResetDay > lastDay {
		actualResetDay = lastDay
	}

	// Calculate next reset date
	resetDate := time.Date(year, month, actualResetDay, 0, 0, 0, 0, loc)

	// If reset date has already passed this month, move to next month
	if current.After(resetDate) || current.Equal(resetDate) {
		// Move to next month
		month++
		if month > 12 {
			month = 1
			year++
		}

		// Recalculate last day for next month
		lastDay = time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
		actualResetDay = resetDay
		if actualResetDay > lastDay {
			actualResetDay = lastDay
		}

		resetDate = time.Date(year, month, actualResetDay, 0, 0, 0, 0, loc)
	}

	return resetDate
}
