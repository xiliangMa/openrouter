package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"massrouter.ai/backend/internal/model"
)

type userQuotaRepository struct {
	*GormRepository[model.UserQuota]
}

func NewUserQuotaRepository(db *gorm.DB) UserQuotaRepository {
	return &userQuotaRepository{
		GormRepository: NewGormRepository[model.UserQuota](db),
	}
}

func (r *userQuotaRepository) FindByUserID(ctx context.Context, userID string) (*model.UserQuota, error) {
	var quota model.UserQuota
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&quota).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default quota if not exists
			defaultQuota := model.DefaultQuota()
			defaultQuota.UserID = userID
			if err := r.db.WithContext(ctx).Create(&defaultQuota).Error; err != nil {
				return nil, fmt.Errorf("failed to create default quota: %w", err)
			}
			return &defaultQuota, nil
		}
		return nil, fmt.Errorf("failed to find quota by user ID: %w", err)
	}
	return &quota, nil
}

func (r *userQuotaRepository) UpdateQuota(ctx context.Context, userID string, quota *model.UserQuota) error {
	// Ensure the quota belongs to the user
	quota.UserID = userID

	// Use Upsert to create or update
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"daily_request_limit",
			"daily_token_limit",
			"daily_cost_limit",
			"monthly_request_limit",
			"monthly_token_limit",
			"monthly_cost_limit",
			"model_limits",
			"per_minute_rate_limit",
			"per_hour_rate_limit",
			"reset_day",
			"timezone",
			"is_active",
			"updated_at",
		}),
	}).Create(quota).Error

	if err != nil {
		return fmt.Errorf("failed to update quota: %w", err)
	}
	return nil
}

func (r *userQuotaRepository) UpdateModelLimits(ctx context.Context, userID string, modelLimits model.JSONB) error {
	result := r.db.WithContext(ctx).Model(&model.UserQuota{}).
		Where("user_id = ?", userID).
		Update("model_limits", modelLimits)

	if result.Error != nil {
		return fmt.Errorf("failed to update model limits: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("quota not found")
	}
	return nil
}

func (r *userQuotaRepository) FindActiveQuotas(ctx context.Context) ([]*model.UserQuota, error) {
	var quotas []*model.UserQuota
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Preload("User").
		Find(&quotas).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find active quotas: %w", err)
	}
	return quotas, nil
}

type userUsageRepository struct {
	*GormRepository[model.UserUsage]
}

func NewUserUsageRepository(db *gorm.DB) UserUsageRepository {
	return &userUsageRepository{
		GormRepository: NewGormRepository[model.UserUsage](db),
	}
}

func (r *userUsageRepository) FindByUserAndDate(ctx context.Context, userID string, date time.Time) (*model.UserUsage, error) {
	var usage model.UserUsage
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND date = ?", userID, date.Format("2006-01-02")).
		First(&usage).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create initial usage record if not exists
			usage = model.UserUsage{
				UserID:     userID,
				Date:       date,
				ModelUsage: model.JSONB{},
			}
			if err := r.db.WithContext(ctx).Create(&usage).Error; err != nil {
				return nil, fmt.Errorf("failed to create usage record: %w", err)
			}
			return &usage, nil
		}
		return nil, fmt.Errorf("failed to find usage by user and date: %w", err)
	}
	return &usage, nil
}

func (r *userUsageRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.UserUsage, error) {
	var usage []*model.UserUsage
	query := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("date DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&usage).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find usage by user ID: %w", err)
	}
	return usage, nil
}

func (r *userUsageRepository) IncrementUsage(ctx context.Context, userID string, date time.Time, requests, tokens int, cost float64, modelUsage model.JSONB) error {
	// Use PostgreSQL-specific UPDATE with atomic increment
	result := r.db.WithContext(ctx).Exec(`
		INSERT INTO user_usage (user_id, date, request_count, token_count, total_cost, model_usage, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON CONFLICT (user_id, date) 
		DO UPDATE SET 
			request_count = user_usage.request_count + EXCLUDED.request_count,
			token_count = user_usage.token_count + EXCLUDED.token_count,
			total_cost = user_usage.total_cost + EXCLUDED.total_cost,
			model_usage = user_usage.model_usage || EXCLUDED.model_usage,
			updated_at = NOW()
	`, userID, date.Format("2006-01-02"), requests, tokens, cost, modelUsage)

	if result.Error != nil {
		return fmt.Errorf("failed to increment usage: %w", result.Error)
	}
	return nil
}

func (r *userUsageRepository) GetUsageForPeriod(ctx context.Context, userID string, startDate, endDate time.Time) ([]*model.UserUsage, error) {
	var usage []*model.UserUsage
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Order("date ASC").
		Find(&usage).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get usage for period: %w", err)
	}
	return usage, nil
}

func (r *userUsageRepository) ResetDailyUsage(ctx context.Context, date time.Time) error {
	result := r.db.WithContext(ctx).Exec(`
		UPDATE user_usage 
		SET request_count = 0, token_count = 0, total_cost = 0, model_usage = '{}', is_exceeded = false, updated_at = NOW()
		WHERE date = ?
	`, date.Format("2006-01-02"))

	if result.Error != nil {
		return fmt.Errorf("failed to reset daily usage: %w", result.Error)
	}
	return nil
}

func (r *userUsageRepository) FindExceededUsage(ctx context.Context, date time.Time) ([]*model.UserUsage, error) {
	var usage []*model.UserUsage
	err := r.db.WithContext(ctx).
		Where("date = ? AND is_exceeded = ?", date.Format("2006-01-02"), true).
		Preload("User").
		Find(&usage).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find exceeded usage: %w", err)
	}
	return usage, nil
}

type monthlyUsageRepository struct {
	*GormRepository[model.MonthlyUsage]
}

func NewMonthlyUsageRepository(db *gorm.DB) MonthlyUsageRepository {
	return &monthlyUsageRepository{
		GormRepository: NewGormRepository[model.MonthlyUsage](db),
	}
}

func (r *monthlyUsageRepository) FindByUserAndMonth(ctx context.Context, userID string, yearMonth string) (*model.MonthlyUsage, error) {
	var usage model.MonthlyUsage
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND year_month = ?", userID, yearMonth).
		First(&usage).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create initial monthly usage record if not exists
			usage = model.MonthlyUsage{
				UserID:     userID,
				YearMonth:  yearMonth,
				ModelUsage: model.JSONB{},
			}
			if err := r.db.WithContext(ctx).Create(&usage).Error; err != nil {
				return nil, fmt.Errorf("failed to create monthly usage record: %w", err)
			}
			return &usage, nil
		}
		return nil, fmt.Errorf("failed to find monthly usage by user and month: %w", err)
	}
	return &usage, nil
}

func (r *monthlyUsageRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.MonthlyUsage, error) {
	var usage []*model.MonthlyUsage
	query := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("year_month DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&usage).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find monthly usage by user ID: %w", err)
	}
	return usage, nil
}

func (r *monthlyUsageRepository) IncrementUsage(ctx context.Context, userID string, yearMonth string, requests, tokens int, cost float64, modelUsage model.JSONB) error {
	// Use PostgreSQL-specific UPDATE with atomic increment
	result := r.db.WithContext(ctx).Exec(`
		INSERT INTO monthly_usage (user_id, year_month, request_count, token_count, total_cost, model_usage, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON CONFLICT (user_id, year_month) 
		DO UPDATE SET 
			request_count = monthly_usage.request_count + EXCLUDED.request_count,
			token_count = monthly_usage.token_count + EXCLUDED.token_count,
			total_cost = monthly_usage.total_cost + EXCLUDED.total_cost,
			model_usage = monthly_usage.model_usage || EXCLUDED.model_usage,
			updated_at = NOW()
	`, userID, yearMonth, requests, tokens, cost, modelUsage)

	if result.Error != nil {
		return fmt.Errorf("failed to increment monthly usage: %w", result.Error)
	}
	return nil
}

func (r *monthlyUsageRepository) GetMonthlySummary(ctx context.Context, yearMonth string) ([]*model.MonthlyUsage, error) {
	var usage []*model.MonthlyUsage
	err := r.db.WithContext(ctx).
		Where("year_month = ?", yearMonth).
		Preload("User").
		Find(&usage).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get monthly summary: %w", err)
	}
	return usage, nil
}

func (r *monthlyUsageRepository) ResetMonthlyUsage(ctx context.Context, yearMonth string) error {
	result := r.db.WithContext(ctx).Exec(`
		UPDATE monthly_usage 
		SET request_count = 0, token_count = 0, total_cost = 0, model_usage = '{}', is_exceeded = false, updated_at = NOW()
		WHERE year_month = ?
	`, yearMonth)

	if result.Error != nil {
		return fmt.Errorf("failed to reset monthly usage: %w", result.Error)
	}
	return nil
}
