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
	"syscall"
	"time"

	linkRepo "short-url-sys/internal/repository/link"
	statsRepo "short-url-sys/internal/repository/stats"
)

type Server struct {
	config      *config.Config
	router      http.Handler
	server      *http.Server
	mysqlDB     *database.MySQLDB
	redisClient *database.RedisClient
	linkRepo    linkRepo.Repository
	statsRepo   statsRepo.Repository
	cacheRepo   *cache.Repository
}

func New(config *config.Config, router http.Handler) *Server {
	return &Server{
		config: config,
		router: router,
	}
}

func (s *Server) initDatabase() error {
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

	log.Printf("init database success\n")
	return nil
}

func (s *Server) Start() error {
	// 初始化数据库
	if err := s.initDatabase(); err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}
	apiServer := s.config.Server.APIServer
	addr := fmt.Sprintf("%s:%d", apiServer.Host, apiServer.Port)
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
	}()

	// 等待中断信号
	s.waitForShutdown()
	return nil
}

func (s *Server) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("Shutting down server...\n")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// 关闭数据库连接
	if s.mysqlDB != nil {
		s.mysqlDB.Close()
	}
	if s.redisClient != nil {
		s.redisClient.Close()
	}
	log.Printf("Server exiting...\n")
}

// GetLinkRepository 获取Repository的方法
func (s *Server) GetLinkRepository() linkRepo.Repository {
	return s.linkRepo
}

func (s *Server) GetStatsRepository() statsRepo.Repository {
	return s.statsRepo
}

func (s *Server) GetCacheRepository() *cache.Repository {
	return s.cacheRepo
}

func (s *Server) GetConfig() *config.Config {
	return s.config
}
