package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"massrouter.ai/backend/internal/model"
)

type oauthProviderRepository struct {
	*GormRepository[model.OAuthProvider]
}

func NewOAuthProviderRepository(db *gorm.DB) OAuthProviderRepository {
	return &oauthProviderRepository{
		GormRepository: NewGormRepository[model.OAuthProvider](db),
	}
}

func (r *oauthProviderRepository) FindByName(ctx context.Context, name string) (*model.OAuthProvider, error) {
	var provider model.OAuthProvider
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&provider).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find by name: %w", err)
	}
	return &provider, nil
}

func (r *oauthProviderRepository) FindEnabledProviders(ctx context.Context) ([]*model.OAuthProvider, error) {
	var providers []*model.OAuthProvider
	err := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&providers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find enabled providers: %w", err)
	}
	return providers, nil
}
