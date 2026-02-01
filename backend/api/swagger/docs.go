package swagger

import (
	_ "massrouter.ai/backend/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterSwagger(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func SetupSwaggerRoutes(api *gin.RouterGroup) {
	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
