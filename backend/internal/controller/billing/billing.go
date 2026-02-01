package billing

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"massrouter.ai/backend/internal/service"
)

type Controller struct {
	billingService service.BillingService
	validator      *validator.Validate
}

func NewController(billingService service.BillingService) *Controller {
	return &Controller{
		billingService: billingService,
		validator:      validator.New(),
	}
}

// GetBalance godoc
// @Summary Get billing balance
// @Description Get current user's billing balance information
// @Tags billing
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Balance retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/billing/balance [get]
func (c *Controller) GetBalance(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}

	balance, err := c.billingService.GetBalance(ctx.Request.Context(), userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to get balance",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    balance,
	})
}

// GetPaymentHistory godoc
// @Summary Get payment history
// @Description Get paginated payment history for current user
// @Tags billing
// @Produce json
// @Security BearerAuth
// @Param page query integer false "Page number" default(1) minimum(1)
// @Param limit query integer false "Items per page" default(20) minimum(1) maximum(100)
// @Success 200 {object} map[string]interface{} "Payment history retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/billing/payments [get]
func (c *Controller) GetPaymentHistory(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}

	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	response, err := c.billingService.GetPaymentHistory(ctx.Request.Context(), userID.(string), page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to get payment history",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// CreatePayment godoc
// @Summary Create payment
// @Description Create a new payment for current user
// @Tags billing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreatePaymentRequest true "Payment creation request"
// @Success 201 {object} map[string]interface{} "Payment created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/billing/payments [post]
func (c *Controller) CreatePayment(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}

	var req service.CreatePaymentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
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

	paymentInfo, err := c.billingService.CreatePayment(ctx.Request.Context(), userID.(string), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to create payment",
			},
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    paymentInfo,
	})
}

// ProcessPaymentWebhook godoc
// @Summary Process payment webhook
// @Description Process payment webhook from payment provider (requires signature verification)
// @Tags billing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Signature header string true "Webhook signature"
// @Param payload body string true "Webhook payload"
// @Success 200 {object} map[string]interface{} "Webhook processed successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - missing signature or invalid payload"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/billing/webhook [post]
func (c *Controller) ProcessPaymentWebhook(ctx *gin.Context) {
	signature := ctx.GetHeader("X-Signature")
	if signature == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Signature header is required",
			},
		})
		return
	}

	payload, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Failed to read request body",
			},
		})
		return
	}

	if err := c.billingService.ProcessPaymentWebhook(ctx.Request.Context(), payload, signature); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to process webhook",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Webhook processed successfully",
		},
	})
}

// GetBillingRecords godoc
// @Summary Get billing records
// @Description Get paginated billing records (usage charges) for current user
// @Tags billing
// @Produce json
// @Security BearerAuth
// @Param page query integer false "Page number" default(1) minimum(1)
// @Param limit query integer false "Items per page" default(20) minimum(1) maximum(100)
// @Success 200 {object} map[string]interface{} "Billing records retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/billing/records [get]
func (c *Controller) GetBillingRecords(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}

	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	response, err := c.billingService.GetBillingRecords(ctx.Request.Context(), userID.(string), page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to get billing records",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// CalculateCost godoc
// @Summary Calculate cost
// @Description Calculate cost for using a model with given token counts
// @Tags billing
// @Produce json
// @Security BearerAuth
// @Param model_id query string true "Model ID"
// @Param input_tokens query integer false "Input tokens" default(0) minimum(0)
// @Param output_tokens query integer false "Output tokens" default(0) minimum(0)
// @Success 200 {object} map[string]interface{} "Cost calculated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - missing or invalid parameters"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Model not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/billing/calculate-cost [post]
func (c *Controller) CalculateCost(ctx *gin.Context) {
	modelID := ctx.Query("model_id")
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

	inputTokens, err := strconv.Atoi(ctx.DefaultQuery("input_tokens", "0"))
	if err != nil || inputTokens < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Invalid input tokens",
			},
		})
		return
	}

	outputTokens, err := strconv.Atoi(ctx.DefaultQuery("output_tokens", "0"))
	if err != nil || outputTokens < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Invalid output tokens",
			},
		})
		return
	}

	calculation, err := c.billingService.CalculateCost(ctx.Request.Context(), modelID, inputTokens, outputTokens)
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
		"data":    calculation,
	})
}
