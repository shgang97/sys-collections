package stats

import (
	"context"
	"short-url-sys/internal/model"
	"short-url-sys/internal/pkg/errors"
	"time"

	"gorm.io/gorm"
)

type MySQLRepository struct {
	db *gorm.DB
}

func (r *MySQLRepository) RecordClick(ctx context.Context, stats *model.ClickStats) error {
	result := r.db.WithContext(ctx).Create(&stats)
	if err := result.Error; err != nil {
		return &errors.RepositoryError{Operation: "CreateClick", Err: err}
	}
	return nil
}

func (r *MySQLRepository) GetStatsSummary(ctx context.Context, shortCode string, startDate, endDate *time.Time) (*model.StatsSummary, error) {
	var summary model.StatsSummary

	// 基础查询
	query := r.db.WithContext(ctx).Model(&model.ClickStats{}).
		Where("short_code = ?", shortCode)
	// 事件范围过滤
	if startDate != nil {
		query = query.Where("start_date >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("end_date <= ?", endDate)
	}

	// 总点击量
	if err := query.Count(&summary.TotalClicks).Error; err != nil {
		return nil, &errors.RepositoryError{Operation: "GetStatsSummary", Err: err}
	}

	// 如果时间范围为空，默认查询最近30天
	if startDate == nil && endDate == nil {
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		startDate = &thirtyDaysAgo
	}

	// 获取每日统计
	dailyStats, err := r.GetDailyStats(ctx, shortCode, 30)
	if err != nil {
		return nil, &errors.RepositoryError{Operation: "GetStatsSummary", Err: err}
	}
	summary.DailyStats = dailyStats

	// 获取来源统计
	var referrerStats []struct {
		Referrer string
		Count    int64
	}
	if err := query.Select("referer, COUNT(*) as count").Group("referer").Find(&referrerStats).Error; err != nil {
		return nil, &errors.RepositoryError{Operation: "GetStatsSummary", Err: err}
	}
	summary.Referrers = make(map[string]int64)
	for _, stat := range referrerStats {
		if stat.Referrer == "" {
			summary.Referrers["direct"] += stat.Count
		} else {
			summary.Referrers[stat.Referrer] += stat.Count
		}
	}

	// 获取国家统计
	var countryStats []struct {
		Country string
		Count   int64
	}
	if err := query.Select("country, COUNT(*) as count").Group("country").Find(&countryStats).Error; err != nil {
		return nil, &errors.RepositoryError{Operation: "GetStatsSummary", Err: err}
	}
	summary.Countries = make(map[string]int64)
	for _, stat := range countryStats {
		if stat.Country != "" {
			summary.Countries[stat.Country] = stat.Count
		}
	}

	// 获取设备统计
	var deviceStats []struct {
		DeviceType string
		Count      int64
	}
	if err := query.Select("device_type, COUNT(*) as count").Group("device_type").Find(&deviceStats).Error; err != nil {
		return nil, &errors.RepositoryError{Operation: "GetStatsSummary", Err: err}
	}
	summary.Devices = make(map[string]int64)
	for _, stat := range deviceStats {
		if stat.DeviceType != "" {
			summary.Devices[stat.DeviceType] = stat.Count
		}
	}

	return &summary, nil
}

func (r *MySQLRepository) GetDailyStats(ctx context.Context, shortCode string, days int) ([]model.DailyStats, error) {
	var stats []model.DailyStats

	startDate := time.Now().AddDate(0, 0, -days)
	query := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as clicks,
			COUNT(DISTINCT ip_address) as ip_address,
		FROM click_stats
		WHERE short_code = ? AND created_at >= ?
		GROUP BY date
		ORDER BY date DESC
	`

	result := r.db.WithContext(ctx).Table("click_stats").Raw(query, shortCode, startDate).Scan(&stats)
	if err := result.Error; err != nil {
		return nil, &errors.RepositoryError{Operation: "GetDailyStats", Err: err}
	}
	return stats, nil
}

func (r *MySQLRepository) GetLastAccessed(ctx context.Context, shortCode string) (*time.Time, error) {
	var lastAccessed time.Time

	result := r.db.WithContext(ctx).Model(&model.ClickStats{}).
		Select("MAX(created_at)").
		Where("short_code = ?", shortCode).
		Scan(&lastAccessed)
	if err := result.Error; err != nil {
		return nil, &errors.RepositoryError{Operation: "GetLastAccessed", Err: err}
	}
	if lastAccessed.IsZero() {
		return nil, nil
	}
	return &lastAccessed, nil
}

func NewMySQLRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}
