package oauth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"massrouter.ai/backend/internal/service"
)

type Controller struct {
	oauthService service.OAuthService
	validator    *validator.Validate
}

func NewController(oauthService service.OAuthService) *Controller {
	return &Controller{
		oauthService: oauthService,
		validator:    validator.New(),
	}
}

// GetOAuthProviders godoc
// @Summary Get enabled OAuth providers
// @Description Get list of enabled OAuth providers for frontend authentication
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{} "List of OAuth providers"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/auth/oauth/providers [get]
func (c *Controller) GetOAuthProviders(ctx *gin.Context) {
	providers, err := c.oauthService.GetEnabledProviders(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to get OAuth providers",
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

// StartOAuthFlow godoc
// @Summary Start OAuth authentication flow
// @Description Redirect user to OAuth provider's authorization page
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.StartOAuthFlowRequest true "OAuth flow request"
// @Success 200 {object} map[string]interface{} "OAuth flow started"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/auth/oauth/start [post]
func (c *Controller) StartOAuthFlow(ctx *gin.Context) {
	var req service.StartOAuthFlowRequest
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

	authURL, err := c.oauthService.StartOAuthFlow(ctx.Request.Context(), &req)
	if err != nil {
		status := http.StatusInternalServerError
		errorCode := "ERR_INTERNAL"

		switch err.Error() {
		case "provider not found":
			status = http.StatusNotFound
			errorCode = "ERR_PROVIDER_NOT_FOUND"
		case "provider disabled":
			status = http.StatusBadRequest
			errorCode = "ERR_PROVIDER_DISABLED"
		case "invalid callback URL":
			status = http.StatusBadRequest
			errorCode = "ERR_INVALID_CALLBACK"
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
		"data": gin.H{
			"auth_url": authURL,
		},
	})
}

// HandleOAuthCallback godoc
// @Summary Handle OAuth callback
// @Description Handle OAuth provider callback and authenticate/register user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.HandleOAuthCallbackRequest true "OAuth callback request"
// @Success 200 {object} map[string]interface{} "OAuth authentication successful"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid code or state"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/auth/oauth/callback [post]
func (c *Controller) HandleOAuthCallback(ctx *gin.Context) {
	var req service.HandleOAuthCallbackRequest
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

	response, err := c.oauthService.HandleOAuthCallback(ctx.Request.Context(), &req)
	if err != nil {
		status := http.StatusInternalServerError
		errorCode := "ERR_INTERNAL"

		switch {
		case err.Error() == "invalid authorization code":
			status = http.StatusUnauthorized
			errorCode = "ERR_INVALID_CODE"
		case err.Error() == "invalid state parameter":
			status = http.StatusUnauthorized
			errorCode = "ERR_INVALID_STATE"
		case err.Error() == "state mismatch":
			status = http.StatusUnauthorized
			errorCode = "ERR_STATE_MISMATCH"
		case err.Error() == "email not provided by provider":
			status = http.StatusBadRequest
			errorCode = "ERR_NO_EMAIL"
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
		"data":    response,
	})
}

// DisconnectOAuthAccount godoc
// @Summary Disconnect OAuth account
// @Description Disconnect user's OAuth account from their profile
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.DisconnectOAuthAccountRequest true "Disconnect request"
// @Success 200 {object} map[string]interface{} "Account disconnected successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 404 {object} map[string]interface{} "Not found - account not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/auth/oauth/disconnect [post]
func (c *Controller) DisconnectOAuthAccount(ctx *gin.Context) {
	var req service.DisconnectOAuthAccountRequest
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

	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	err := c.oauthService.DisconnectOAuthAccount(ctx.Request.Context(), userID.(string), &req)
	if err != nil {
		status := http.StatusInternalServerError
		errorCode := "ERR_INTERNAL"

		switch err.Error() {
		case "account not found":
			status = http.StatusNotFound
			errorCode = "ERR_ACCOUNT_NOT_FOUND"
		case "cannot disconnect last authentication method":
			status = http.StatusBadRequest
			errorCode = "ERR_LAST_AUTH_METHOD"
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
		"data": gin.H{
			"message": "OAuth account disconnected successfully",
		},
	})
}
