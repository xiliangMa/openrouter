package model

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"massrouter.ai/backend/internal/service"
)

type Controller struct {
	modelService service.ModelService
	validator    *validator.Validate
}

func NewController(modelService service.ModelService) *Controller {
	return &Controller{
		modelService: modelService,
		validator:    validator.New(),
	}
}

// ListModels godoc
// @Summary List models
// @Description Get paginated list of available models with filtering
// @Tags model
// @Produce json
// @Param page query integer false "Page number" default(1) minimum(1)
// @Param limit query integer false "Items per page" default(20) minimum(1) maximum(100)
// @Param category query string false "Filter by category"
// @Param provider query string false "Filter by provider"
// @Param search query string false "Search query"
// @Param is_free query boolean false "Filter by free status"
// @Param sort_by query string false "Sort field (name, price, created_at)"
// @Param sort_order query string false "Sort order (asc, desc)"
// @Success 200 {object} map[string]interface{} "Models retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/models [get]
func (c *Controller) ListModels(ctx *gin.Context) {
	var req service.ListModelsRequest

	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	req.Page = page

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}
	req.Limit = limit

	req.Category = ctx.Query("category")
	req.Provider = ctx.Query("provider")
	req.Search = ctx.Query("search")
	req.SortBy = ctx.Query("sort_by")
	req.SortOrder = ctx.Query("sort_order")

	if isFreeStr := ctx.Query("is_free"); isFreeStr != "" {
		isFree, err := strconv.ParseBool(isFreeStr)
		if err == nil {
			req.IsFree = &isFree
		}
	}

	if err := c.validator.Struct(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_VALIDATION",
				"message": "Validation failed",
				"details": err.Error(),
			},
		})
		return
	}

	response, err := c.modelService.ListModels(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to list models",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetModelDetails godoc
// @Summary Get model details
// @Description Get detailed information about a specific model
// @Tags model
// @Produce json
// @Param id path string true "Model ID"
// @Success 200 {object} map[string]interface{} "Model details retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - missing ID"
// @Failure 404 {object} map[string]interface{} "Model not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/models/{id} [get]
func (c *Controller) GetModelDetails(ctx *gin.Context) {
	modelID := ctx.Param("id")
	if modelID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Model ID is required",
			},
		})
		return
	}

	details, err := c.modelService.GetModelDetails(ctx.Request.Context(), modelID)
	if err != nil {
		status := http.StatusInternalServerError
		errorCode := "ERR_INTERNAL"

		if err.Error() == "model not found" {
			status = http.StatusNotFound
			errorCode = "ERR_MODEL_NOT_FOUND"
		}

		ctx.JSON(status, gin.H{
			"success": false,
			"error": gin.H{
				"code":    errorCode,
				"message": err.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    details,
	})
}

// SearchModels godoc
// @Summary Search models
// @Description Search models with advanced filtering
// @Tags model
// @Produce json
// @Param q query string false "Search query"
// @Param categories query []string false "Filter by categories"
// @Param providers query []string false "Filter by providers"
// @Param min_price query number false "Minimum price per token"
// @Param max_price query number false "Maximum price per token"
// @Param is_free query boolean false "Filter by free status"
// @Success 200 {object} map[string]interface{} "Search results retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/models/search [get]
func (c *Controller) SearchModels(ctx *gin.Context) {
	query := ctx.Query("q")

	var filters service.ModelFilters

	if categories := ctx.QueryArray("categories"); len(categories) > 0 {
		filters.Categories = categories
	}

	if providers := ctx.QueryArray("providers"); len(providers) > 0 {
		filters.Providers = providers
	}

	if minPriceStr := ctx.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			filters.MinPrice = minPrice
		}
	}

	if maxPriceStr := ctx.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			filters.MaxPrice = maxPrice
		}
	}

	if isFreeStr := ctx.Query("is_free"); isFreeStr != "" {
		if isFree, err := strconv.ParseBool(isFreeStr); err == nil {
			filters.IsFree = &isFree
		}
	}

	response, err := c.modelService.SearchModels(ctx.Request.Context(), query, &filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to search models",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetModelProviders godoc
// @Summary Get model providers
// @Description Get list of all model providers
// @Tags model
// @Produce json
// @Success 200 {object} map[string]interface{} "Providers retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/models/providers [get]
func (c *Controller) GetModelProviders(ctx *gin.Context) {
	providers, err := c.modelService.GetModelProviders(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to get model providers",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"providers": providers,
		},
	})
}

// GetModelCategories godoc
// @Summary Get model categories
// @Description Get list of all model categories
// @Tags model
// @Produce json
// @Success 200 {object} map[string]interface{} "Categories retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/models/categories [get]
func (c *Controller) GetModelCategories(ctx *gin.Context) {
	categories, err := c.modelService.GetModelCategories(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to get model categories",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"categories": categories,
		},
	})
}
