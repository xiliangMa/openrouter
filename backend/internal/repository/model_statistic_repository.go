package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"massrouter.ai/backend/internal/model"
)

type modelStatisticRepository struct {
	*GormRepository[model.ModelStatistic]
}

func NewModelStatisticRepository(db *gorm.DB) ModelStatisticRepository {
	return &modelStatisticRepository{
		GormRepository: NewGormRepository[model.ModelStatistic](db),
	}
}

func (r *modelStatisticRepository) FindByModelAndDate(ctx context.Context, modelID string, date time.Time) (*model.ModelStatistic, error) {
	var statistic model.ModelStatistic
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	err := r.db.WithContext(ctx).
		Where("model_id = ? AND date = ?", modelID, startOfDay).
		First(&statistic).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find model statistic by model and date: %w", err)
	}
	return &statistic, nil
}

func (r *modelStatisticRepository) FindByModelID(ctx context.Context, modelID string, limit, offset int) ([]*model.ModelStatistic, error) {
	var statistics []*model.ModelStatistic

	query := r.db.WithContext(ctx).
		Where("model_id = ?", modelID).
		Order("date DESC")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Find(&statistics).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find model statistics by model ID: %w", err)
	}
	return statistics, nil
}

func (r *modelStatisticRepository) GetDailyStatistics(ctx context.Context, date time.Time) ([]*model.ModelStatistic, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	var statistics []*model.ModelStatistic
	err := r.db.WithContext(ctx).
		Where("date = ?", startOfDay).
		Preload("Model").
		Order("total_requests DESC").
		Find(&statistics).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get daily statistics: %w", err)
	}
	return statistics, nil
}

func (r *modelStatisticRepository) UpdateStatistics(ctx context.Context, modelID string, date time.Time, requests, tokens int, avgResponseTime, successRate float64) error {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	statistic, err := r.FindByModelAndDate(ctx, modelID, date)
	if err != nil {
		return fmt.Errorf("failed to find existing statistic: %w", err)
	}

	if statistic == nil {
		// Create new statistic
		statistic = &model.ModelStatistic{
			ModelID:         modelID,
			Date:            startOfDay,
			TotalRequests:   requests,
			TotalTokens:     tokens,
			AvgResponseTime: avgResponseTime,
			SuccessRate:     successRate,
			CreatedAt:       time.Now(),
		}

		if err := r.Create(ctx, statistic); err != nil {
			return fmt.Errorf("failed to create model statistic: %w", err)
		}
	} else {
		// Update existing statistic
		statistic.TotalRequests += requests
		statistic.TotalTokens += tokens

		// Calculate weighted average for response time
		totalRequests := statistic.TotalRequests
		previousRequests := totalRequests - requests

		if totalRequests > 0 {
			statistic.AvgResponseTime = (statistic.AvgResponseTime*float64(previousRequests) + avgResponseTime*float64(requests)) / float64(totalRequests)
			statistic.SuccessRate = (statistic.SuccessRate*float64(previousRequests) + successRate*float64(requests)) / float64(totalRequests)
		}

		if err := r.Update(ctx, statistic); err != nil {
			return fmt.Errorf("failed to update model statistic: %w", err)
		}
	}

	return nil
}

func (r *modelStatisticRepository) GetTopModels(ctx context.Context, limit int, startDate, endDate time.Time) ([]*model.ModelStatistic, error) {
	var statistics []*model.ModelStatistic

	query := r.db.WithContext(ctx).
		Select("model_id, SUM(total_requests) as total_requests, SUM(total_tokens) as total_tokens, AVG(avg_response_time) as avg_response_time, AVG(success_rate) as success_rate").
		Where("date >= ? AND date <= ?", startDate, endDate).
		Group("model_id").
		Order("SUM(total_requests) DESC").
		Preload("Model")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&statistics).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get top models: %w", err)
	}
	return statistics, nil
}
