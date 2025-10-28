package database

import (
	"fmt"
	"log"
	"short-url-sys/internal/config"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MySQLDB struct {
	DB *gorm.DB
}

func NewSQLDB(cfg *config.MySQLConfig) (*MySQLDB, error) {
	// 常见 GORM 配置
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(cfg.DSN), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 连接吃配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime * time.Second)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to database\n")
	return &MySQLDB{DB: db}, nil
}

func (m *MySQLDB) Close() error {
	sqlDB, err := m.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// HealthCheck 检查数据库连接状态
func (m *MySQLDB) HealthCheck() error {
	sqlDB, err := m.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
