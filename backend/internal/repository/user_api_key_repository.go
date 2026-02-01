package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"massrouter.ai/backend/internal/model"
)

type userAPIKeyRepository struct {
	*GormRepository[model.UserAPIKey]
}

func NewUserAPIKeyRepository(db *gorm.DB) UserAPIKeyRepository {
	return &userAPIKeyRepository{
		GormRepository: NewGormRepository[model.UserAPIKey](db),
	}
}

func (r *userAPIKeyRepository) FindByAPIKey(ctx context.Context, apiKey string) (*model.UserAPIKey, error) {
	var key model.UserAPIKey
	err := r.db.WithContext(ctx).
		Where("api_key = ? AND is_active = ?", apiKey, true).
		Preload("User").
		First(&key).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find by API key: %w", err)
	}
	return &key, nil
}

func (r *userAPIKeyRepository) FindByUserID(ctx context.Context, userID string) ([]*model.UserAPIKey, error) {
	var keys []*model.UserAPIKey
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&keys).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find API keys by user: %w", err)
	}
	return keys, nil
}

func (r *userAPIKeyRepository) FindActiveKeysByUserID(ctx context.Context, userID string) ([]*model.UserAPIKey, error) {
	var keys []*model.UserAPIKey
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("created_at DESC").
		Find(&keys).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find active API keys by user: %w", err)
	}
	return keys, nil
}

func (r *userAPIKeyRepository) UpdateLastUsed(ctx context.Context, keyID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&model.UserAPIKey{}).
		Where("id = ?", keyID).
		Update("last_used_at", now)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update last used: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}
	return nil
}

func (r *userAPIKeyRepository) RevokeKey(ctx context.Context, keyID string) error {
	result := r.db.WithContext(ctx).Model(&model.UserAPIKey{}).
		Where("id = ?", keyID).
		Update("is_active", false)
	
	if result.Error != nil {
		return fmt.Errorf("failed to revoke key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}
	return nil
}

func (r *userAPIKeyRepository) ValidateKey(ctx context.Context, apiKey string) (*model.UserAPIKey, error) {
	key, err := r.FindByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, nil
	}
	
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		result := r.db.WithContext(ctx).Model(&model.UserAPIKey{}).
			Where("id = ?", key.ID).
			Update("is_active", false)
		
		if result.Error != nil {
			return nil, fmt.Errorf("failed to deactivate expired key: %w", result.Error)
		}
		return nil, nil
	}
	
	if err := r.UpdateLastUsed(ctx, key.ID); err != nil {
		return nil, fmt.Errorf("failed to update last used: %w", err)
	}
	
	return key, nil
}
