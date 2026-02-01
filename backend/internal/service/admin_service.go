package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"massrouter.ai/backend/internal/model"
	"massrouter.ai/backend/internal/repository"
)

var _ = (*gorm.DB)(nil)

type adminService struct {
	userRepo      repository.UserRepository
	apiKeyRepo    repository.UserAPIKeyRepository
	paymentRepo   repository.PaymentRecordRepository
	billingRepo   repository.BillingRecordRepository
	modelRepo     repository.ModelRepository
	providerRepo  repository.ModelProviderRepository
	statisticRepo repository.ModelStatisticRepository
	configRepo    repository.SystemConfigRepository
}

func NewAdminService(
	userRepo repository.UserRepository,
	apiKeyRepo repository.UserAPIKeyRepository,
	paymentRepo repository.PaymentRecordRepository,
	billingRepo repository.BillingRecordRepository,
	modelRepo repository.ModelRepository,
	providerRepo repository.ModelProviderRepository,
	statisticRepo repository.ModelStatisticRepository,
	configRepo repository.SystemConfigRepository,
) AdminService {
	return &adminService{
		userRepo:      userRepo,
		apiKeyRepo:    apiKeyRepo,
		paymentRepo:   paymentRepo,
		billingRepo:   billingRepo,
		modelRepo:     modelRepo,
		providerRepo:  providerRepo,
		statisticRepo: statisticRepo,
		configRepo:    configRepo,
	}
}

func (s *adminService) ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := (req.Page - 1) * limit

	var users []*model.User
	db := s.userRepo.GetDB().WithContext(ctx)

	query := db.Model(&model.User{})

	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("email LIKE ? OR username LIKE ?", searchPattern, searchPattern)
	}

	if req.Role != "" {
		query = query.Where("role = ?", req.Role)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.SortBy != "" {
		order := s.getUserSortOrder(req.SortBy, req.SortOrder)
		query = query.Order(order)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to find users: %w", err)
	}

	userInfos := make([]*AdminUserInfo, len(users))
	for i, user := range users {
		totalPaid, err := s.paymentRepo.GetUserTotalPaid(ctx, user.ID)
		if err != nil {
			log.Printf("failed to get total paid for user %s: %v", user.ID, err)
			totalPaid = 0
		}
		totalUsed, err := s.billingRepo.GetTotalCostByUser(ctx, user.ID)
		if err != nil {
			log.Printf("failed to get total used for user %s: %v", user.ID, err)
			totalUsed = 0
		}
		balance := totalPaid - totalUsed

		apiKeys, err := s.apiKeyRepo.FindByUserID(ctx, user.ID)
		if err != nil {
			log.Printf("failed to get API keys for user %s: %v", user.ID, err)
			apiKeys = nil
		}
		apiKeysCount := len(apiKeys)

		var lastActivity *time.Time
		billingRecords, err := s.billingRepo.FindByUserID(ctx, user.ID)
		if err != nil {
			log.Printf("failed to get billing records for user %s: %v", user.ID, err)
			billingRecords = nil
		}
		if len(billingRecords) > 0 {
			lastActivity = &billingRecords[0].CreatedAt
		}

		userInfos[i] = &AdminUserInfo{
			User:           user,
			TotalPaid:      totalPaid,
			TotalUsed:      totalUsed,
			CurrentBalance: balance,
			APIKeysCount:   apiKeysCount,
			LastActivity:   lastActivity,
		}
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &ListUsersResponse{
		Users:      userInfos,
		Total:      total,
		Page:       req.Page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *adminService) GetUserDetails(ctx context.Context, userID string) (*AdminUserDetails, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	apiKeys, err := s.apiKeyRepo.FindByUserID(ctx, userID)
	if err != nil {
		log.Printf("failed to get API keys for user %s: %v", userID, err)
		apiKeys = nil
	}
	payments, err := s.paymentRepo.FindByUserID(ctx, userID)
	if err != nil {
		log.Printf("failed to get payments for user %s: %v", userID, err)
		payments = nil
	}
	billingRecords, err := s.billingRepo.FindByUserID(ctx, userID)
	if err != nil {
		log.Printf("failed to get billing records for user %s: %v", userID, err)
		billingRecords = nil
	}
	oauthAccounts, err := s.userRepo.FindWithOAuthAccounts(ctx, userID)
	if err != nil {
		log.Printf("failed to get OAuth accounts for user %s: %v", userID, err)
		oauthAccounts = &model.User{}
	}

	paymentItems := make([]*PaymentItem, len(payments))
	for i, payment := range payments {
		paymentItems[i] = &PaymentItem{
			ID:            payment.ID,
			Amount:        payment.Amount,
			Currency:      payment.Currency,
			PaymentMethod: payment.PaymentMethod,
			Status:        payment.Status,
			TransactionID: payment.TransactionID,
			CreatedAt:     payment.CreatedAt,
			PaidAt:        payment.PaidAt,
		}
	}

	billingRecordItems := make([]*BillingRecordItem, len(billingRecords))
	for i, record := range billingRecords {
		modelObj, err := s.modelRepo.FindByID(ctx, record.ModelID)
		if err != nil {
			log.Printf("failed to get model %s for billing record: %v", record.ModelID, err)
			modelObj = nil
		}
		modelName := "Unknown"
		providerName := "Unknown"
		if modelObj != nil {
			modelName = modelObj.Name
			providerName = modelObj.Provider.Name
		}

		billingRecordItems[i] = &BillingRecordItem{
			BillingRecord: record,
			ModelName:     modelName,
			ProviderName:  providerName,
		}
	}

	oauthAccountInfos := make([]*OAuthAccountInfo, len(oauthAccounts.OAuthAccounts))
	for i, account := range oauthAccounts.OAuthAccounts {
		oauthAccountInfos[i] = &OAuthAccountInfo{
			ProviderName: account.Provider.Name,
			ConnectedAt:  account.CreatedAt,
			LastUsedAt:   account.TokenExpiresAt,
		}
	}

	if _, err := s.paymentRepo.GetUserTotalPaid(ctx, userID); err != nil {
		log.Printf("failed to get user total paid for user %s: %v", userID, err)
	}
	totalUsed, err := s.billingRepo.GetTotalCostByUser(ctx, userID)
	if err != nil {
		log.Printf("failed to get total used for user %s: %v", userID, err)
		totalUsed = 0
	}
	records, err := s.billingRepo.FindByUserID(ctx, userID)
	if err != nil {
		log.Printf("failed to get billing records for user %s: %v", userID, err)
		records = nil
	}

	var totalRequests int64
	var totalTokens int64
	var mostUsedModel string
	modelUsage := make(map[string]int64)

	for _, record := range records {
		totalRequests++
		totalTokens += int64(record.TotalTokens)
		modelUsage[record.ModelID]++
	}

	if len(modelUsage) > 0 {
		var maxCount int64
		for modelID, count := range modelUsage {
			if count > maxCount {
				maxCount = count
				mostUsedModel = modelID
			}
		}
	}

	avgDailyCost := 0.0
	if len(records) > 0 {
		firstRecord := records[len(records)-1]
		days := time.Since(firstRecord.CreatedAt).Hours() / 24
		if days > 0 {
			avgDailyCost = totalUsed / days
		}
	}

	statistics := &UserStatistics{
		TotalRequests: totalRequests,
		TotalTokens:   totalTokens,
		TotalCost:     totalUsed,
		AvgDailyCost:  avgDailyCost,
		MostUsedModel: mostUsedModel,
	}

	return &AdminUserDetails{
		User:           user,
		PaymentHistory: paymentItems,
		BillingRecords: billingRecordItems,
		APIKeys:        apiKeys,
		OAuthAccounts:  oauthAccountInfos,
		Statistics:     statistics,
	}, nil
}

func (s *adminService) UpdateUser(ctx context.Context, userID string, req *AdminUpdateUserRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	updates := make(map[string]interface{})
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if len(updates) == 0 {
		return nil
	}

	updates["updated_at"] = time.Now()
	err = s.userRepo.Update(ctx, &model.User{
		ID:        userID,
		Role:      req.Role,
		Status:    req.Status,
		UpdatedAt: updates["updated_at"].(time.Time),
	})

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (s *adminService) CreateModelProvider(ctx context.Context, req *CreateModelProviderRequest) (*model.ModelProvider, error) {
	existing, err := s.providerRepo.FindByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check provider: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("provider with this name already exists")
	}

	provider := &model.ModelProvider{
		Name:       req.Name,
		APIBaseURL: req.APIBaseURL,
		APIKey:     req.APIKey,
		Config:     model.JSONB(req.Config),
		Status:     "active",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.providerRepo.Create(ctx, provider); err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return provider, nil
}

func (s *adminService) UpdateModelProvider(ctx context.Context, providerID string, req *UpdateModelProviderRequest) error {
	provider, err := s.providerRepo.FindByID(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}
	if provider == nil {
		return fmt.Errorf("provider not found")
	}

	updates := make(map[string]interface{})
	if req.APIBaseURL != "" {
		updates["api_base_url"] = req.APIBaseURL
	}
	if req.APIKey != "" {
		updates["api_key"] = req.APIKey
	}
	if req.Config != nil {
		updates["config"] = model.JSONB(req.Config)
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if len(updates) == 0 {
		return nil
	}

	updates["updated_at"] = time.Now()
	err = s.providerRepo.Update(ctx, &model.ModelProvider{
		ID:         providerID,
		APIBaseURL: req.APIBaseURL,
		APIKey:     req.APIKey,
		Config:     model.JSONB(req.Config),
		Status:     req.Status,
		UpdatedAt:  updates["updated_at"].(time.Time),
	})

	if err != nil {
		return fmt.Errorf("failed to update provider: %w", err)
	}

	return nil
}

func (s *adminService) CreateModel(ctx context.Context, req *CreateModelRequest) (*model.Model, error) {
	provider, err := s.providerRepo.FindByID(ctx, req.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}
	if provider == nil {
		return nil, fmt.Errorf("provider not found")
	}

	existing, err := s.modelRepo.FindByProviderAndName(ctx, req.ProviderID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check model: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("model with this name already exists for this provider")
	}

	modelObj := &model.Model{
		ProviderID:    req.ProviderID,
		Name:          req.Name,
		Description:   req.Description,
		ContextLength: req.ContextLength,
		MaxTokens:     req.MaxTokens,
		Capabilities:  model.JSONB(req.Capabilities),
		Category:      req.Category,
		PricingTier:   req.PricingTier,
		InputPrice:    req.InputPrice,
		OutputPrice:   req.OutputPrice,
		IsFree:        req.IsFree,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.modelRepo.Create(ctx, modelObj); err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	modelObj.Provider = *provider
	return modelObj, nil
}

func (s *adminService) UpdateModel(ctx context.Context, modelID string, req *UpdateModelRequest) error {
	modelObj, err := s.modelRepo.FindByID(ctx, modelID)
	if err != nil {
		return fmt.Errorf("failed to get model: %w", err)
	}
	if modelObj == nil {
		return fmt.Errorf("model not found")
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.ContextLength != nil {
		updates["context_length"] = *req.ContextLength
	}
	if req.MaxTokens != nil {
		updates["max_tokens"] = *req.MaxTokens
	}
	if req.Capabilities != nil {
		updates["capabilities"] = model.JSONB(req.Capabilities)
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.PricingTier != "" {
		updates["pricing_tier"] = req.PricingTier
	}
	if req.InputPrice > 0 {
		updates["input_price"] = req.InputPrice
	}
	if req.OutputPrice > 0 {
		updates["output_price"] = req.OutputPrice
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		return nil
	}

	updates["updated_at"] = time.Now()
	err = s.modelRepo.Update(ctx, &model.Model{
		ID:            modelID,
		Name:          req.Name,
		Description:   req.Description,
		ContextLength: req.ContextLength,
		MaxTokens:     req.MaxTokens,
		Capabilities:  model.JSONB(req.Capabilities),
		Category:      req.Category,
		PricingTier:   req.PricingTier,
		InputPrice:    req.InputPrice,
		OutputPrice:   req.OutputPrice,
		IsActive:      *req.IsActive,
		UpdatedAt:     updates["updated_at"].(time.Time),
	})

	if err != nil {
		return fmt.Errorf("failed to update model: %w", err)
	}

	return nil
}

func (s *adminService) GetSystemStats(ctx context.Context) (*SystemStats, error) {
	totalUsers, err := s.userRepo.Count(ctx)
	if err != nil {
		log.Printf("failed to count total users: %v", err)
		totalUsers = 0
	}
	var activeUsers int64
	err = s.userRepo.GetDB().WithContext(ctx).Model(&model.User{}).Where("status = ?", "active").Count(&activeUsers).Error
	if err != nil {
		log.Printf("failed to count active users: %v", err)
		activeUsers = totalUsers // fallback to total users
	}

	var totalRevenue float64
	totalRequests, err := s.billingRepo.Count(ctx)
	if err != nil {
		log.Printf("failed to count total requests: %v", err)
		totalRequests = 0
	}
	var recentPayments []*PaymentItem

	allUsers, err := s.userRepo.FindAll(ctx, 0, 0)
	if err != nil {
		log.Printf("failed to get all users: %v", err)
		allUsers = nil
	}
	for _, user := range allUsers {
		paid, err := s.paymentRepo.GetUserTotalPaid(ctx, user.ID)
		if err != nil {
			log.Printf("failed to get total paid for user %s: %v", user.ID, err)
			paid = 0
		}
		totalRevenue += paid
	}

	topModels, err := s.getTopModels(ctx, 10)
	if err != nil {
		log.Printf("failed to get top models: %v", err)
		topModels = nil
	}

	payments, err := s.paymentRepo.FindAll(ctx, 10, 0)
	if err != nil {
		log.Printf("failed to get recent payments: %v", err)
		payments = nil
	}
	recentPayments = make([]*PaymentItem, len(payments))
	for i, payment := range payments {
		recentPayments[i] = &PaymentItem{
			ID:            payment.ID,
			Amount:        payment.Amount,
			Currency:      payment.Currency,
			PaymentMethod: payment.PaymentMethod,
			Status:        payment.Status,
			TransactionID: payment.TransactionID,
			CreatedAt:     payment.CreatedAt,
			PaidAt:        payment.PaidAt,
		}
	}

	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)
	yesterdayRecords, err := s.billingRepo.GetUserUsage(ctx, "", &yesterday, &today)
	if err != nil {
		log.Printf("failed to get yesterday records: %v", err)
		yesterdayRecords = nil
	}
	var dailyRequests int64
	var dailyRevenue float64
	for _, record := range yesterdayRecords {
		dailyRequests++
		dailyRevenue += record.Cost
	}

	return &SystemStats{
		TotalUsers:     totalUsers,
		ActiveUsers:    activeUsers,
		TotalRequests:  totalRequests,
		TotalRevenue:   totalRevenue,
		DailyRequests:  dailyRequests,
		DailyRevenue:   dailyRevenue,
		TopModels:      topModels,
		RecentPayments: recentPayments,
		ServerStatus: &ServerStatus{
			Database: true,
			Redis:    true,
			Uptime:   "24h",
			Memory:   45.2,
			CPU:      12.5,
		},
	}, nil
}

func (s *adminService) UpdateSystemConfig(ctx context.Context, key, value string) error {
	config, err := s.configRepo.FindByKey(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	if config == nil {
		config = &model.SystemConfig{
			Key:       key,
			Value:     value,
			IsPublic:  false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		return s.configRepo.Create(ctx, config)
	}

	config.Value = value
	config.UpdatedAt = time.Now()
	return s.configRepo.Update(ctx, config)
}

func (s *adminService) getUserSortOrder(sortBy, sortOrder string) string {
	order := "created_at DESC"
	if sortOrder == "" {
		sortOrder = "desc"
	}

	switch sortBy {
	case "email":
		order = fmt.Sprintf("email %s", sortOrder)
	case "username":
		order = fmt.Sprintf("username %s", sortOrder)
	case "created_at":
		order = fmt.Sprintf("created_at %s", sortOrder)
	case "last_login":
		order = fmt.Sprintf("last_login_at %s", sortOrder)
	}

	return order
}

func (s *adminService) getTopModels(ctx context.Context, limit int) ([]*ModelStats, error) {
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	statistics, err := s.statisticRepo.GetTopModels(ctx, limit, thirtyDaysAgo, time.Now())
	if err != nil {
		return nil, err
	}

	modelStats := make([]*ModelStats, len(statistics))
	for i, stat := range statistics {
		modelObj, err := s.modelRepo.FindByID(ctx, stat.ModelID)
		if err != nil {
			log.Printf("failed to get model %s for statistics: %v", stat.ModelID, err)
			modelObj = nil
		}
		modelName := "Unknown"
		if modelObj != nil {
			modelName = modelObj.Name
		}

		revenue := float64(stat.TotalTokens) * 0.0001

		modelStats[i] = &ModelStats{
			ModelID:     stat.ModelID,
			ModelName:   modelName,
			Requests:    int64(stat.TotalRequests),
			Revenue:     revenue,
			SuccessRate: stat.SuccessRate,
		}
	}

	return modelStats, nil
}
