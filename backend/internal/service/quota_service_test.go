package service

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"
	"massrouter.ai/backend/internal/model"
)

// Mock repositories for testing
type mockUserQuotaRepository struct {
	quotas map[string]*model.UserQuota
}

func (m *mockUserQuotaRepository) FindByUserID(ctx context.Context, userID string) (*model.UserQuota, error) {
	quota, exists := m.quotas[userID]
	if !exists {
		// Return default quota
		quota = &model.UserQuota{
			UserID:              userID,
			DailyRequestLimit:   100,
			DailyTokenLimit:     10000,
			DailyCostLimit:      10.0,
			MonthlyRequestLimit: 1000,
			MonthlyTokenLimit:   100000,
			MonthlyCostLimit:    100.0,
			PerMinuteRateLimit:  10,
			PerHourRateLimit:    100,
			ResetDay:            1,
			Timezone:            "UTC",
			IsActive:            true,
		}
		m.quotas[userID] = quota
	}
	return quota, nil
}

func (m *mockUserQuotaRepository) UpdateQuota(ctx context.Context, userID string, quota *model.UserQuota) error {
	m.quotas[userID] = quota
	return nil
}

func (m *mockUserQuotaRepository) CreateDefaultQuota(ctx context.Context, userID string) error {
	m.quotas[userID] = &model.UserQuota{
		UserID:              userID,
		DailyRequestLimit:   100,
		DailyTokenLimit:     10000,
		DailyCostLimit:      10.0,
		MonthlyRequestLimit: 1000,
		MonthlyTokenLimit:   100000,
		MonthlyCostLimit:    100.0,
		PerMinuteRateLimit:  10,
		PerHourRateLimit:    100,
		ResetDay:            1,
		Timezone:            "UTC",
		IsActive:            true,
	}
	return nil
}

// BaseRepository methods for mockUserQuotaRepository
func (m *mockUserQuotaRepository) FindByID(ctx context.Context, id string) (*model.UserQuota, error) {
	// Not implemented for quota repository
	return nil, nil
}

func (m *mockUserQuotaRepository) FindAll(ctx context.Context, limit, offset int) ([]*model.UserQuota, error) {
	var result []*model.UserQuota
	for _, quota := range m.quotas {
		result = append(result, quota)
	}
	return result, nil
}

func (m *mockUserQuotaRepository) Create(ctx context.Context, entity *model.UserQuota) error {
	m.quotas[entity.UserID] = entity
	return nil
}

func (m *mockUserQuotaRepository) Update(ctx context.Context, entity *model.UserQuota) error {
	m.quotas[entity.UserID] = entity
	return nil
}

func (m *mockUserQuotaRepository) Delete(ctx context.Context, id string) error {
	delete(m.quotas, id)
	return nil
}

func (m *mockUserQuotaRepository) Count(ctx context.Context) (int64, error) {
	return int64(len(m.quotas)), nil
}

func (m *mockUserQuotaRepository) Exists(ctx context.Context, id string) (bool, error) {
	_, exists := m.quotas[id]
	return exists, nil
}

func (m *mockUserQuotaRepository) GetDB() *gorm.DB {
	return nil
}

func (m *mockUserQuotaRepository) UpdateModelLimits(ctx context.Context, userID string, modelLimits model.JSONB) error {
	quota, exists := m.quotas[userID]
	if !exists {
		quota = &model.UserQuota{
			UserID: userID,
		}
		m.quotas[userID] = quota
	}
	quota.ModelLimits = modelLimits
	return nil
}

func (m *mockUserQuotaRepository) FindActiveQuotas(ctx context.Context) ([]*model.UserQuota, error) {
	var result []*model.UserQuota
	for _, quota := range m.quotas {
		if quota.IsActive {
			result = append(result, quota)
		}
	}
	return result, nil
}

type mockUserUsageRepository struct {
	dailyUsage map[string]*model.UserUsage // key: "userID_date"
}

func (m *mockUserUsageRepository) FindByUserAndDate(ctx context.Context, userID string, date time.Time) (*model.UserUsage, error) {
	key := userID + "_" + date.Format("2006-01-02")
	usage, exists := m.dailyUsage[key]
	if !exists {
		usage = &model.UserUsage{
			UserID:       userID,
			Date:         date,
			RequestCount: 0,
			TokenCount:   0,
			TotalCost:    0,
			ModelUsage:   model.JSONB{},
			IsExceeded:   false,
		}
		m.dailyUsage[key] = usage
	}
	return usage, nil
}

func (m *mockUserUsageRepository) IncrementUsage(ctx context.Context, userID string, date time.Time, requests, tokens int, cost float64, modelUsage model.JSONB) error {
	key := userID + "_" + date.Format("2006-01-02")
	usage, exists := m.dailyUsage[key]
	if !exists {
		usage = &model.UserUsage{
			UserID:       userID,
			Date:         date,
			RequestCount: 0,
			TokenCount:   0,
			TotalCost:    0,
			ModelUsage:   model.JSONB{},
			IsExceeded:   false,
		}
		m.dailyUsage[key] = usage
	}

	usage.RequestCount += requests
	usage.TokenCount += tokens
	usage.TotalCost += cost

	// Merge model usage - accumulate values instead of overwriting
	if usage.ModelUsage == nil {
		usage.ModelUsage = model.JSONB{}
	}

	for modelID, newData := range modelUsage {
		if newDataMap, ok := newData.(model.JSONB); ok {
			if existingData, exists := usage.ModelUsage[modelID]; exists {
				if existingMap, ok := existingData.(model.JSONB); ok {
					// Merge existing and new data
					merged := model.JSONB{}

					// Copy existing values
					for k, v := range existingMap {
						merged[k] = v
					}

					// Add new values (accumulate where appropriate)
					for k, newVal := range newDataMap {
						switch k {
						case "requests", "tokens":
							// Accumulate integers
							existingVal := 0
							if v, ok := merged[k]; ok {
								switch val := v.(type) {
								case int:
									existingVal = val
								case float64:
									existingVal = int(val)
								}
							}

							var newValInt int
							switch val := newVal.(type) {
							case int:
								newValInt = val
							case float64:
								newValInt = int(val)
							default:
								newValInt = 0
							}

							merged[k] = existingVal + newValInt
						case "cost":
							// Accumulate floats
							existingVal := 0.0
							if v, ok := merged[k]; ok {
								switch val := v.(type) {
								case float64:
									existingVal = val
								case int:
									existingVal = float64(val)
								}
							}

							var newValFloat float64
							switch val := newVal.(type) {
							case float64:
								newValFloat = val
							case int:
								newValFloat = float64(val)
							default:
								newValFloat = 0.0
							}

							merged[k] = existingVal + newValFloat
						default:
							merged[k] = newVal
						}
					}

					usage.ModelUsage[modelID] = merged
				}
			} else {
				// No existing data for this model, just set it
				usage.ModelUsage[modelID] = newDataMap
			}
		}
	}

	return nil
}

func (m *mockUserUsageRepository) Update(ctx context.Context, usage *model.UserUsage) error {
	key := usage.UserID + "_" + usage.Date.Format("2006-01-02")
	m.dailyUsage[key] = usage
	return nil
}

func (m *mockUserUsageRepository) GetUsageForPeriod(ctx context.Context, userID string, startDate, endDate time.Time) ([]*model.UserUsage, error) {
	var result []*model.UserUsage
	for _, usage := range m.dailyUsage {
		if usage.UserID == userID &&
			(usage.Date.Equal(startDate) || usage.Date.After(startDate)) &&
			(usage.Date.Equal(endDate) || usage.Date.Before(endDate)) {
			result = append(result, usage)
		}
	}
	return result, nil
}

// BaseRepository methods for mockUserUsageRepository
func (m *mockUserUsageRepository) FindByID(ctx context.Context, id string) (*model.UserUsage, error) {
	// Not implemented - usage doesn't have ID field in this mock
	for _, usage := range m.dailyUsage {
		if usage.UserID == id {
			return usage, nil
		}
	}
	return nil, nil
}

func (m *mockUserUsageRepository) FindAll(ctx context.Context, limit, offset int) ([]*model.UserUsage, error) {
	var result []*model.UserUsage
	for _, usage := range m.dailyUsage {
		result = append(result, usage)
	}
	return result, nil
}

func (m *mockUserUsageRepository) Create(ctx context.Context, entity *model.UserUsage) error {
	key := entity.UserID + "_" + entity.Date.Format("2006-01-02")
	m.dailyUsage[key] = entity
	return nil
}

func (m *mockUserUsageRepository) Delete(ctx context.Context, id string) error {
	// Delete by userID_date format
	for key := range m.dailyUsage {
		m.dailyUsage[key] = nil
		delete(m.dailyUsage, key)
	}
	return nil
}

func (m *mockUserUsageRepository) Count(ctx context.Context) (int64, error) {
	return int64(len(m.dailyUsage)), nil
}

func (m *mockUserUsageRepository) Exists(ctx context.Context, id string) (bool, error) {
	// Check if any usage exists for this user
	for _, usage := range m.dailyUsage {
		if usage.UserID == id {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockUserUsageRepository) GetDB() *gorm.DB {
	return nil
}

func (m *mockUserUsageRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.UserUsage, error) {
	var result []*model.UserUsage
	count := 0
	for _, usage := range m.dailyUsage {
		if usage.UserID == userID {
			if offset > 0 {
				offset--
				continue
			}
			result = append(result, usage)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *mockUserUsageRepository) ResetDailyUsage(ctx context.Context, date time.Time) error {
	for key, usage := range m.dailyUsage {
		if usage.Date.Format("2006-01-02") == date.Format("2006-01-02") {
			usage.RequestCount = 0
			usage.TokenCount = 0
			usage.TotalCost = 0
			usage.ModelUsage = model.JSONB{}
			usage.IsExceeded = false
			m.dailyUsage[key] = usage
		}
	}
	return nil
}

func (m *mockUserUsageRepository) FindExceededUsage(ctx context.Context, date time.Time) ([]*model.UserUsage, error) {
	var result []*model.UserUsage
	for _, usage := range m.dailyUsage {
		if usage.IsExceeded && usage.Date.Format("2006-01-02") == date.Format("2006-01-02") {
			result = append(result, usage)
		}
	}
	return result, nil
}

type mockMonthlyUsageRepository struct {
	monthlyUsage map[string]*model.MonthlyUsage // key: "userID_yearMonth"
}

func (m *mockMonthlyUsageRepository) FindByUserAndMonth(ctx context.Context, userID, yearMonth string) (*model.MonthlyUsage, error) {
	key := userID + "_" + yearMonth
	usage, exists := m.monthlyUsage[key]
	if !exists {
		usage = &model.MonthlyUsage{
			UserID:       userID,
			YearMonth:    yearMonth,
			RequestCount: 0,
			TokenCount:   0,
			TotalCost:    0,
			ModelUsage:   model.JSONB{},
			IsExceeded:   false,
		}
		m.monthlyUsage[key] = usage
	}
	return usage, nil
}

func (m *mockMonthlyUsageRepository) IncrementUsage(ctx context.Context, userID, yearMonth string, requests, tokens int, cost float64, modelUsage model.JSONB) error {
	key := userID + "_" + yearMonth
	usage, exists := m.monthlyUsage[key]
	if !exists {
		usage = &model.MonthlyUsage{
			UserID:       userID,
			YearMonth:    yearMonth,
			RequestCount: 0,
			TokenCount:   0,
			TotalCost:    0,
			ModelUsage:   model.JSONB{},
			IsExceeded:   false,
		}
		m.monthlyUsage[key] = usage
	}

	usage.RequestCount += requests
	usage.TokenCount += tokens
	usage.TotalCost += cost

	// Merge model usage - accumulate values instead of overwriting
	if usage.ModelUsage == nil {
		usage.ModelUsage = model.JSONB{}
	}

	for modelID, newData := range modelUsage {
		if newDataMap, ok := newData.(model.JSONB); ok {
			if existingData, exists := usage.ModelUsage[modelID]; exists {
				if existingMap, ok := existingData.(model.JSONB); ok {
					// Merge existing and new data
					merged := model.JSONB{}

					// Copy existing values
					for k, v := range existingMap {
						merged[k] = v
					}

					// Add new values (accumulate where appropriate)
					for k, newVal := range newDataMap {
						switch k {
						case "requests", "tokens":
							// Accumulate integers
							existingVal := 0
							if v, ok := merged[k]; ok {
								switch val := v.(type) {
								case int:
									existingVal = val
								case float64:
									existingVal = int(val)
								}
							}

							var newValInt int
							switch val := newVal.(type) {
							case int:
								newValInt = val
							case float64:
								newValInt = int(val)
							default:
								newValInt = 0
							}

							merged[k] = existingVal + newValInt
						case "cost":
							// Accumulate floats
							existingVal := 0.0
							if v, ok := merged[k]; ok {
								switch val := v.(type) {
								case float64:
									existingVal = val
								case int:
									existingVal = float64(val)
								}
							}

							var newValFloat float64
							switch val := newVal.(type) {
							case float64:
								newValFloat = val
							case int:
								newValFloat = float64(val)
							default:
								newValFloat = 0.0
							}

							merged[k] = existingVal + newValFloat
						default:
							merged[k] = newVal
						}
					}

					usage.ModelUsage[modelID] = merged
				}
			} else {
				// No existing data for this model, just set it
				usage.ModelUsage[modelID] = newDataMap
			}
		}
	}

	return nil
}

func (m *mockMonthlyUsageRepository) Update(ctx context.Context, usage *model.MonthlyUsage) error {
	key := usage.UserID + "_" + usage.YearMonth
	m.monthlyUsage[key] = usage
	return nil
}

// BaseRepository methods for mockMonthlyUsageRepository
func (m *mockMonthlyUsageRepository) FindByID(ctx context.Context, id string) (*model.MonthlyUsage, error) {
	// Not implemented - monthly usage doesn't have ID field in this mock
	for _, usage := range m.monthlyUsage {
		if usage.UserID == id {
			return usage, nil
		}
	}
	return nil, nil
}

func (m *mockMonthlyUsageRepository) FindAll(ctx context.Context, limit, offset int) ([]*model.MonthlyUsage, error) {
	var result []*model.MonthlyUsage
	for _, usage := range m.monthlyUsage {
		result = append(result, usage)
	}
	return result, nil
}

func (m *mockMonthlyUsageRepository) Create(ctx context.Context, entity *model.MonthlyUsage) error {
	key := entity.UserID + "_" + entity.YearMonth
	m.monthlyUsage[key] = entity
	return nil
}

func (m *mockMonthlyUsageRepository) Delete(ctx context.Context, id string) error {
	// Delete by userID_yearMonth format
	for key := range m.monthlyUsage {
		m.monthlyUsage[key] = nil
		delete(m.monthlyUsage, key)
	}
	return nil
}

func (m *mockMonthlyUsageRepository) Count(ctx context.Context) (int64, error) {
	return int64(len(m.monthlyUsage)), nil
}

func (m *mockMonthlyUsageRepository) Exists(ctx context.Context, id string) (bool, error) {
	// Check if any monthly usage exists for this user
	for _, usage := range m.monthlyUsage {
		if usage.UserID == id {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockMonthlyUsageRepository) GetDB() *gorm.DB {
	return nil
}

func (m *mockMonthlyUsageRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.MonthlyUsage, error) {
	var result []*model.MonthlyUsage
	count := 0
	for _, usage := range m.monthlyUsage {
		if usage.UserID == userID {
			if offset > 0 {
				offset--
				continue
			}
			result = append(result, usage)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *mockMonthlyUsageRepository) GetMonthlySummary(ctx context.Context, yearMonth string) ([]*model.MonthlyUsage, error) {
	var result []*model.MonthlyUsage
	for _, usage := range m.monthlyUsage {
		if usage.YearMonth == yearMonth {
			result = append(result, usage)
		}
	}
	return result, nil
}

func (m *mockMonthlyUsageRepository) ResetMonthlyUsage(ctx context.Context, yearMonth string) error {
	for key, usage := range m.monthlyUsage {
		if usage.YearMonth == yearMonth {
			usage.RequestCount = 0
			usage.TokenCount = 0
			usage.TotalCost = 0
			usage.ModelUsage = model.JSONB{}
			usage.IsExceeded = false
			m.monthlyUsage[key] = usage
		}
	}
	return nil
}

func TestQuotaService_CheckQuota(t *testing.T) {
	ctx := context.Background()
	userID := "test-user-123"

	// Setup mock repositories
	quotaRepo := &mockUserQuotaRepository{
		quotas: make(map[string]*model.UserQuota),
	}
	usageRepo := &mockUserUsageRepository{
		dailyUsage: make(map[string]*model.UserUsage),
	}
	monthlyRepo := &mockMonthlyUsageRepository{
		monthlyUsage: make(map[string]*model.MonthlyUsage),
	}

	service := &quotaService{
		quotaRepo:   quotaRepo,
		usageRepo:   usageRepo,
		monthlyRepo: monthlyRepo,
	}

	// Test 1: Check quota for new user (should pass)
	t.Run("NewUserQuotaCheckPasses", func(t *testing.T) {
		result, err := service.CheckQuota(ctx, userID, 100, 0.5)
		if err != nil {
			t.Fatalf("CheckQuota failed: %v", err)
		}

		if !result.Allowed {
			t.Errorf("Expected quota check to pass for new user, but it failed: %s", result.Reason)
		}

		if result.DailyRequests != 0 {
			t.Errorf("Expected daily requests to be 0, got %d", result.DailyRequests)
		}

		if result.DailyRequestsLimit != 100 {
			t.Errorf("Expected daily request limit to be 100, got %d", result.DailyRequestsLimit)
		}
	})

	// Test 2: Check quota with low limits (should fail)
	t.Run("DailyRequestLimitExceeded", func(t *testing.T) {
		// Set up user with very low limits
		quotaRepo.quotas[userID] = &model.UserQuota{
			UserID:              userID,
			DailyRequestLimit:   2, // Very low
			DailyTokenLimit:     10000,
			DailyCostLimit:      10.0,
			MonthlyRequestLimit: 1000,
			MonthlyTokenLimit:   100000,
			MonthlyCostLimit:    100.0,
			PerMinuteRateLimit:  10,
			PerHourRateLimit:    100,
			ResetDay:            1,
			Timezone:            "UTC",
			IsActive:            true,
		}

		// First request should pass
		result1, err := service.CheckQuota(ctx, userID, 100, 0.5)
		if err != nil {
			t.Fatalf("First CheckQuota failed: %v", err)
		}
		if !result1.Allowed {
			t.Errorf("First quota check should pass, but failed: %s", result1.Reason)
		}

		// Record the usage (simulate a successful API call)
		if err := service.RecordUsage(ctx, userID, "test-model", 100, 0.5); err != nil {
			t.Fatalf("RecordUsage failed: %v", err)
		}

		// Second request should also pass (2nd request, limit is 2)
		result2, err := service.CheckQuota(ctx, userID, 100, 0.5)
		if err != nil {
			t.Fatalf("Second CheckQuota failed: %v", err)
		}
		if !result2.Allowed {
			t.Errorf("Second quota check should pass, but failed: %s", result2.Reason)
		}

		// Record second usage
		if err := service.RecordUsage(ctx, userID, "test-model", 100, 0.5); err != nil {
			t.Fatalf("RecordUsage failed: %v", err)
		}

		// Third request should fail (exceeds daily limit of 2)
		result3, err := service.CheckQuota(ctx, userID, 100, 0.5)
		if err != nil {
			t.Fatalf("Third CheckQuota failed: %v", err)
		}
		if result3.Allowed {
			t.Error("Third quota check should fail (daily limit exceeded), but it passed")
		}
		if result3.Reason != "daily request limit exceeded" {
			t.Errorf("Expected reason 'daily request limit exceeded', got '%s'", result3.Reason)
		}
	})

	// Test 3: Check quota with inactive quota
	t.Run("InactiveQuota", func(t *testing.T) {
		quotaRepo.quotas[userID] = &model.UserQuota{
			UserID:              userID,
			DailyRequestLimit:   100,
			DailyTokenLimit:     10000,
			DailyCostLimit:      10.0,
			MonthlyRequestLimit: 1000,
			MonthlyTokenLimit:   100000,
			MonthlyCostLimit:    100.0,
			PerMinuteRateLimit:  10,
			PerHourRateLimit:    100,
			ResetDay:            1,
			Timezone:            "UTC",
			IsActive:            false, // Inactive quota
		}

		result, err := service.CheckQuota(ctx, userID, 100, 0.5)
		if err != nil {
			t.Fatalf("CheckQuota failed: %v", err)
		}

		if result.Allowed {
			t.Error("Quota check should fail for inactive quota, but it passed")
		}
		if result.Reason != "quota is not active" {
			t.Errorf("Expected reason 'quota is not active', got '%s'", result.Reason)
		}
	})

	// Test 4: Test daily token limit
	t.Run("DailyTokenLimitExceeded", func(t *testing.T) {
		quotaRepo.quotas[userID] = &model.UserQuota{
			UserID:              userID,
			DailyRequestLimit:   100,
			DailyTokenLimit:     150, // Low token limit
			DailyCostLimit:      10.0,
			MonthlyRequestLimit: 1000,
			MonthlyTokenLimit:   100000,
			MonthlyCostLimit:    100.0,
			PerMinuteRateLimit:  10,
			PerHourRateLimit:    100,
			ResetDay:            1,
			Timezone:            "UTC",
			IsActive:            true,
		}

		// Clear usage for this test
		usageRepo.dailyUsage = make(map[string]*model.UserUsage)
		monthlyRepo.monthlyUsage = make(map[string]*model.MonthlyUsage)

		// First request with 100 tokens should pass
		result1, err := service.CheckQuota(ctx, userID, 100, 0.5)
		if err != nil {
			t.Fatalf("First CheckQuota failed: %v", err)
		}
		if !result1.Allowed {
			t.Errorf("First quota check should pass, but failed: %s", result1.Reason)
		}

		// Record usage
		if err := service.RecordUsage(ctx, userID, "test-model", 100, 0.5); err != nil {
			t.Fatalf("RecordUsage failed: %v", err)
		}

		// Second request with 100 tokens should fail (200 > 150 limit)
		result2, err := service.CheckQuota(ctx, userID, 100, 0.5)
		if err != nil {
			t.Fatalf("Second CheckQuota failed: %v", err)
		}
		if result2.Allowed {
			t.Error("Second quota check should fail (token limit exceeded), but it passed")
		}
		if result2.Reason != "daily token limit exceeded" {
			t.Errorf("Expected reason 'daily token limit exceeded', got '%s'", result2.Reason)
		}
	})
}

func TestQuotaService_RecordUsage(t *testing.T) {
	ctx := context.Background()
	userID := "test-user-456"
	modelID := "gpt-4o"

	// Setup mock repositories
	quotaRepo := &mockUserQuotaRepository{
		quotas: make(map[string]*model.UserQuota),
	}
	usageRepo := &mockUserUsageRepository{
		dailyUsage: make(map[string]*model.UserUsage),
	}
	monthlyRepo := &mockMonthlyUsageRepository{
		monthlyUsage: make(map[string]*model.MonthlyUsage),
	}

	service := &quotaService{
		quotaRepo:   quotaRepo,
		usageRepo:   usageRepo,
		monthlyRepo: monthlyRepo,
	}

	// Test: Record usage increments counters correctly
	t.Run("RecordUsageIncrementsCounters", func(t *testing.T) {
		// Clear any existing usage
		usageRepo.dailyUsage = make(map[string]*model.UserUsage)
		monthlyRepo.monthlyUsage = make(map[string]*model.MonthlyUsage)

		// Record first usage
		if err := service.RecordUsage(ctx, userID, modelID, 100, 0.75); err != nil {
			t.Fatalf("First RecordUsage failed: %v", err)
		}

		// Check daily usage
		now := time.Now()
		dailyUsage, err := service.GetDailyUsage(ctx, userID, now)
		if err != nil {
			t.Fatalf("GetDailyUsage failed: %v", err)
		}

		if dailyUsage.RequestCount != 1 {
			t.Errorf("Expected daily request count to be 1, got %d", dailyUsage.RequestCount)
		}
		if dailyUsage.TokenCount != 100 {
			t.Errorf("Expected daily token count to be 100, got %d", dailyUsage.TokenCount)
		}
		if dailyUsage.TotalCost != 0.75 {
			t.Errorf("Expected daily total cost to be 0.75, got %f", dailyUsage.TotalCost)
		}

		// Check model usage in daily record
		if dailyUsage.ModelUsage == nil {
			t.Error("Expected model usage to be recorded")
		} else if modelData, ok := dailyUsage.ModelUsage[modelID].(model.JSONB); ok {
			// Handle both int and float64 types for requests
			var requests int
			switch v := modelData["requests"].(type) {
			case int:
				requests = v
			case float64:
				requests = int(v)
			default:
				t.Errorf("Expected model requests to be int or float64, got %T", modelData["requests"])
				requests = 0
			}
			if requests != 1 {
				t.Errorf("Expected model requests to be 1, got %v", modelData["requests"])
			}

			// Handle both int and float64 types for tokens
			var tokens int
			switch v := modelData["tokens"].(type) {
			case int:
				tokens = v
			case float64:
				tokens = int(v)
			default:
				t.Errorf("Expected model tokens to be int or float64, got %T", modelData["tokens"])
				tokens = 0
			}
			if tokens != 100 {
				t.Errorf("Expected model tokens to be 100, got %v", modelData["tokens"])
			}

			// Handle both float64 and int types for cost
			var cost float64
			switch v := modelData["cost"].(type) {
			case float64:
				cost = v
			case int:
				cost = float64(v)
			default:
				t.Errorf("Expected model cost to be float64 or int, got %T", modelData["cost"])
				cost = 0
			}
			if cost != 0.75 {
				t.Errorf("Expected model cost to be 0.75, got %v", modelData["cost"])
			}
		} else {
			t.Error("Model usage data not found in expected format")
		}

		// Check monthly usage
		yearMonth := now.Format("2006-01")
		monthlyUsage, err := service.GetMonthlyUsage(ctx, userID, yearMonth)
		if err != nil {
			t.Fatalf("GetMonthlyUsage failed: %v", err)
		}

		if monthlyUsage.RequestCount != 1 {
			t.Errorf("Expected monthly request count to be 1, got %d", monthlyUsage.RequestCount)
		}
		if monthlyUsage.TokenCount != 100 {
			t.Errorf("Expected monthly token count to be 100, got %d", monthlyUsage.TokenCount)
		}
		if monthlyUsage.TotalCost != 0.75 {
			t.Errorf("Expected monthly total cost to be 0.75, got %f", monthlyUsage.TotalCost)
		}

		// Record second usage
		if err := service.RecordUsage(ctx, userID, modelID, 200, 1.5); err != nil {
			t.Fatalf("Second RecordUsage failed: %v", err)
		}

		// Verify cumulative counts
		dailyUsage2, _ := service.GetDailyUsage(ctx, userID, now)
		if dailyUsage2.RequestCount != 2 {
			t.Errorf("Expected daily request count to be 2 after second usage, got %d", dailyUsage2.RequestCount)
		}
		if dailyUsage2.TokenCount != 300 {
			t.Errorf("Expected daily token count to be 300 after second usage, got %d", dailyUsage2.TokenCount)
		}
		if dailyUsage2.TotalCost != 2.25 {
			t.Errorf("Expected daily total cost to be 2.25 after second usage, got %f", dailyUsage2.TotalCost)
		}

		// Verify model usage updated
		if modelData, ok := dailyUsage2.ModelUsage[modelID].(model.JSONB); ok {
			// Handle both int and float64 types for requests
			var requests int
			switch v := modelData["requests"].(type) {
			case int:
				requests = v
			case float64:
				requests = int(v)
			default:
				t.Errorf("Expected model requests to be int or float64, got %T", modelData["requests"])
				requests = 0
			}
			if requests != 2 {
				t.Errorf("Expected model requests to be 2 after second usage, got %v", modelData["requests"])
			}

			// Handle both int and float64 types for tokens
			var tokens int
			switch v := modelData["tokens"].(type) {
			case int:
				tokens = v
			case float64:
				tokens = int(v)
			default:
				t.Errorf("Expected model tokens to be int or float64, got %T", modelData["tokens"])
				tokens = 0
			}
			if tokens != 300 {
				t.Errorf("Expected model tokens to be 300 after second usage, got %v", modelData["tokens"])
			}

			// Handle both float64 and int types for cost
			var cost float64
			switch v := modelData["cost"].(type) {
			case float64:
				cost = v
			case int:
				cost = float64(v)
			default:
				t.Errorf("Expected model cost to be float64 or int, got %T", modelData["cost"])
				cost = 0
			}
			if cost != 2.25 {
				t.Errorf("Expected model cost to be 2.25 after second usage, got %v", modelData["cost"])
			}
		}
	})

	// Test: Record usage sets exceeded flag when limits are exceeded
	t.Run("RecordUsageSetsExceededFlag", func(t *testing.T) {
		// Set up user with very low limits
		quotaRepo.quotas[userID] = &model.UserQuota{
			UserID:              userID,
			DailyRequestLimit:   1, // Very low
			DailyTokenLimit:     100,
			DailyCostLimit:      1.0,
			MonthlyRequestLimit: 10,
			MonthlyTokenLimit:   1000,
			MonthlyCostLimit:    10.0,
			PerMinuteRateLimit:  10,
			PerHourRateLimit:    100,
			ResetDay:            1,
			Timezone:            "UTC",
			IsActive:            true,
		}

		// Clear usage
		usageRepo.dailyUsage = make(map[string]*model.UserUsage)
		monthlyRepo.monthlyUsage = make(map[string]*model.MonthlyUsage)

		// Record usage that exceeds daily request limit
		if err := service.RecordUsage(ctx, userID, modelID, 50, 0.5); err != nil {
			t.Fatalf("RecordUsage failed: %v", err)
		}

		// Record second usage that puts us over the limit
		if err := service.RecordUsage(ctx, userID, modelID, 50, 0.5); err != nil {
			t.Fatalf("Second RecordUsage failed: %v", err)
		}

		// Check that exceeded flag is set
		now := time.Now()
		dailyUsage, err := service.GetDailyUsage(ctx, userID, now)
		if err != nil {
			t.Fatalf("GetDailyUsage failed: %v", err)
		}

		if !dailyUsage.IsExceeded {
			t.Error("Expected IsExceeded to be true when daily request limit is exceeded")
		}
	})
}

func TestQuotaService_ResetQuotas(t *testing.T) {
	ctx := context.Background()
	userID := "test-user-789"

	// Setup mock repositories
	quotaRepo := &mockUserQuotaRepository{
		quotas: make(map[string]*model.UserQuota),
	}
	usageRepo := &mockUserUsageRepository{
		dailyUsage: make(map[string]*model.UserUsage),
	}
	monthlyRepo := &mockMonthlyUsageRepository{
		monthlyUsage: make(map[string]*model.MonthlyUsage),
	}

	service := &quotaService{
		quotaRepo:   quotaRepo,
		usageRepo:   usageRepo,
		monthlyRepo: monthlyRepo,
	}

	// Test: Reset daily quota
	t.Run("ResetDailyQuota", func(t *testing.T) {
		// Record some usage
		if err := service.RecordUsage(ctx, userID, "test-model", 100, 1.0); err != nil {
			t.Fatalf("RecordUsage failed: %v", err)
		}

		// Verify usage was recorded
		now := time.Now()
		dailyUsage, err := service.GetDailyUsage(ctx, userID, now)
		if err != nil {
			t.Fatalf("GetDailyUsage failed: %v", err)
		}
		if dailyUsage.RequestCount == 0 {
			t.Error("Expected usage to be recorded")
		}

		// Reset daily quota
		if err := service.ResetDailyQuota(ctx, userID); err != nil {
			t.Fatalf("ResetDailyQuota failed: %v", err)
		}

		// Verify usage was reset
		dailyUsageAfterReset, err := service.GetDailyUsage(ctx, userID, now)
		if err != nil {
			t.Fatalf("GetDailyUsage after reset failed: %v", err)
		}
		if dailyUsageAfterReset.RequestCount != 0 {
			t.Errorf("Expected request count to be 0 after reset, got %d", dailyUsageAfterReset.RequestCount)
		}
		if dailyUsageAfterReset.TokenCount != 0 {
			t.Errorf("Expected token count to be 0 after reset, got %d", dailyUsageAfterReset.TokenCount)
		}
		if dailyUsageAfterReset.TotalCost != 0 {
			t.Errorf("Expected total cost to be 0 after reset, got %f", dailyUsageAfterReset.TotalCost)
		}
		if dailyUsageAfterReset.IsExceeded {
			t.Error("Expected IsExceeded to be false after reset")
		}
	})

	// Test: Reset monthly quota
	t.Run("ResetMonthlyQuota", func(t *testing.T) {
		// Record some usage
		if err := service.RecordUsage(ctx, userID, "test-model", 500, 5.0); err != nil {
			t.Fatalf("RecordUsage failed: %v", err)
		}

		// Verify monthly usage was recorded
		now := time.Now()
		yearMonth := now.Format("2006-01")
		monthlyUsage, err := service.GetMonthlyUsage(ctx, userID, yearMonth)
		if err != nil {
			t.Fatalf("GetMonthlyUsage failed: %v", err)
		}
		if monthlyUsage.RequestCount == 0 {
			t.Error("Expected monthly usage to be recorded")
		}

		// Reset monthly quota
		if err := service.ResetMonthlyQuota(ctx, userID); err != nil {
			t.Fatalf("ResetMonthlyQuota failed: %v", err)
		}

		// Verify monthly usage was reset
		monthlyUsageAfterReset, err := service.GetMonthlyUsage(ctx, userID, yearMonth)
		if err != nil {
			t.Fatalf("GetMonthlyUsage after reset failed: %v", err)
		}
		if monthlyUsageAfterReset.RequestCount != 0 {
			t.Errorf("Expected monthly request count to be 0 after reset, got %d", monthlyUsageAfterReset.RequestCount)
		}
		if monthlyUsageAfterReset.TokenCount != 0 {
			t.Errorf("Expected monthly token count to be 0 after reset, got %d", monthlyUsageAfterReset.TokenCount)
		}
		if monthlyUsageAfterReset.TotalCost != 0 {
			t.Errorf("Expected monthly total cost to be 0 after reset, got %f", monthlyUsageAfterReset.TotalCost)
		}
		if monthlyUsageAfterReset.IsExceeded {
			t.Error("Expected monthly IsExceeded to be false after reset")
		}
	})
}
