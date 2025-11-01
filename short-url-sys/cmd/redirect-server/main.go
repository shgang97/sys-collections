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

	// 创建服务器
	srv := server.NewRedirectServer(cfg)
	// 启动服务器
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start redirect-server: %v", err)
	}
}
