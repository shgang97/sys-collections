package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"short-url-sys/internal/config"
	"short-url-sys/internal/pkg/database"
	"short-url-sys/internal/repository/cache"
	linkRepo "short-url-sys/internal/repository/link"
	statsRepo "short-url-sys/internal/repository/stats"
	redirectService "short-url-sys/internal/service/redirect"
	"syscall"
	"time"
)

type RedirectServer struct {
	config      *config.Config
	router      http.Handler
	server      *http.Server
	mysqlDB     *database.MySQLDB
	redisClient *database.RedisClient
	redirectSvc redirectService.Service
	linkRepo    linkRepo.Repository
	statsRepo   statsRepo.Repository
	cacheRepo   *cache.Repository
}

func NewRedirectServer(config *config.Config) *RedirectServer {
	return &RedirectServer{
		config: config,
	}
}

func (s *RedirectServer) initDatabase() error {
	// 初始化MySQL
	mysqlDB, err := database.NewSQLDB(&s.config.Database.MySQL)
	if err != nil {
		return fmt.Errorf("init mysql failed: %w", err)
	}
	s.mysqlDB = mysqlDB

	// 初始化Redis
	redisClient, err := database.NewRedis(&s.config.Redis)
	if err != nil {
		return fmt.Errorf("init redis failed: %w", err)
	}
	s.redisClient = redisClient

	// 初始化Repository
	s.linkRepo = linkRepo.NewMySQLRepository(mysqlDB.DB)
	s.statsRepo = statsRepo.NewMySQLRepository(mysqlDB.DB)
	s.cacheRepo = cache.NewRepository(redisClient.Client, &s.config.Cache)

	log.Printf("✅ init database success\n")
	return nil
}

func (s *RedirectServer) initServices() error {

	// 初始化重定向服务
	s.redirectSvc = redirectService.NewRedirectRequest(
		s.linkRepo,
		s.statsRepo,
		s.cacheRepo,
	)

	log.Println("✅ Services initialized successfully")
	return nil
}

func (s *RedirectServer) Start() error {
	// 初始化数据库
	if err := s.initDatabase(); err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}

	// 初始化服务
	if err := s.initServices(); err != nil {
		return fmt.Errorf("failed to init services: %w", err)
	}

	// 设置路由
	router := SetupRedirectRouter(s.config, s)
	s.router = router
	redirectServer := s.config.Server.RedirectServer
	addr := fmt.Sprintf("%s:%d", redirectServer.Host, redirectServer.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	go func() {
		log.Printf("server listening on %s", addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
		log.Printf("server started")
	}()

	// 等待中断信号
	s.waitForShutdown()
	return nil
}

func (s *RedirectServer) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("Shutting down server...\n")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Printf("Server exiting...\n")
}
