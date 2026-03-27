package server

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"comb-dockerfile/internal/handler/health"

	healthRes "comb-dockerfile/internal/handler/health/res"
	_ "comb-dockerfile/docs"

)

// SetupRouter 初始化gin入口，路由信息
func (s *Server) SetupRouter() {
	// 客户端通过变量进行传递
	v1Router := s.Gin.Group("/api/v1")
	healthRouter(v1Router)

	s.Gin.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func healthRouter(group *gin.RouterGroup) {
	healthRes.RegisterCode()
	group.GET("/health", health.HealthHandler())
}
