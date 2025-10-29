package redirect

import (
	"context"
	"log"
	"short-url-sys/internal/model"
	"short-url-sys/internal/pkg/errors"
	"short-url-sys/internal/repository/cache"
	linkRepo "short-url-sys/internal/repository/link"
	statsRepo "short-url-sys/internal/repository/stats"
	"strings"
	"time"
)

// Service 重定向接口
type Service interface {
	Redirect(ctx context.Context, shortCode string, req *RedirectRequest) (string, error)
	RecordClick(ctx context.Context, shortCode string, req *RedirectRequest) error
}

// RedirectRequest 重定向请求
type RedirectRequest struct {
	IPAddress string
	UserAgent string
	Referer   string
	Country   string
	Region    string
	City      string
}

type redirectService struct {
	linkRepo  linkRepo.Repository
	statsRepo statsRepo.Repository
	cacheRepo *cache.Repository
}

// Redirect 执行重定向
func (s *redirectService) Redirect(ctx context.Context, shortCode string, req *RedirectRequest) (string, error) {
	// 从缓存或数据库获取长链接
	longURL, err := s.getLongURL(ctx, shortCode)
	if err != nil {
		return "", err
	}

	// 异步记录点击统计
	go s.RecordClick(context.Background(), shortCode, req)
	return longURL, nil
}

// RecordClick 记录点击统计
func (s *redirectService) RecordClick(ctx context.Context, shortCode string, req *RedirectRequest) error {
	return s.recordClick(ctx, shortCode, req)
}

// NewRedirectRequest 创建重定向服务实例
func NewRedirectRequest(
	linkRepo linkRepo.Repository,
	statsRepo statsRepo.Repository,
	cacheRepo *cache.Repository,
) Service {
	return &redirectService{
		linkRepo:  linkRepo,
		statsRepo: statsRepo,
		cacheRepo: cacheRepo,
	}
}

func (s *redirectService) getLongURL(ctx context.Context, shortCode string) (string, error) {
	// 首先尝试从缓存中获取
	longURL, err := s.cacheRepo.GetShortURL(ctx, shortCode)
	if err == nil {
		return longURL, nil
	}

	// 混存未命中，从数据库获取
	link, err := s.linkRepo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	// 检查链接状态
	if !link.IsActive() {
		if link.Status == model.LinkStatusExpired {
			return "", errors.ErrLinkExpired
		}
	}

	// 更新缓存
	go func() {
		ctx := context.Background()
		err = s.cacheRepo.SetShortURL(ctx, shortCode, longURL)
		if err != nil {
			log.Println("An error: %w occurred while set short url", err)
		}
	}()

	return longURL, nil
}

func (s *redirectService) recordClick(ctx context.Context, shortCode string, req *RedirectRequest) error {
	stats := &model.ClickStats{
		ShortCode:  shortCode,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
		Referer:    req.Referer,
		Country:    req.Country,
		Region:     req.Region,
		DeviceType: s.getDeviceType(req.UserAgent),
		CreatedAt:  time.Now(),
	}

	// 记录点击
	if err := s.statsRepo.RecordClick(ctx, stats); err != nil {
		// 记录错误，但不影响重定向
		log.Printf("An error: %w occurred while recording short url\n", err)
		return err
	}

	// 更新点击计数
	if err := s.linkRepo.UpdateClickCount(ctx, shortCode, 1); err != nil {
		// 记录错误，但不影响重定向
		log.Printf("An error: %w occurred while recording short url\n", err)
		return err
	}

	// 更新缓存中的点击计数
	_, err := s.cacheRepo.IncrementClickCount(ctx, shortCode)
	if err != nil {
		log.Printf("An error: %w occurred while recording short url\n", err)
		return err
	}
	return nil
}

func (s *redirectService) getDeviceType(userAgent string) string {
	ua := strings.ToLower(userAgent)

	switch {
	case strings.Contains(ua, "mobile"):
		return "mobile"
	case strings.Contains(ua, "tablet"):
		return "tablet"
	case strings.Contains(ua, "bot"):
		return "bot"
	case strings.Contains(ua, "curl"):
		return "curl"
	case strings.Contains(ua, "wget"):
		return "wget"
	default:
		return "desktop"
	}
}
