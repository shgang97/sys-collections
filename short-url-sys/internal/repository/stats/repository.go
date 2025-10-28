package stats

import (
	"context"
	"short-url-sys/internal/model"
	"time"
)

type Repository interface {
	// RecordClick 记录点击
	RecordClick(ctx context.Context, stats *model.ClickStats) error

	// GetStatsSummary 查询统计
	GetStatsSummary(ctx context.Context, shortCode string, startDate, endDate *time.Time) (*model.StatsSummary, error)
	GetDailyStats(ctx context.Context, shortCode string, days int) ([]model.DailyStats, error)
	GetLastAccessed(ctx context.Context, shortCode string) (*time.Time, error)
}
