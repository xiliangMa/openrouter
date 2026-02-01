package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"massrouter.ai/backend/pkg/cache"
	"massrouter.ai/backend/pkg/database"
)

type Controller struct {
	db        *database.Database
	redis     *cache.RedisClient
	startedAt time.Time
}

type HealthStatus struct {
	Status    string           `json:"status"`
	Timestamp time.Time        `json:"timestamp"`
	Checks    map[string]Check `json:"checks,omitempty"`
}

type Check struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewController(db *database.Database, redis *cache.RedisClient) *Controller {
	return &Controller{
		db:        db,
		redis:     redis,
		startedAt: time.Now(),
	}
}

// HealthCheck godoc
// @Summary Health check
// @Description Comprehensive health check of all system components
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{} "System is healthy"
// @Success 503 {object} map[string]interface{} "System is degraded"
// @Router /api/v1/health [get]
func (c *Controller) HealthCheck(ctx *gin.Context) {
	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Checks:    make(map[string]Check),
	}

	start := time.Now()
	if err := c.db.HealthCheck(ctx.Request.Context()); err != nil {
		status.Checks["database"] = Check{
			Status: "unhealthy",
			Error:  err.Error(),
		}
		status.Status = "degraded"
	} else {
		status.Checks["database"] = Check{
			Status:  "healthy",
			Latency: time.Since(start).String(),
		}
	}

	if c.redis != nil {
		start = time.Now()
		if err := c.redis.HealthCheck(ctx.Request.Context()); err != nil {
			status.Checks["redis"] = Check{
				Status: "unhealthy",
				Error:  err.Error(),
			}
			status.Status = "degraded"
		} else {
			status.Checks["redis"] = Check{
				Status:  "healthy",
				Latency: time.Since(start).String(),
			}
		}
	}

	status.Checks["uptime"] = Check{
		Status:  "healthy",
		Latency: time.Since(c.startedAt).String(),
	}

	if status.Status == "healthy" {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    status,
		})
	} else {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"data":    status,
		})
	}
}

// ReadinessCheck godoc
// @Summary Readiness check
// @Description Check if system is ready to accept traffic
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{} "System is ready"
// @Success 503 {object} map[string]interface{} "System is not ready"
// @Router /api/v1/ready [get]
func (c *Controller) ReadinessCheck(ctx *gin.Context) {
	status := HealthStatus{
		Status:    "ready",
		Timestamp: time.Now(),
		Checks:    make(map[string]Check),
	}

	start := time.Now()
	if err := c.db.HealthCheck(ctx.Request.Context()); err != nil {
		status.Checks["database"] = Check{
			Status: "unhealthy",
			Error:  err.Error(),
		}
		status.Status = "not ready"
	} else {
		status.Checks["database"] = Check{
			Status:  "ready",
			Latency: time.Since(start).String(),
		}
	}

	if c.redis != nil {
		start = time.Now()
		if err := c.redis.HealthCheck(ctx.Request.Context()); err != nil {
			status.Checks["redis"] = Check{
				Status: "unhealthy",
				Error:  err.Error(),
			}
			status.Status = "not ready"
		} else {
			status.Checks["redis"] = Check{
				Status:  "ready",
				Latency: time.Since(start).String(),
			}
		}
	}

	if status.Status == "ready" {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    status,
		})
	} else {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"data":    status,
		})
	}
}

// LivenessCheck godoc
// @Summary Liveness check
// @Description Check if application is running (no dependencies)
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{} "Application is alive"
// @Router /api/v1/live [get]
func (c *Controller) LivenessCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"status":    "alive",
			"timestamp": time.Now(),
			"uptime":    time.Since(c.startedAt).String(),
		},
	})
}
