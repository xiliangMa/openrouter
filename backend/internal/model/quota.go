package model

import (
	"time"
)

// UserQuota represents the quota limits for a user
type UserQuota struct {
	ID     string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID string `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`

	// Daily limits
	DailyRequestLimit int     `gorm:"not null;default:100" json:"daily_request_limit"`                   // Maximum requests per day
	DailyTokenLimit   int     `gorm:"not null;default:100000" json:"daily_token_limit"`                  // Maximum tokens per day
	DailyCostLimit    float64 `gorm:"type:decimal(10,4);not null;default:10.00" json:"daily_cost_limit"` // Maximum cost per day

	// Monthly limits
	MonthlyRequestLimit int     `gorm:"not null;default:3000" json:"monthly_request_limit"`                   // Maximum requests per month
	MonthlyTokenLimit   int     `gorm:"not null;default:3000000" json:"monthly_token_limit"`                  // Maximum tokens per month
	MonthlyCostLimit    float64 `gorm:"type:decimal(10,4);not null;default:300.00" json:"monthly_cost_limit"` // Maximum cost per month

	// Model-specific limits (stored as JSON for flexibility)
	ModelLimits JSONB `gorm:"type:jsonb;not null;default:'{}'" json:"model_limits"`

	// Rate limiting
	PerMinuteRateLimit int `gorm:"not null;default:60" json:"per_minute_rate_limit"` // Requests per minute
	PerHourRateLimit   int `gorm:"not null;default:1000" json:"per_hour_rate_limit"` // Requests per hour

	// Reset configuration
	ResetDay int    `gorm:"not null;default:1" json:"reset_day"`                      // Day of month for monthly reset (1-31)
	Timezone string `gorm:"type:varchar(100);not null;default:'UTC'" json:"timezone"` // Timezone for reset calculations

	// Status
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// UserUsage represents daily usage tracking for a user
type UserUsage struct {
	ID     string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Date   time.Time `gorm:"type:date;not null;index" json:"date"` // Date for daily usage

	// Daily usage counts
	RequestCount int     `gorm:"not null;default:0" json:"request_count"`
	TokenCount   int     `gorm:"not null;default:0" json:"token_count"`
	TotalCost    float64 `gorm:"type:decimal(10,8);not null;default:0" json:"total_cost"`

	// Model-specific usage (stored as JSON for flexibility)
	ModelUsage JSONB `gorm:"type:jsonb;not null;default:'{}'" json:"model_usage"`

	// Peak usage tracking
	PeakRequestsPerMinute int `gorm:"not null;default:0" json:"peak_requests_per_minute"`
	PeakTokensPerMinute   int `gorm:"not null;default:0" json:"peak_tokens_per_minute"`

	// Status
	IsExceeded bool      `gorm:"not null;default:false" json:"is_exceeded"` // True if any limit was exceeded
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt  time.Time `gorm:"not null" json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// MonthlyUsage represents monthly usage summary for a user
type MonthlyUsage struct {
	ID        string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID    string `gorm:"type:uuid;not null;index" json:"user_id"`
	YearMonth string `gorm:"type:varchar(7);not null;index" json:"year_month"` // Format: "2025-01"

	// Monthly usage counts
	RequestCount int     `gorm:"not null;default:0" json:"request_count"`
	TokenCount   int     `gorm:"not null;default:0" json:"token_count"`
	TotalCost    float64 `gorm:"type:decimal(10,4);not null;default:0" json:"total_cost"`

	// Model-specific usage (stored as JSON for flexibility)
	ModelUsage JSONB `gorm:"type:jsonb;not null;default:'{}'" json:"model_usage"`

	// Status
	IsExceeded bool      `gorm:"not null;default:false" json:"is_exceeded"` // True if any limit was exceeded
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt  time.Time `gorm:"not null" json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// QuotaExceededError represents a quota limit exceeded error
type QuotaExceededError struct {
	LimitType string    `json:"limit_type"` // "daily_requests", "daily_tokens", "daily_cost", "monthly_requests", "monthly_tokens", "monthly_cost"
	Limit     float64   `json:"limit"`
	Used      float64   `json:"used"`
	Remaining float64   `json:"remaining"`
	ResetTime time.Time `json:"reset_time"`
}

func (e *QuotaExceededError) Error() string {
	return "quota limit exceeded"
}

func (UserQuota) TableName() string {
	return "user_quotas"
}

func (UserUsage) TableName() string {
	return "user_usage"
}

func (MonthlyUsage) TableName() string {
	return "monthly_usage"
}

// DefaultQuota returns the default quota configuration
func DefaultQuota() UserQuota {
	return UserQuota{
		DailyRequestLimit:   100,
		DailyTokenLimit:     100000,
		DailyCostLimit:      10.00,
		MonthlyRequestLimit: 3000,
		MonthlyTokenLimit:   3000000,
		MonthlyCostLimit:    300.00,
		PerMinuteRateLimit:  60,
		PerHourRateLimit:    1000,
		ResetDay:            1,
		Timezone:            "UTC",
		ModelLimits:         JSONB{},
		IsActive:            true,
	}
}
