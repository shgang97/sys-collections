package server

import (
	"short-url-sys/internal/config"
	"short-url-sys/internal/handler"
	"short-url-sys/internal/model"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRedirectRouter(config *config.Config, srv *RedirectServer) *gin.Engine {
	// 设置Gin模式
	gin.SetMode(config.Server.RedirectServer.Mode)
	router := gin.New()

	// 设置全局中间件
	//router.Use(gin.Recovery())
	//router.Use(middleware.ErrorHandler())

	// 初始化处理器
	redirectHandler := handler.NewRedirectHandler(srv.redirectSvc)

	// 健康检查点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "redirect",
			"timestamp": time.Now().Unix(),
		})
	})

	// 重定向路由
	router.GET("/:code", redirectHandler.Redirect)

	// 跟路径
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "short-url-redirect",
			"version": "1.0.0",
		})
	})

	// 404处理
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, model.ErrorResponse{
			Error:   "not found",
			Message: "Short link not found",
		})
	})

	return router
}
