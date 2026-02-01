package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"massrouter.ai/backend/internal/model"
)

type modelRepository struct {
	*GormRepository[model.Model]
}

func NewModelRepository(db *gorm.DB) ModelRepository {
	return &modelRepository{
		GormRepository: NewGormRepository[model.Model](db),
	}
}

func (r *modelRepository) FindByProviderAndName(ctx context.Context, providerID string, name string) (*model.Model, error) {
	var modelObj model.Model
	err := r.db.WithContext(ctx).
		Where("provider_id = ? AND name = ?", providerID, name).
		First(&modelObj).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find model by provider and name: %w", err)
	}
	return &modelObj, nil
}

func (r *modelRepository) FindActiveModels(ctx context.Context) ([]*model.Model, error) {
	var models []*model.Model
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Preload("Provider").
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find active models: %w", err)
	}
	return models, nil
}

func (r *modelRepository) FindByCategory(ctx context.Context, category string) ([]*model.Model, error) {
	var models []*model.Model
	err := r.db.WithContext(ctx).
		Where("category = ? AND is_active = ?", category, true).
		Preload("Provider").
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find models by category: %w", err)
	}
	return models, nil
}

func (r *modelRepository) SearchModels(ctx context.Context, query string, limit, offset int) ([]*model.Model, error) {
	var models []*model.Model
	
	dbQuery := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Preload("Provider")
	
	if query != "" {
		searchPattern := "%" + query + "%"
		dbQuery = dbQuery.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}
	
	if limit > 0 {
		dbQuery = dbQuery.Limit(limit)
	}
	if offset > 0 {
		dbQuery = dbQuery.Offset(offset)
	}
	
	err := dbQuery.Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to search models: %w", err)
	}
	return models, nil
}

func (r *modelRepository) UpdatePricing(ctx context.Context, modelID string, inputPrice, outputPrice float64) error {
	result := r.db.WithContext(ctx).Model(&model.Model{}).
		Where("id = ?", modelID).
		Updates(map[string]interface{}{
			"input_price":  inputPrice,
			"output_price": outputPrice,
		})
	
	if result.Error != nil {
		return fmt.Errorf("failed to update pricing: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("model not found")
	}
	return nil
}

func (r *modelRepository) UpdateStatus(ctx context.Context, modelID string, isActive bool) error {
	result := r.db.WithContext(ctx).Model(&model.Model{}).
		Where("id = ?", modelID).
		Update("is_active", isActive)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("model not found")
	}
	return nil
}
