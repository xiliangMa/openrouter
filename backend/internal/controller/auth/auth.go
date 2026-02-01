package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"massrouter.ai/backend/internal/service"
)

type Controller struct {
	authService service.AuthService
	validator   *validator.Validate
}

func NewController(authService service.AuthService) *Controller {
	return &Controller{
		authService: authService,
		validator:   validator.New(),
	}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account with email, username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.RegisterRequest true "Register request"
// @Success 201 {object} map[string]interface{} "Registration successful"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 409 {object} map[string]interface{} "Conflict - email or username already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/auth/register [post]
func (c *Controller) Register(ctx *gin.Context) {
	var req service.RegisterRequest
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

	user, err := c.authService.Register(ctx.Request.Context(), &req)
	if err != nil {
		status := http.StatusInternalServerError
		errorCode := "ERR_INTERNAL"

		switch err.Error() {
		case "email already registered":
			status = http.StatusConflict
			errorCode = "ERR_EMAIL_EXISTS"
		case "username already taken":
			status = http.StatusConflict
			errorCode = "ERR_USERNAME_EXISTS"
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

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"user":    user,
			"message": "Registration successful. Please check your email to verify your account.",
		},
	})
}

// Login godoc
// @Summary User login
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "Login request"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid credentials"
// @Failure 403 {object} map[string]interface{} "Forbidden - account suspended or deleted"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/auth/login [post]
func (c *Controller) Login(ctx *gin.Context) {
	var req service.LoginRequest
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

	response, err := c.authService.Login(ctx.Request.Context(), &req)
	if err != nil {
		status := http.StatusInternalServerError
		errorCode := "ERR_INTERNAL"

		switch {
		case err.Error() == "invalid credentials":
			status = http.StatusUnauthorized
			errorCode = "ERR_INVALID_CREDENTIALS"
		case err.Error() == "account is suspended":
			status = http.StatusForbidden
			errorCode = "ERR_ACCOUNT_SUSPENDED"
		case err.Error() == "account is deleted":
			status = http.StatusForbidden
			errorCode = "ERR_ACCOUNT_DELETED"
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

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Refresh token request"
// @Success 200 {object} map[string]interface{} "Token refresh successful"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized - refresh token expired or invalid"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/auth/refresh [post]
func (c *Controller) RefreshToken(ctx *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

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

	response, err := c.authService.RefreshToken(ctx.Request.Context(), req.RefreshToken)
	if err != nil {
		status := http.StatusInternalServerError
		errorCode := "ERR_INTERNAL"

		if err.Error() == "refresh token expired" {
			status = http.StatusUnauthorized
			errorCode = "ERR_TOKEN_EXPIRED"
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

// Logout godoc
// @Summary User logout
// @Description Invalidate user's refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Logout successful"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/auth/logout [post]
func (c *Controller) Logout(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"message": "Logged out successfully",
			},
		})
		return
	}

	if err := c.authService.Logout(ctx.Request.Context(), userID.(string)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to logout",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Logged out successfully",
		},
	})
}

// VerifyEmail godoc
// @Summary Verify email address
// @Description Verify user's email address with token
// @Tags auth
// @Accept json
// @Produce json
// @Param token query string true "Verification token"
// @Success 200 {object} map[string]interface{} "Email verification endpoint"
// @Failure 400 {object} map[string]interface{} "Bad request - token missing"
// @Router /api/v1/auth/verify [get]
func (c *Controller) VerifyEmail(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Verification token is required",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Email verification endpoint. Token validation logic to be implemented.",
		},
	})
}

// RequestPasswordReset godoc
// @Summary Request password reset
// @Description Send password reset email to user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Password reset request"
// @Success 200 {object} map[string]interface{} "Reset email sent if account exists"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/auth/password/reset/request [post]
func (c *Controller) RequestPasswordReset(ctx *gin.Context) {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

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

	if err := c.authService.RequestPasswordReset(ctx.Request.Context(), req.Email); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to process password reset request",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "If your email is registered, you will receive a password reset link shortly.",
		},
	})
}

// ResetPassword godoc
// @Summary Reset password with token
// @Description Reset user password using verification token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Password reset confirmation"
// @Success 200 {object} map[string]interface{} "Password reset successful"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/auth/password/reset [post]
func (c *Controller) ResetPassword(ctx *gin.Context) {
	var req struct {
		Token       string `json:"token" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}

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

	if err := c.authService.ResetPassword(ctx.Request.Context(), req.Token, req.NewPassword); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to reset password",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Password has been reset successfully. You can now login with your new password.",
		},
	})
}
