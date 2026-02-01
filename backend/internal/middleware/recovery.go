package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func Recovery(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()

				requestID, _ := c.Get("request_id")
				logger.Error().
					Str("request_id", fmt.Sprintf("%v", requestID)).
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Str("client_ip", c.ClientIP()).
					Interface("panic", err).
					Bytes("stack", stack).
					Msg("Panic recovered")

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "ERR_500",
						"message": "Internal server error",
						"details": "An unexpected error occurred",
					},
				})
			}
		}()

		c.Next()
	}
}

func RecoveryWithConfig(logger zerolog.Logger, printStack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()

				requestID, _ := c.Get("request_id")
				logEntry := logger.Error().
					Str("request_id", fmt.Sprintf("%v", requestID)).
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Str("client_ip", c.ClientIP()).
					Interface("panic", err)

				if printStack {
					logEntry = logEntry.Bytes("stack", stack)
				}

				logEntry.Msg("Panic recovered")

				response := gin.H{
					"success": false,
					"error": gin.H{
						"code":    "ERR_500",
						"message": "Internal server error",
					},
				}

				if printStack && gin.Mode() == gin.DebugMode {
					response["error"].(gin.H)["stack"] = string(stack)
				}

				c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			}
		}()

		c.Next()
	}
}
