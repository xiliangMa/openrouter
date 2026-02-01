package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"massrouter.ai/backend/internal/service"
)

type Controller struct {
	adminService service.AdminService
	validator    *validator.Validate
}

func NewController(adminService service.AdminService) *Controller {
	return &Controller{
		adminService: adminService,
		validator:    validator.New(),
	}
}

// ListUsers godoc
// @Summary List users (admin)
// @Description Get paginated list of users with filtering (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param page query integer false "Page number" default(1) minimum(1)
// @Param limit query integer false "Items per page" default(20) minimum(1) maximum(100)
// @Param search query string false "Search query"
// @Param role query string false "Filter by role (user, admin)"
// @Param status query string false "Filter by status (active, suspended, deleted)"
// @Param sort_by query string false "Sort field"
// @Param sort_order query string false "Sort order (asc, desc)"
// @Success 200 {object} map[string]interface{} "Users retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid parameters"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/admin/users [get]
func (c *Controller) ListUsers(ctx *gin.Context) {
	var req service.ListUsersRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Invalid request parameters",
			},
		})
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	users, err := c.adminService.ListUsers(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to list users",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
	})
}

// GetUserDetails godoc
// @Summary Get user details (admin)
// @Description Get detailed user information (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{} "User details retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - missing ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/admin/users/{id} [get]
func (c *Controller) GetUserDetails(ctx *gin.Context) {
	userID := ctx.Param("id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "User ID is required",
			},
		})
		return
	}

	userDetails, err := c.adminService.GetUserDetails(ctx.Request.Context(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "ERR_NOT_FOUND",
					"message": "User not found",
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to get user details",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    userDetails,
	})
}

// UpdateUser godoc
// @Summary Update user (admin)
// @Description Update user information (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body service.AdminUpdateUserRequest true "User update request"
// @Success 200 {object} map[string]interface{} "User updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/admin/users/{id} [put]
func (c *Controller) UpdateUser(ctx *gin.Context) {
	userID := ctx.Param("id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "User ID is required",
			},
		})
		return
	}

	var req service.AdminUpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
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

	if err := c.adminService.UpdateUser(ctx.Request.Context(), userID, &req); err != nil {
		if err.Error() == "user not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "ERR_NOT_FOUND",
					"message": "User not found",
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to update user",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "User updated successfully",
		},
	})
}

// CreateModelProvider godoc
// @Summary Create model provider (admin)
// @Description Create a new model provider (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateModelProviderRequest true "Model provider creation request"
// @Success 201 {object} map[string]interface{} "Model provider created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/admin/providers [post]
func (c *Controller) CreateModelProvider(ctx *gin.Context) {
	var req service.CreateModelProviderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
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

	provider, err := c.adminService.CreateModelProvider(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to create model provider",
			},
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    provider,
	})
}

// UpdateModelProvider godoc
// @Summary Update model provider (admin)
// @Description Update model provider information (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Provider ID"
// @Param request body service.UpdateModelProviderRequest true "Model provider update request"
// @Success 200 {object} map[string]interface{} "Model provider updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure 404 {object} map[string]interface{} "Model provider not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/admin/providers/{id} [put]
func (c *Controller) UpdateModelProvider(ctx *gin.Context) {
	providerID := ctx.Param("id")
	if providerID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Provider ID is required",
			},
		})
		return
	}

	var req service.UpdateModelProviderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
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

	if err := c.adminService.UpdateModelProvider(ctx.Request.Context(), providerID, &req); err != nil {
		if err.Error() == "provider not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "ERR_NOT_FOUND",
					"message": "Model provider not found",
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to update model provider",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Model provider updated successfully",
		},
	})
}

// CreateModel godoc
// @Summary Create model (admin)
// @Description Create a new model (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateModelRequest true "Model creation request"
// @Success 201 {object} map[string]interface{} "Model created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/admin/models [post]
func (c *Controller) CreateModel(ctx *gin.Context) {
	var req service.CreateModelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
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

	model, err := c.adminService.CreateModel(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to create model",
			},
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    model,
	})
}

// UpdateModel godoc
// @Summary Update model (admin)
// @Description Update model information (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Model ID"
// @Param request body service.UpdateModelRequest true "Model update request"
// @Success 200 {object} map[string]interface{} "Model updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure 404 {object} map[string]interface{} "Model not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/admin/models/{id} [put]
func (c *Controller) UpdateModel(ctx *gin.Context) {
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

	var req service.UpdateModelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
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

	if err := c.adminService.UpdateModel(ctx.Request.Context(), modelID, &req); err != nil {
		if err.Error() == "model not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "ERR_NOT_FOUND",
					"message": "Model not found",
				},
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to update model",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Model updated successfully",
		},
	})
}

// GetSystemStats godoc
// @Summary Get system statistics (admin)
// @Description Get system statistics and metrics (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "System statistics retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/admin/stats [get]
func (c *Controller) GetSystemStats(ctx *gin.Context) {
	stats, err := c.adminService.GetSystemStats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to get system statistics",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// UpdateSystemConfig godoc
// @Summary Update system config (admin)
// @Description Update system configuration value (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "Config key"
// @Param request body map[string]string true "Config value"
// @Success 200 {object} map[string]interface{} "System config updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - admin access required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/admin/config/{key} [put]
func (c *Controller) UpdateSystemConfig(ctx *gin.Context) {
	key := ctx.Param("key")
	if key == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Config key is required",
			},
		})
		return
	}

	var req struct {
		Value string `json:"value" validate:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
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

	if err := c.adminService.UpdateSystemConfig(ctx.Request.Context(), key, req.Value); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to update system config",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "System config updated successfully",
		},
	})
}
