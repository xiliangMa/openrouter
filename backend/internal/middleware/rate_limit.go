package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

func (rl *RateLimiter) Limit(requests int, window time.Duration, keyFunc func(c *gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rl.client == nil {
			c.Next()
			return
		}

		key := keyFunc(c)
		if key == "" {
			c.Next()
			return
		}

		ctx := context.Background()
		now := time.Now().UnixNano()
		windowNanos := window.Nanoseconds()
		clearBefore := now - windowNanos

		pipe := rl.client.Pipeline()
		// Individual pipeline command errors are handled by pipe.Exec()
		_ = pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(clearBefore, 10))
		_ = pipe.ZAdd(ctx, key, redis.Z{
			Score:  float64(now),
			Member: fmt.Sprintf("%d:%s", now, c.ClientIP()),
		})
		zcard := pipe.ZCard(ctx, key)
		pipe.Expire(ctx, key, window)

		_, err := pipe.Exec(ctx)
		if err != nil {
			c.Next()
			return
		}

		count, err := zcard.Result()
		if err != nil {
			c.Next()
			return
		}

		if count > int64(requests) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "ERR_429",
					"message": "Rate limit exceeded",
					"details": fmt.Sprintf("Limit: %d requests per %v", requests, window),
				},
			})
			return
		}

		remaining := int64(requests) - count
		resetTime := time.Unix(0, now+windowNanos)

		c.Writer.Header().Set("X-RateLimit-Limit", strconv.Itoa(requests))
		c.Writer.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Writer.Header().Set("X-RateLimit-Reset", resetTime.Format(time.RFC3339))

		c.Next()
	}
}

func IPBasedKey(c *gin.Context) string {
	return "rate_limit:" + c.ClientIP()
}

func UserBasedKey(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	return "rate_limit:user:" + fmt.Sprintf("%v", userID)
}

func APIBasedKey(c *gin.Context) string {
	apiKey := c.Request.Header.Get("Authorization")
	if apiKey == "" {
		return ""
	}
	return "rate_limit:api:" + apiKey
}

func DefaultRateLimit(client *redis.Client) gin.HandlerFunc {
	rl := NewRateLimiter(client)
	return rl.Limit(100, time.Minute, IPBasedKey)
}

func AuthRateLimit(client *redis.Client) gin.HandlerFunc {
	rl := NewRateLimiter(client)
	return rl.Limit(1000, time.Hour, UserBasedKey)
}

func APIRateLimit(client *redis.Client) gin.HandlerFunc {
	rl := NewRateLimiter(client)
	return rl.Limit(10000, time.Hour, APIBasedKey)
}
