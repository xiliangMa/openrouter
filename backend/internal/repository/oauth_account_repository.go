package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"massrouter.ai/backend/internal/model"
)

type oauthAccountRepository struct {
	*GormRepository[model.OAuthAccount]
}

func NewOAuthAccountRepository(db *gorm.DB) OAuthAccountRepository {
	return &oauthAccountRepository{
		GormRepository: NewGormRepository[model.OAuthAccount](db),
	}
}

func (r *oauthAccountRepository) FindByProviderAndUserID(ctx context.Context, providerID, userID string) (*model.OAuthAccount, error) {
	var account model.OAuthAccount
	err := r.db.WithContext(ctx).Where("provider_id = ? AND user_id = ?", providerID, userID).First(&account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find by provider and user: %w", err)
	}
	return &account, nil
}

func (r *oauthAccountRepository) FindByProviderUserID(ctx context.Context, providerID, providerUserID string) (*model.OAuthAccount, error) {
	var account model.OAuthAccount
	err := r.db.WithContext(ctx).Where("provider_id = ? AND provider_user_id = ?", providerID, providerUserID).First(&account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find by provider user id: %w", err)
	}
	return &account, nil
}

func (r *oauthAccountRepository) FindByUserID(ctx context.Context, userID string) ([]*model.OAuthAccount, error) {
	var accounts []*model.OAuthAccount
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&accounts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find by user id: %w", err)
	}
	return accounts, nil
}

func (r *oauthAccountRepository) UpdateTokens(ctx context.Context, accountID, accessToken, refreshToken string, expiresAt *time.Time) error {
	updateData := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}
	if expiresAt != nil {
		updateData["token_expires_at"] = expiresAt
	}

	result := r.db.WithContext(ctx).Model(&model.OAuthAccount{}).
		Where("id = ?", accountID).
		Updates(updateData)

	if result.Error != nil {
		return fmt.Errorf("failed to update tokens: %w", result.Error)
	}
	return nil
}
