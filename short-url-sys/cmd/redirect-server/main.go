package main

import (
	"log"
	"short-url-sys/internal/config"
	"short-url-sys/internal/server"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 重定向服务使用不同的端口配置
	cfg.Server.APIServer.Port = cfg.Server.RedirectServer.Port
	cfg.Server.APIServer.Host = cfg.Server.RedirectServer.Host
	cfg.Server.APIServer.Mode = cfg.Server.RedirectServer.Mode

	// 设置路由
	router := server.SetupRouter(cfg)

	// 创建服务器
	srv := server.New(cfg, router)

	// 启动服务器
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start redirect-server: %v", err)
	}
}
