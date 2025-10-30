package server

import (
	"short-url-sys/internal/config"
	"short-url-sys/internal/handler"
	"short-url-sys/internal/model"
	"short-url-sys/internal/server/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRouter(config *config.Config, srv *Server) {
	// 设置Gin模式
	gin.SetMode(config.Server.APIServer.Mode)
	router := gin.New()

	// 全局中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.ErrorHandler())

	// 初始化处理器
	linkHandler := handler.NewLinkHandler(srv.linkSvc, config.Server.APIServer.BaseURL)
	redirectHandler := handler.NewRedirectHandler(srv.redirectSvc)

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		health := model.HealthResponse{
			Status:    "ok",
			Timestamp: time.Now().Unix(),
			Services:  make(map[string]string),
		}

		// 检查MySQL连接
		if srv.mysqlDB != nil {
			if err := srv.mysqlDB.HealthCheck(); err != nil {
				health.Status = "degraded"
				health.Services["mysql"] = "unhealthy"
			} else {
				health.Services["mysql"] = "healthy"
			}
		}

		// 检查Redis连接
		if srv.redisClient != nil {
			if err := srv.redisClient.HealthCheck(); err != nil {
				health.Status = "degraded"
				health.Services["redis"] = "unhealthy"
			} else {
				health.Services["redis"] = "healthy"
			}
		}
		c.JSON(200, health)
	})

	api := router.Group("/api/v1")
	{
		links := api.Group("/links")
		{
			links.POST("/short", linkHandler.CreateShortURL)
			links.GET("/:code", linkHandler.GetLinkInfo)
			links.PUT("/:code", linkHandler.UpdateLink)
			links.DELETE("/:code", linkHandler.DeleteLink)
			links.GET("", linkHandler.ListLinks)
			links.POST("/short/batch", linkHandler.BatchCreate)
		}

		api.GET("/info", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"service": "short-url-sys",
				"version": "1.0.0",
			})
		})
	}

	// 重定向路由（放在最后，避免被API路由捕获）
	router.GET("/:code", redirectHandler.Redirect)
	srv.router = router
}
