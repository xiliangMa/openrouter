package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"massrouter.ai/backend/internal/model"
	"massrouter.ai/backend/internal/service"
)

type Controller struct {
	modelService   service.ModelService
	billingService service.BillingService
	quotaService   service.QuotaService
	validator      *validator.Validate
}

func NewController(
	modelService service.ModelService,
	billingService service.BillingService,
	quotaService service.QuotaService,
) *Controller {
	return &Controller{
		modelService:   modelService,
		billingService: billingService,
		quotaService:   quotaService,
		validator:      validator.New(),
	}
}

// ChatCompletionRequest represents the OpenAI-compatible chat completion request
type ChatCompletionRequest struct {
	Model       string                  `json:"model" binding:"required"`
	Messages    []ChatCompletionMessage `json:"messages" binding:"required"`
	MaxTokens   *int                    `json:"max_tokens,omitempty"`
	Temperature *float64                `json:"temperature,omitempty"`
	TopP        *float64                `json:"top_p,omitempty"`
	Stream      bool                    `json:"stream,omitempty"`
}

type ChatCompletionMessage struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// ChatCompletion handles chat completion requests
func (c *Controller) ChatCompletion(ctx *gin.Context) {
	// Get user from context (set by auth middleware)
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_401",
				"message": "Authentication required",
			},
		})
		return
	}

	// Parse request
	var req ChatCompletionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_400",
				"message": "Invalid request format",
				"details": err.Error(),
			},
		})
		return
	}

	// Validate request
	if err := c.validator.Struct(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_400",
				"message": "Validation failed",
				"details": err.Error(),
			},
		})
		return
	}

	// Marshal request back to JSON for forwarding
	requestBody, err := json.Marshal(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_500",
				"message": "Failed to marshal request",
			},
		})
		return
	}

	// Get API key from header
	apiKey := ctx.GetHeader("X-API-Key")
	if apiKey == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_400",
				"message": "API key required",
				"details": "Provide API key in X-API-Key header",
			},
		})
		return
	}

	// Verify API key belongs to user and is active
	// This would be done through a service call
	// For now, we'll just log it
	fmt.Printf("User %s using API key: %s\n", userID, apiKey[:min(8, len(apiKey))])

	// Get model details by name (validate model exists and get provider info)
	modelObj, provider, err := c.findModelByName(ctx, req.Model)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_404",
				"message": "Model not found",
				"details": err.Error(),
			},
		})
		return
	}

	if provider == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_500",
				"message": "Model provider not found",
			},
		})
		return
	}

	// Check if provider has API key configured
	// Development mode simulation for empty or test API keys
	if gin.Mode() == gin.DebugMode && (provider.APIKey == "" || strings.HasPrefix(provider.APIKey, "sk-test-")) {
		// Return simulated response for development
		simulatedResponse := gin.H{
			"id":      "chatcmpl-simulated-123",
			"object":  "chat.completion",
			"created": time.Now().Unix(),
			"model":   req.Model,
			"choices": []gin.H{
				{
					"index": 0,
					"message": gin.H{
						"role":    "assistant",
						"content": fmt.Sprintf("This is a simulated response from %s in development mode. Model: %s", provider.Name, req.Model),
					},
					"finish_reason": "stop",
				},
			},
			"usage": gin.H{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		}
		ctx.JSON(http.StatusOK, simulatedResponse)
		return
	}

	if provider.APIKey == "" {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_503",
				"message": "Provider not configured",
				"details": fmt.Sprintf("API key not configured for provider: %s", provider.Name),
			},
		})
		return
	}

	// Calculate token count (simplified)
	inputTokens := estimateTokens(req.Messages)
	outputTokens := 0
	if req.MaxTokens != nil {
		outputTokens = *req.MaxTokens
	} else {
		outputTokens = 100 // default
	}

	// Calculate cost
	costResp, err := c.billingService.CalculateCost(ctx, modelObj.ID, inputTokens, outputTokens)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_500",
				"message": "Failed to calculate cost",
			},
		})
		return
	}

	// Check quota limits
	totalTokens := inputTokens + outputTokens
	quotaCheck, err := c.quotaService.CheckQuota(ctx, userID.(string), totalTokens, costResp.TotalCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_500",
				"message": "Failed to check quota",
			},
		})
		return
	}

	if !quotaCheck.Allowed {
		ctx.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_429",
				"message": "Quota limit exceeded",
				"details": quotaCheck.Reason,
				"limits": gin.H{
					"daily_requests": gin.H{
						"used":      quotaCheck.DailyRequests,
						"limit":     quotaCheck.DailyRequestsLimit,
						"remaining": quotaCheck.DailyRequestsLimit - quotaCheck.DailyRequests,
					},
					"daily_tokens": gin.H{
						"used":      quotaCheck.DailyTokens,
						"limit":     quotaCheck.DailyTokensLimit,
						"remaining": quotaCheck.DailyTokensLimit - quotaCheck.DailyTokens,
					},
					"daily_cost": gin.H{
						"used":      quotaCheck.DailyCost,
						"limit":     quotaCheck.DailyCostLimit,
						"remaining": quotaCheck.DailyCostLimit - quotaCheck.DailyCost,
					},
					"monthly_requests": gin.H{
						"used":      quotaCheck.MonthlyRequests,
						"limit":     quotaCheck.MonthlyRequestsLimit,
						"remaining": quotaCheck.MonthlyRequestsLimit - quotaCheck.MonthlyRequests,
					},
					"monthly_tokens": gin.H{
						"used":      quotaCheck.MonthlyTokens,
						"limit":     quotaCheck.MonthlyTokensLimit,
						"remaining": quotaCheck.MonthlyTokensLimit - quotaCheck.MonthlyTokens,
					},
					"monthly_cost": gin.H{
						"used":      quotaCheck.MonthlyCost,
						"limit":     quotaCheck.MonthlyCostLimit,
						"remaining": quotaCheck.MonthlyCostLimit - quotaCheck.MonthlyCost,
					},
				},
				"next_reset": quotaCheck.NextReset,
			},
		})
		return
	}

	// Check user balance
	balance, err := c.billingService.GetBalance(ctx, userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_500",
				"message": "Failed to check balance",
			},
		})
		return
	}

	if balance.Balance < costResp.TotalCost {
		ctx.JSON(http.StatusPaymentRequired, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_402",
				"message": "Insufficient balance",
				"details": fmt.Sprintf("Required: %.6f, Available: %.2f", costResp.TotalCost, balance.Balance),
			},
		})
		return
	}

	// Forward request to provider based on provider type
	providerURL, authField, authValue := buildProviderRequestInfo(provider, modelObj.Name)

	// Create request to provider
	providerReq, err := http.NewRequestWithContext(ctx, "POST", providerURL, bytes.NewReader(requestBody))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_500",
				"message": "Failed to create provider request",
			},
		})
		return
	}

	// Set headers
	providerReq.Header.Set("Content-Type", "application/json")
	if authField != "" && authValue != "" {
		providerReq.Header.Set(authField, authValue)
	}

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(providerReq)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_502",
				"message": "Provider request failed",
				"details": err.Error(),
			},
		})
		return
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_500",
				"message": "Failed to read provider response",
			},
		})
		return
	}

	// Parse response to get token counts
	var providerResp map[string]interface{}
	json.Unmarshal(body, &providerResp)

	// Extract actual token usage from response
	actualInputTokens := inputTokens
	actualOutputTokens := outputTokens

	if usage, ok := providerResp["usage"].(map[string]interface{}); ok {
		if prompt, ok := usage["prompt_tokens"].(float64); ok {
			actualInputTokens = int(prompt)
		}
		if completion, ok := usage["completion_tokens"].(float64); ok {
			actualOutputTokens = int(completion)
		}
	}

	// Create billing record asynchronously via Redis queue
	actualTotalTokens := actualInputTokens + actualOutputTokens
	if err := c.billingService.CreateBillingRecord(ctx, &service.CreateBillingRecordRequest{
		UserID:         userID.(string),
		APIKeyID:       nil, // TODO: Get API key ID from validation
		ModelID:        modelObj.ID,
		RequestTokens:  actualInputTokens,
		ResponseTokens: actualOutputTokens,
		TotalTokens:    actualTotalTokens,
		Cost:           costResp.TotalCost,
		Metadata: map[string]interface{}{
			"provider":   provider.Name,
			"model_name": modelObj.Name,
			"api_key":    ctx.GetHeader("X-API-Key")[:min(8, len(ctx.GetHeader("X-API-Key")))],
			"request_id": generateRequestID(),
		},
	}); err != nil {
		// Log the error but don't fail the request
		fmt.Printf("Failed to create billing record (async): %v\n", err)
	}

	// Record usage for quota tracking (actualTotalTokens is already declared above)
	if err := c.quotaService.RecordUsage(ctx, userID.(string), modelObj.ID, actualTotalTokens, costResp.TotalCost); err != nil {
		// Log the error but don't fail the request - quota was already checked
		fmt.Printf("Failed to record quota usage: %v\n", err)
	}

	// Return provider response
	ctx.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

// estimateTokens estimates token count (simplified)
func estimateTokens(messages []ChatCompletionMessage) int {
	total := 0
	for _, msg := range messages {
		// Rough estimate: 1 token ≈ 4 characters for English
		total += len(msg.Content)/4 + 1
	}
	return total
}

// findModelByName finds a model by its name (e.g., "gpt-4o")
func (c *Controller) findModelByName(ctx *gin.Context, modelName string) (*model.Model, *model.ModelProvider, error) {
	// Get all active models
	modelsResp, err := c.modelService.ListModels(ctx, &service.ListModelsRequest{
		Page:  1,
		Limit: 100,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list models: %w", err)
	}

	// Find model by name
	for _, modelInfo := range modelsResp.Models {
		if modelInfo.Name == modelName {
			// Get model details with provider info
			modelDetails, err := c.modelService.GetModelDetails(ctx, modelInfo.ID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get model details: %w", err)
			}
			return modelDetails.Model, modelDetails.Provider, nil
		}
	}

	return nil, nil, fmt.Errorf("model not found: %s", modelName)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// buildProviderRequestInfo returns the provider URL, authentication header field, and value
// based on the provider type and model name
func buildProviderRequestInfo(provider *model.ModelProvider, modelName string) (string, string, string) {
	baseURL := provider.APIBaseURL
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	// Default to OpenAI-compatible format
	url := baseURL + "chat/completions"
	authField := "Authorization"
	authValue := "Bearer " + provider.APIKey

	// Handle different provider types
	switch strings.ToLower(provider.Name) {
	case "openai":
		// OpenAI uses standard chat/completions endpoint
		url = baseURL + "chat/completions"
		authField = "Authorization"
		authValue = "Bearer " + provider.APIKey
	case "anthropic":
		// Anthropic uses messages endpoint
		url = baseURL + "v1/messages"
		// Anthropic uses x-api-key header instead of Authorization
		authField = "x-api-key"
		authValue = provider.APIKey
	case "google":
		// Google uses generateContent endpoint
		// Extract model name from full model ID (e.g., models/gemini-1.5-pro)
		modelPath := modelName
		if strings.Contains(modelName, "/") {
			// Already has models/ prefix
			modelPath = modelName
		} else {
			modelPath = "models/" + modelName
		}
		url = baseURL + "v1beta/" + modelPath + ":generateContent"
		// Google uses API key in query parameter, not header
		// Add API key as query parameter
		if provider.APIKey != "" {
			if strings.Contains(url, "?") {
				url += "&key=" + provider.APIKey
			} else {
				url += "?key=" + provider.APIKey
			}
		}
		authField = "" // No authorization header for Google
		authValue = ""
	case "meta":
		// Meta (Llama) likely uses OpenAI-compatible API
		url = baseURL + "chat/completions"
		authField = "Authorization"
		authValue = "Bearer " + provider.APIKey
	case "cohere":
		// Cohere uses generate endpoint
		url = baseURL + "v1/generate"
		authField = "Authorization"
		authValue = "Bearer " + provider.APIKey
	}

	return url, authField, authValue
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}
