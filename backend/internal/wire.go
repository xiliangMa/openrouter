//go:build wireinject
// +build wireinject

package internal

import (
	"massrouter.ai/backend/internal/config"
	"massrouter.ai/backend/internal/controller/admin"
	"massrouter.ai/backend/internal/controller/auth"
	"massrouter.ai/backend/internal/controller/billing"
	"massrouter.ai/backend/internal/controller/health"
	"massrouter.ai/backend/internal/controller/model"
	"massrouter.ai/backend/internal/controller/user"
	"massrouter.ai/backend/internal/repository"
	"massrouter.ai/backend/internal/service"
	"massrouter.ai/backend/pkg/auth"
	"massrouter.ai/backend/pkg/cache"
	"massrouter.ai/backend/pkg/database"

	"github.com/google/wire"
)

var RepositorySet = wire.NewSet(
	repository.NewUserRepository,
	repository.NewModelRepository,
	repository.NewModelProviderRepository,
	repository.NewUserAPIKeyRepository,
	repository.NewBillingRecordRepository,
	repository.NewPaymentRecordRepository,
	repository.NewModelStatisticRepository,
	repository.NewSystemConfigRepository,
)

var ServiceSet = wire.NewSet(
	service.NewAuthService,
	service.NewUserService,
	service.NewModelService,
	service.NewBillingService,
	service.NewAdminService,
)

var ControllerSet = wire.NewSet(
	auth.NewController,
	user.NewController,
	model.NewController,
	billing.NewController,
	health.NewController,
	admin.NewController,
)

func InitializeServer(cfg *config.Config) (*Server, error) {
	wire.Build(
		database.NewPostgresDB,
		cache.NewRedisClient,
		jwt.NewJWTManager,
		RepositorySet,
		ServiceSet,
		ControllerSet,
		NewServer,
	)
	var server *Server
	return server, nil
}
