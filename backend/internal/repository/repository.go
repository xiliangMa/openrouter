package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"massrouter.ai/backend/internal/model"
)

type BaseRepository[T any] interface {
	FindByID(ctx context.Context, id string) (*T, error)
	FindAll(ctx context.Context, limit, offset int) ([]*T, error)
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
	Exists(ctx context.Context, id string) (bool, error)
	GetDB() *gorm.DB
}

type UserRepository interface {
	BaseRepository[model.User]
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	UpdateLastLogin(ctx context.Context, userID string) error
	UpdateStatus(ctx context.Context, userID, status string) error
	FindWithAPIKeys(ctx context.Context, userID string) (*model.User, error)
	FindWithOAuthAccounts(ctx context.Context, userID string) (*model.User, error)
}

type OAuthProviderRepository interface {
	BaseRepository[model.OAuthProvider]
	FindByName(ctx context.Context, name string) (*model.OAuthProvider, error)
	FindEnabledProviders(ctx context.Context) ([]*model.OAuthProvider, error)
}

type OAuthAccountRepository interface {
	BaseRepository[model.OAuthAccount]
	FindByProviderAndUserID(ctx context.Context, providerID, userID string) (*model.OAuthAccount, error)
	FindByProviderUserID(ctx context.Context, providerID, providerUserID string) (*model.OAuthAccount, error)
	FindByUserID(ctx context.Context, userID string) ([]*model.OAuthAccount, error)
	UpdateTokens(ctx context.Context, accountID, accessToken, refreshToken string, expiresAt *time.Time) error
}

type ModelProviderRepository interface {
	BaseRepository[model.ModelProvider]
	FindByName(ctx context.Context, name string) (*model.ModelProvider, error)
	FindActiveProviders(ctx context.Context) ([]*model.ModelProvider, error)
	UpdateAPIKey(ctx context.Context, providerID, apiKey string) error
	UpdateStatus(ctx context.Context, providerID, status string) error
}

type ModelRepository interface {
	BaseRepository[model.Model]
	FindByProviderAndName(ctx context.Context, providerID string, name string) (*model.Model, error)
	FindActiveModels(ctx context.Context) ([]*model.Model, error)
	FindByCategory(ctx context.Context, category string) ([]*model.Model, error)
	SearchModels(ctx context.Context, query string, limit, offset int) ([]*model.Model, error)
	UpdatePricing(ctx context.Context, modelID string, inputPrice, outputPrice float64) error
	UpdateStatus(ctx context.Context, modelID string, isActive bool) error
}

type UserAPIKeyRepository interface {
	BaseRepository[model.UserAPIKey]
	FindByAPIKey(ctx context.Context, apiKey string) (*model.UserAPIKey, error)
	FindByUserID(ctx context.Context, userID string) ([]*model.UserAPIKey, error)
	FindActiveKeysByUserID(ctx context.Context, userID string) ([]*model.UserAPIKey, error)
	UpdateLastUsed(ctx context.Context, keyID string) error
	RevokeKey(ctx context.Context, keyID string) error
	ValidateKey(ctx context.Context, apiKey string) (*model.UserAPIKey, error)
}

type PaymentRecordRepository interface {
	BaseRepository[model.PaymentRecord]
	FindByUserID(ctx context.Context, userID string) ([]*model.PaymentRecord, error)
	FindByTransactionID(ctx context.Context, transactionID string) (*model.PaymentRecord, error)
	FindByStatus(ctx context.Context, status string) ([]*model.PaymentRecord, error)
	UpdateStatus(ctx context.Context, paymentID, status string, paidAt *time.Time) error
	GetUserTotalPaid(ctx context.Context, userID string) (float64, error)
}

type BillingRecordRepository interface {
	BaseRepository[model.BillingRecord]
	FindByUserID(ctx context.Context, userID string) ([]*model.BillingRecord, error)
	FindByAPIKeyID(ctx context.Context, apiKeyID string) ([]*model.BillingRecord, error)
	FindByModelID(ctx context.Context, modelID string) ([]*model.BillingRecord, error)
	GetUserUsage(ctx context.Context, userID string, startDate, endDate *time.Time) ([]*model.BillingRecord, error)
	GetTotalCostByUser(ctx context.Context, userID string) (float64, error)
	GetDailyUsage(ctx context.Context, userID string, date time.Time) ([]*model.BillingRecord, error)
}

type ModelStatisticRepository interface {
	BaseRepository[model.ModelStatistic]
	FindByModelAndDate(ctx context.Context, modelID string, date time.Time) (*model.ModelStatistic, error)
	FindByModelID(ctx context.Context, modelID string, limit, offset int) ([]*model.ModelStatistic, error)
	GetDailyStatistics(ctx context.Context, date time.Time) ([]*model.ModelStatistic, error)
	UpdateStatistics(ctx context.Context, modelID string, date time.Time, requests, tokens int, avgResponseTime, successRate float64) error
	GetTopModels(ctx context.Context, limit int, startDate, endDate time.Time) ([]*model.ModelStatistic, error)
}

type SystemConfigRepository interface {
	BaseRepository[model.SystemConfig]
	FindByKey(ctx context.Context, key string) (*model.SystemConfig, error)
	FindPublicConfigs(ctx context.Context) ([]*model.SystemConfig, error)
	UpdateValue(ctx context.Context, key, value string) error
}

type MigrationRecordRepository interface {
	BaseRepository[model.MigrationRecord]
	FindLatestBatch(ctx context.Context) (int, error)
	FindByBatch(ctx context.Context, batch int) ([]*model.MigrationRecord, error)
	MigrationApplied(ctx context.Context, name string, batch int) error
}

type UserQuotaRepository interface {
	BaseRepository[model.UserQuota]
	FindByUserID(ctx context.Context, userID string) (*model.UserQuota, error)
	UpdateQuota(ctx context.Context, userID string, quota *model.UserQuota) error
	UpdateModelLimits(ctx context.Context, userID string, modelLimits model.JSONB) error
	FindActiveQuotas(ctx context.Context) ([]*model.UserQuota, error)
}

type UserUsageRepository interface {
	BaseRepository[model.UserUsage]
	FindByUserAndDate(ctx context.Context, userID string, date time.Time) (*model.UserUsage, error)
	FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.UserUsage, error)
	IncrementUsage(ctx context.Context, userID string, date time.Time, requests, tokens int, cost float64, modelUsage model.JSONB) error
	GetUsageForPeriod(ctx context.Context, userID string, startDate, endDate time.Time) ([]*model.UserUsage, error)
	ResetDailyUsage(ctx context.Context, date time.Time) error
	FindExceededUsage(ctx context.Context, date time.Time) ([]*model.UserUsage, error)
}

type MonthlyUsageRepository interface {
	BaseRepository[model.MonthlyUsage]
	FindByUserAndMonth(ctx context.Context, userID string, yearMonth string) (*model.MonthlyUsage, error)
	FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.MonthlyUsage, error)
	IncrementUsage(ctx context.Context, userID string, yearMonth string, requests, tokens int, cost float64, modelUsage model.JSONB) error
	GetMonthlySummary(ctx context.Context, yearMonth string) ([]*model.MonthlyUsage, error)
	ResetMonthlyUsage(ctx context.Context, yearMonth string) error
}
