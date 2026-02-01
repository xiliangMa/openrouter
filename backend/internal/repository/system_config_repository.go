package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"massrouter.ai/backend/internal/model"
)

type systemConfigRepository struct {
	*GormRepository[model.SystemConfig]
}

func NewSystemConfigRepository(db *gorm.DB) SystemConfigRepository {
	return &systemConfigRepository{
		GormRepository: NewGormRepository[model.SystemConfig](db),
	}
}

func (r *systemConfigRepository) FindByKey(ctx context.Context, key string) (*model.SystemConfig, error) {
	var config model.SystemConfig
	err := r.db.WithContext(ctx).
		Where("key = ?", key).
		First(&config).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find system config by key: %w", err)
	}
	return &config, nil
}

func (r *systemConfigRepository) FindPublicConfigs(ctx context.Context) ([]*model.SystemConfig, error) {
	var configs []*model.SystemConfig
	err := r.db.WithContext(ctx).
		Where("is_public = ?", true).
		Order("key ASC").
		Find(&configs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find public configs: %w", err)
	}
	return configs, nil
}

func (r *systemConfigRepository) UpdateValue(ctx context.Context, key, value string) error {
	config, err := r.FindByKey(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to find config: %w", err)
	}

	if config == nil {
		// Create new config
		config = &model.SystemConfig{
			Key:   key,
			Value: value,
		}

		if err := r.Create(ctx, config); err != nil {
			return fmt.Errorf("failed to create system config: %w", err)
		}
	} else {
		// Update existing config
		config.Value = value

		if err := r.Update(ctx, config); err != nil {
			return fmt.Errorf("failed to update system config: %w", err)
		}
	}

	return nil
}
