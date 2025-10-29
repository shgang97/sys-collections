package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type LinkStatus string

const (
	LinkStatusActive   LinkStatus = "active"
	LinkStatusDisabled LinkStatus = "disabled"
	LinkStatusExpired  LinkStatus = "expired"
)

// 实现数据库接口扫描
func (ls *LinkStatus) Scan(value interface{}) error {
	if value == nil {
		*ls = LinkStatusActive
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*ls = LinkStatus(v)
	case string:
		*ls = LinkStatus(v)
	default:
		return fmt.Errorf("unsupported type for LinkStatus: %T", value)
	}
	return nil
}

// Value 实现数据库值接口
func (ls LinkStatus) Value() (driver.Value, error) {
	return string(ls), nil
}

// Link 短链接模型
type Link struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ShortCode   string     `gorm:"size:10;not null;uniqueIndex" json:"short_code"`
	LongURL     string     `gorm:"type:text;not null" json:"long_url"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	ClickCount  int64      `gorm:"default:0" json:"click_count"`
	Status      LinkStatus `gorm:"size:20;default:active" json:"status"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy   string     `gorm:"size:100" json:"created_by,omitempty"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	UpdatedBy   string     `gorm:"size:100" json:"updated_by,omitempty"`
	Description string     `gorm:"size:500" json:"description,omitempty"`
	DeleteFlag  string     `gorm:"size:1" json:"delete_flag,omitempty"`
	Version     uint       `gorm:"default:0" json:"version"`
}

// TableName 指定表名
func (l *Link) TableName() string {
	return "links"
}

// IsActive 检查链接是否有效
func (l *Link) IsActive() bool {
	if l.Status != LinkStatusActive {
		return false
	}

	if l.ExpiresAt != nil && l.ExpiresAt.Before(time.Now()) {
		return false
	}
	return true
}

// ClickStats 点击统计模型
type ClickStats struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ShortCode   string    `gorm:"size:10;not null;index" json:"short_code"`
	IPAddress   string    `gorm:"size:45" json:"ip_address"` // 支持IPv6
	UserAgent   string    `gorm:"type:text" json:"user_agent"`
	Referer     string    `gorm:"type:text" json:"referer"`
	Country     string    `gorm:"size:2" json:"country"` // ISO 3166-1 alpha-2
	Region      string    `gorm:"size:100" json:"region"`
	City        string    `gorm:"size:100" json:"city"`
	DeviceType  string    `gorm:"size:50" json:"device_type"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy   string    `gorm:"size:100" json:"created_by,omitempty"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	UpdatedBy   string    `gorm:"size:100" json:"updated_by,omitempty"`
	Description string    `gorm:"size:500" json:"description,omitempty"`
	DeleteFlag  string    `gorm:"size:1" json:"delete_flag,omitempty"`
	Version     uint      `gorm:"default:0" json:"version"`
}

// TableName 指定表名
func (c *ClickStats) TableName() string {
	return "click_stats"
}

// DailyStats 每日统计摘要
type DailyStats struct {
	Date      string `json:"date"`
	Clicks    int64  `json:"clicks"`
	UniqueIPs int64  `json:"unique_ips"`
}

// StatsSummary 统计摘要
type StatsSummary struct {
	TotalClicks int64            `json:"total_clicks"`
	DailyStats  []DailyStats     `json:"daily_stats,omitempty"`
	Referrers   map[string]int64 `json:"referrers,omitempty"`
	Countries   map[string]int64 `json:"countries,omitempty"`
	Devices     map[string]int64 `json:"devices,omitempty"`
	Last30Days  []DailyStats     `json:"last_30_days,omitempty"`
}
