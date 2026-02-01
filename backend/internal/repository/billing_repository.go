package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"massrouter.ai/backend/internal/model"
)

type billingRecordRepository struct {
	*GormRepository[model.BillingRecord]
}

func NewBillingRecordRepository(db *gorm.DB) BillingRecordRepository {
	return &billingRecordRepository{
		GormRepository: NewGormRepository[model.BillingRecord](db),
	}
}

func (r *billingRecordRepository) FindByUserID(ctx context.Context, userID string) ([]*model.BillingRecord, error) {
	var records []*model.BillingRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Model").
		Preload("Model.Provider").
		Order("created_at DESC").
		Find(&records).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find billing records by user: %w", err)
	}
	return records, nil
}

func (r *billingRecordRepository) FindByAPIKeyID(ctx context.Context, apiKeyID string) ([]*model.BillingRecord, error) {
	var records []*model.BillingRecord
	err := r.db.WithContext(ctx).
		Where("api_key_id = ?", apiKeyID).
		Preload("Model").
		Order("created_at DESC").
		Find(&records).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find billing records by API key: %w", err)
	}
	return records, nil
}

func (r *billingRecordRepository) FindByModelID(ctx context.Context, modelID string) ([]*model.BillingRecord, error) {
	var records []*model.BillingRecord
	err := r.db.WithContext(ctx).
		Where("model_id = ?", modelID).
		Order("created_at DESC").
		Find(&records).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find billing records by model: %w", err)
	}
	return records, nil
}

func (r *billingRecordRepository) GetUserUsage(ctx context.Context, userID string, startDate, endDate *time.Time) ([]*model.BillingRecord, error) {
	var records []*model.BillingRecord
	
	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Model")
	
	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", endDate)
	}
	
	err := query.Order("created_at ASC").Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user usage: %w", err)
	}
	return records, nil
}

func (r *billingRecordRepository) GetTotalCostByUser(ctx context.Context, userID string) (float64, error) {
	var totalCost float64
	
	err := r.db.WithContext(ctx).
		Model(&model.BillingRecord{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(cost), 0)").
		Scan(&totalCost).Error
	
	if err != nil {
		return 0, fmt.Errorf("failed to get total cost by user: %w", err)
	}
	return totalCost, nil
}

func (r *billingRecordRepository) GetDailyUsage(ctx context.Context, userID string, date time.Time) ([]*model.BillingRecord, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	var records []*model.BillingRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startOfDay, endOfDay).
		Preload("Model").
		Order("created_at ASC").
		Find(&records).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get daily usage: %w", err)
	}
	return records, nil
}
