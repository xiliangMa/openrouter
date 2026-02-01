package service

import (
	"context"
	"fmt"
	"time"

	"massrouter.ai/backend/internal/model"
	"massrouter.ai/backend/internal/repository"
	"massrouter.ai/backend/pkg/utils"
)

type userService struct {
	userRepo    repository.UserRepository
	apiKeyRepo  repository.UserAPIKeyRepository
	billingRepo repository.BillingRecordRepository
	paymentRepo repository.PaymentRecordRepository
}

func NewUserService(
	userRepo repository.UserRepository,
	apiKeyRepo repository.UserAPIKeyRepository,
	billingRepo repository.BillingRecordRepository,
	paymentRepo repository.PaymentRecordRepository,
) UserService {
	return &userService{
		userRepo:    userRepo,
		apiKeyRepo:  apiKeyRepo,
		billingRepo: billingRepo,
		paymentRepo: paymentRepo,
	}
}

func (s *userService) GetProfile(ctx context.Context, userID string) (*UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	apiKeys, err := s.apiKeyRepo.FindActiveKeysByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get API keys: %w", err)
	}

	balance, err := s.getUserBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	totalUsage, err := s.billingRepo.GetTotalCostByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total usage: %w", err)
	}

	return &UserProfile{
		User:       user,
		APIKeys:    apiKeys,
		Balance:    balance,
		TotalUsage: totalUsage,
	}, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	if req.Username != "" && req.Username != user.Username {
		existingUser, err := s.userRepo.FindByUsername(ctx, req.Username)
		if err != nil {
			return fmt.Errorf("failed to check username: %w", err)
		}
		if existingUser != nil {
			return fmt.Errorf("username already taken")
		}
		user.Username = req.Username
	}

	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (s *userService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	if !utils.VerifyPassword(oldPassword, user.PasswordHash) {
		return fmt.Errorf("current password is incorrect")
	}

	newHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, newHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *userService) ListAPIKeys(ctx context.Context, userID string) ([]*model.UserAPIKey, error) {
	keys, err := s.apiKeyRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	return keys, nil
}

func (s *userService) CreateAPIKey(ctx context.Context, userID string, req *CreateAPIKeyRequest) (*model.UserAPIKey, error) {
	apiKey, err := utils.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	key := &model.UserAPIKey{
		UserID:      userID,
		Name:        req.Name,
		APIKey:      apiKey,
		Prefix:      apiKey[:10],
		Permissions: model.JSONB{"permissions": req.Permissions},
		RateLimit:   req.RateLimit,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if req.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)
		key.ExpiresAt = &expiresAt
	}

	if err := s.apiKeyRepo.Create(ctx, key); err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return key, nil
}

func (s *userService) DeleteAPIKey(ctx context.Context, userID, keyID string) error {
	key, err := s.apiKeyRepo.FindByID(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed to find API key: %w", err)
	}
	if key == nil {
		return fmt.Errorf("API key not found")
	}

	if key.UserID != userID {
		return fmt.Errorf("unauthorized to delete this API key")
	}

	if err := s.apiKeyRepo.RevokeKey(ctx, keyID); err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	return nil
}

func (s *userService) GetUserBalance(ctx context.Context, userID string) (*UserBalance, error) {
	totalPaid, err := s.paymentRepo.GetUserTotalPaid(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total paid: %w", err)
	}

	totalUsed, err := s.billingRepo.GetTotalCostByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total used: %w", err)
	}

	balance := totalPaid - totalUsed

	var lastPayment *time.Time
	var lastActivity *time.Time

	payments, err := s.paymentRepo.FindByUserID(ctx, userID)
	if err == nil && len(payments) > 0 {
		lastPayment = &payments[0].CreatedAt
	}

	billingRecords, err := s.billingRepo.FindByUserID(ctx, userID)
	if err == nil && len(billingRecords) > 0 {
		lastActivity = &billingRecords[0].CreatedAt
	}

	return &UserBalance{
		Balance:      balance,
		TotalPaid:    totalPaid,
		TotalUsed:    totalUsed,
		LastPayment:  lastPayment,
		LastActivity: lastActivity,
	}, nil
}

func (s *userService) GetUsageStatistics(ctx context.Context, userID string, startDate, endDate *time.Time) (*UsageStatistics, error) {
	records, err := s.billingRepo.GetUserUsage(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage records: %w", err)
	}

	usageMap := make(map[time.Time]*DailyUsage)
	modelUsageMap := make(map[string]*ModelUsage)
	var totalCost float64
	var totalTokens int64

	for _, record := range records {
		date := record.CreatedAt.Truncate(24 * time.Hour)
		if daily, exists := usageMap[date]; exists {
			daily.Cost += record.Cost
			daily.Tokens += int64(record.TotalTokens)
			daily.Requests++
		} else {
			usageMap[date] = &DailyUsage{
				Date:     date,
				Cost:     record.Cost,
				Tokens:   int64(record.TotalTokens),
				Requests: 1,
			}
		}

		modelKey := record.ModelID
		if modelUsage, exists := modelUsageMap[modelKey]; exists {
			modelUsage.Cost += record.Cost
			modelUsage.Tokens += int64(record.TotalTokens)
			modelUsage.Requests++
		} else {
			modelUsageMap[modelKey] = &ModelUsage{
				ModelID:   record.ModelID,
				ModelName: "Unknown",
				Cost:      record.Cost,
				Tokens:    int64(record.TotalTokens),
				Requests:  1,
			}
		}

		totalCost += record.Cost
		totalTokens += int64(record.TotalTokens)
	}

	dailyUsage := make([]*DailyUsage, 0, len(usageMap))
	for _, daily := range usageMap {
		dailyUsage = append(dailyUsage, daily)
	}

	topModels := make([]*ModelUsage, 0, len(modelUsageMap))
	for _, modelUsage := range modelUsageMap {
		topModels = append(topModels, modelUsage)
	}

	return &UsageStatistics{
		DailyUsage:  dailyUsage,
		TotalCost:   totalCost,
		TotalTokens: totalTokens,
		TopModels:   topModels[:min(5, len(topModels))],
	}, nil
}

func (s *userService) getUserBalance(ctx context.Context, userID string) (float64, error) {
	balance, err := s.GetUserBalance(ctx, userID)
	if err != nil {
		return 0, err
	}
	return balance.Balance, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
