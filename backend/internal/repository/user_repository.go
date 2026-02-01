package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"massrouter.ai/backend/internal/model"
)

type userRepository struct {
	*GormRepository[model.User]
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		GormRepository: NewGormRepository[model.User](db),
	}
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find by email: %w", err)
	}
	return &user, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find by username: %w", err)
	}
	return &user, nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	result := r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		Update("password_hash", passwordHash)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update password: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		Update("last_login_at", now)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update last login: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r *userRepository) UpdateStatus(ctx context.Context, userID, status string) error {
	result := r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		Update("status", status)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r *userRepository) FindWithAPIKeys(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Preload("APIKeys").
		Where("id = ?", userID).
		First(&user).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user with API keys: %w", err)
	}
	return &user, nil
}

func (r *userRepository) FindWithOAuthAccounts(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Preload("OAuthAccounts").
		Preload("OAuthAccounts.Provider").
		Where("id = ?", userID).
		First(&user).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user with OAuth accounts: %w", err)
	}
	return &user, nil
}
