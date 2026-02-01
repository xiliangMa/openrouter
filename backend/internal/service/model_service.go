package service

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"massrouter.ai/backend/internal/model"
	"massrouter.ai/backend/internal/repository"
)

var _ = (*gorm.DB)(nil)

type modelService struct {
	modelRepo     repository.ModelRepository
	providerRepo  repository.ModelProviderRepository
	statisticRepo repository.ModelStatisticRepository
}

func NewModelService(
	modelRepo repository.ModelRepository,
	providerRepo repository.ModelProviderRepository,
	statisticRepo repository.ModelStatisticRepository,
) ModelService {
	return &modelService{
		modelRepo:     modelRepo,
		providerRepo:  providerRepo,
		statisticRepo: statisticRepo,
	}
}

func (s *modelService) ListModels(ctx context.Context, req *ListModelsRequest) (*ListModelsResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := (req.Page - 1) * limit

	var models []*model.Model
	var total int64
	var err error

	if req.Category != "" || req.Provider != "" || req.Search != "" || req.IsFree != nil {
		models, total, err = s.searchModelsWithFilters(ctx, req, limit, offset)
	} else {
		models, total, err = s.getAllModels(ctx, limit, offset, req.SortBy, req.SortOrder)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	modelInfos := make([]*ModelInfo, len(models))
	for i, m := range models {
		modelInfos[i] = &ModelInfo{
			Model:        m,
			ProviderName: m.Provider.Name,
		}
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &ListModelsResponse{
		Models:     modelInfos,
		Total:      total,
		Page:       req.Page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *modelService) searchModelsWithFilters(ctx context.Context, req *ListModelsRequest, limit, offset int) ([]*model.Model, int64, error) {
	var models []*model.Model
	db := s.modelRepo.GetDB().WithContext(ctx).Model(&model.Model{}).Where("is_active = ?", true).Preload("Provider")

	if req.Category != "" {
		db = db.Where("category = ?", req.Category)
	}

	if req.Provider != "" {
		db = db.Joins("JOIN model_providers ON models.provider_id = model_providers.id").
			Where("model_providers.name = ?", req.Provider)
	}

	if req.Search != "" {
		searchPattern := "%" + strings.ToLower(req.Search) + "%"
		db = db.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern)
	}

	if req.IsFree != nil {
		db = db.Where("is_free = ?", *req.IsFree)
	}

	if req.SortBy != "" {
		order := s.getSortOrder(req.SortBy, req.SortOrder)
		db = db.Order(order)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count models: %w", err)
	}

	if err := db.Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find models: %w", err)
	}

	return models, total, nil
}

func (s *modelService) getAllModels(ctx context.Context, limit, offset int, sortBy, sortOrder string) ([]*model.Model, int64, error) {
	db := s.modelRepo.GetDB().WithContext(ctx).Model(&model.Model{}).Where("is_active = ?", true).Preload("Provider")

	if sortBy != "" {
		order := s.getSortOrder(sortBy, sortOrder)
		db = db.Order(order)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count models: %w", err)
	}

	var models []*model.Model
	if err := db.Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find models: %w", err)
	}

	return models, total, nil
}

func (s *modelService) getSortOrder(sortBy, sortOrder string) string {
	order := "created_at DESC"
	if sortOrder == "" {
		sortOrder = "desc"
	}

	switch sortBy {
	case "name":
		order = fmt.Sprintf("name %s", strings.ToUpper(sortOrder))
	case "price":
		order = fmt.Sprintf("input_price %s", strings.ToUpper(sortOrder))
	case "created_at":
		order = fmt.Sprintf("created_at %s", strings.ToUpper(sortOrder))
	}

	return order
}

func (s *modelService) GetModelDetails(ctx context.Context, modelID string) (*ModelDetails, error) {
	var modelObj model.Model
	// Load model with provider relationship preloaded
	err := s.modelRepo.GetDB().WithContext(ctx).
		Preload("Provider").
		First(&modelObj, "id = ?", modelID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("model not found")
		}
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	var dailyRequests int
	var successRate, avgLatency float64

	statistics, err := s.statisticRepo.FindByModelID(ctx, modelID, 30, 0)
	if err == nil && len(statistics) > 0 {
		var totalRequests, totalLatency float64
		for _, stat := range statistics {
			totalRequests += float64(stat.TotalRequests)
			totalLatency += stat.AvgResponseTime * float64(stat.TotalRequests)
		}
		if totalRequests > 0 {
			dailyRequests = int(totalRequests / float64(len(statistics)))
			avgLatency = totalLatency / totalRequests

			var totalSuccess float64
			for _, stat := range statistics {
				totalSuccess += stat.SuccessRate * float64(stat.TotalRequests)
			}
			successRate = totalSuccess / totalRequests
		}
	}

	return &ModelDetails{
		Model:         &modelObj,
		Provider:      &modelObj.Provider,
		DailyRequests: dailyRequests,
		SuccessRate:   successRate,
		AvgLatency:    avgLatency,
	}, nil
}

func (s *modelService) SearchModels(ctx context.Context, query string, filters *ModelFilters) (*ListModelsResponse, error) {
	if filters == nil {
		filters = &ModelFilters{}
	}

	var models []*model.Model
	dbQuery := s.modelRepo.GetDB().WithContext(ctx).Model(&model.Model{}).Where("is_active = ?", true).Preload("Provider")

	if query != "" {
		searchPattern := "%" + strings.ToLower(query) + "%"
		dbQuery = dbQuery.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern)
	}

	if len(filters.Categories) > 0 {
		dbQuery = dbQuery.Where("category IN (?)", filters.Categories)
	}

	if len(filters.Providers) > 0 {
		dbQuery = dbQuery.Joins("JOIN model_providers ON models.provider_id = model_providers.id").
			Where("model_providers.name IN (?)", filters.Providers)
	}

	if filters.MinPrice > 0 {
		dbQuery = dbQuery.Where("input_price >= ?", filters.MinPrice)
	}

	if filters.MaxPrice > 0 {
		dbQuery = dbQuery.Where("input_price <= ?", filters.MaxPrice)
	}

	if filters.IsFree != nil {
		dbQuery = dbQuery.Where("is_free = ?", *filters.IsFree)
	}

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count models: %w", err)
	}

	if err := dbQuery.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to search models: %w", err)
	}

	modelInfos := make([]*ModelInfo, len(models))
	for i, m := range models {
		modelInfos[i] = &ModelInfo{
			Model:        m,
			ProviderName: m.Provider.Name,
		}
	}

	return &ListModelsResponse{
		Models:     modelInfos,
		Total:      total,
		Page:       1,
		Limit:      int(total),
		TotalPages: 1,
	}, nil
}

func (s *modelService) GetModelProviders(ctx context.Context) ([]*model.ModelProvider, error) {
	providers, err := s.providerRepo.FindActiveProviders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}
	return providers, nil
}

func (s *modelService) GetModelCategories(ctx context.Context) ([]string, error) {
	var categories []string
	err := s.modelRepo.GetDB().WithContext(ctx).Raw("SELECT DISTINCT category FROM models WHERE category IS NOT NULL AND category != '' AND is_active = true ORDER BY category").Scan(&categories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	return categories, nil
}
