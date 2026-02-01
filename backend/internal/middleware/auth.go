package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"massrouter.ai/backend/pkg/auth"
)

const (
	UserIDKey  = "user_id"
	UserKey    = "user"
	RoleKey    = "role"
	IsAdminKey = "is_admin"
)

var (
	ErrNoTokenProvided = errors.New("no token provided")
	ErrInvalidToken    = errors.New("invalid token")
	ErrInvalidClaims   = errors.New("invalid claims")
)

func JWTAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "ERR_401",
					"message": "Authentication required",
					"details": "No token provided",
				},
			})
			return
		}

		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "ERR_401",
					"message": "Authentication failed",
					"details": "Invalid or expired token",
				},
			})
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UserKey, claims.Username)
		c.Set(RoleKey, claims.Role)
		c.Set(IsAdminKey, claims.Role == "admin")

		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		return c.Query("token")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(RoleKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "ERR_401",
					"message": "Authentication required",
				},
			})
			return
		}

		if userRole != role && userRole != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "ERR_403",
					"message": "Insufficient permissions",
					"details": fmt.Sprintf("Required role: %s", role),
				},
			})
			return
		}

		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}

func RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(RoleKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "ERR_401",
					"message": "Authentication required",
				},
			})
			return
		}

		if userRole == "admin" {
			c.Next()
			return
		}

		for _, role := range roles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_403",
				"message": "Insufficient permissions",
				"details": fmt.Sprintf("Required one of roles: %v", roles),
			},
		})
	}
}

func OptionalAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			c.Next()
			return
		}

		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UserKey, claims.Username)
		c.Set(RoleKey, claims.Role)
		c.Set(IsAdminKey, claims.Role == "admin")

		c.Next()
	}
}

func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return "", false
	}
	return userID.(string), true
}

func GetRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(RoleKey)
	if !exists {
		return "", false
	}
	return role.(string), true
}

func IsAdmin(c *gin.Context) bool {
	isAdmin, exists := c.Get(IsAdminKey)
	if !exists {
		return false
	}
	return isAdmin.(bool)
}
