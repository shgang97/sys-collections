package server

import (
	"short-url-sys/internal/config"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRouter(config *config.Config) *gin.Engine {
	// 设置Gin模式
	gin.SetMode(config.Server.APIServer.Mode)
	router := gin.New()

	// 全局中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	api := router.Group("/api/v1")
	{
		api.GET("/info", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"service": "short-url-sys",
				"version": "1.0.0",
			})
		})
	}
	return router
}
