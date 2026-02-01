package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"massrouter.ai/backend/internal/model"
)

type modelProviderRepository struct {
	*GormRepository[model.ModelProvider]
}

func NewModelProviderRepository(db *gorm.DB) ModelProviderRepository {
	return &modelProviderRepository{
		GormRepository: NewGormRepository[model.ModelProvider](db),
	}
}

func (r *modelProviderRepository) FindByName(ctx context.Context, name string) (*model.ModelProvider, error) {
	var provider model.ModelProvider
	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&provider).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider by name: %w", err)
	}
	return &provider, nil
}

func (r *modelProviderRepository) FindActiveProviders(ctx context.Context) ([]*model.ModelProvider, error) {
	var providers []*model.ModelProvider
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Find(&providers).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find active providers: %w", err)
	}
	return providers, nil
}

func (r *modelProviderRepository) UpdateAPIKey(ctx context.Context, providerID, apiKey string) error {
	result := r.db.WithContext(ctx).Model(&model.ModelProvider{}).
		Where("id = ?", providerID).
		Update("api_key", apiKey)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update API key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("provider not found")
	}
	return nil
}

func (r *modelProviderRepository) UpdateStatus(ctx context.Context, providerID, status string) error {
	result := r.db.WithContext(ctx).Model(&model.ModelProvider{}).
		Where("id = ?", providerID).
		Update("status", status)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("provider not found")
	}
	return nil
}
