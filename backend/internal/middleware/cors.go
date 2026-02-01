package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			c.Next()
			return
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Disposition")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			c.Next()
			return
		}

		allowed := false
		for _, allowedOrigin := range config.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin ||
				(strings.HasPrefix(allowedOrigin, "*.") && strings.HasSuffix(origin, allowedOrigin[1:])) {
				allowed = true
				break
			}
		}

		if !allowed {
			c.Next()
			return
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		if config.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if len(config.AllowedHeaders) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		}

		if len(config.AllowedMethods) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		}

		if config.ExposeHeaders != "" {
			c.Writer.Header().Set("Access-Control-Expose-Headers", config.ExposeHeaders)
		}

		if c.Request.Method == "OPTIONS" {
			if config.MaxAge != "" {
				c.Writer.Header().Set("Access-Control-Max-Age", config.MaxAge)
			}
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposeHeaders    string
	AllowCredentials bool
	MaxAge           string
}
