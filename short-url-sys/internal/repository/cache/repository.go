package cache

import (
	"context"
	"fmt"
	"short-url-sys/internal/config"
	"short-url-sys/internal/pkg/errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Repository struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

func NewRepository(client *redis.Client, cfg *config.CacheConfig) *Repository {
	return &Repository{
		client: client,
		prefix: cfg.Prefix,
		ttl:    time.Duration(cfg.TTL) * time.Second,
	}
}

func (r *Repository) getKey(typ string, id string) string {
	return fmt.Sprintf("%s:%s:%s", r.prefix, typ, id)
}

// GetShortURL 从缓存获取短链映射
func (r *Repository) GetShortURL(ctx context.Context, shortCode string) (string, error) {
	key := r.getKey("url", shortCode)
	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", errors.ErrLinkNotFound
		}
		return "", &errors.RepositoryError{Operation: "GetShortURL", Err: err}
	}
	return result, nil
}

// SetShortURL 设置短链映射到缓存
func (r *Repository) SetShortURL(ctx context.Context, shortCode string, longURL string) error {
	key := r.getKey("url", shortCode)
	err := r.client.Set(ctx, key, longURL, r.ttl).Err()
	if err != nil {
		return &errors.RepositoryError{Operation: "SetShortURL", Err: err}
	}
	return nil
}

// DeleteShortURL 删除缓存中的短链接
func (r *Repository) DeleteShortURL(ctx context.Context, shortCode string) error {
	key := r.getKey("url", shortCode)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return &errors.RepositoryError{Operation: "DeleteShortURL", Err: err}
	}
	return nil
}

// IncrementClickCount 增加点击计数
func (r *Repository) IncrementClickCount(ctx context.Context, shortCode string) (int64, error) {
	key := r.getKey("clicks", shortCode)
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, &errors.RepositoryError{Operation: "IncrementClickCount", Err: err}
	}

	// 设置过期时间（如果key是新创建的）
	if count == 1 {
		r.client.Expire(ctx, key, 24*time.Hour)
	}
	return count, nil
}

func (r *Repository) GetClickCount(ctx context.Context, shortCode string) (int64, error) {
	key := r.getKey("clicks", shortCode)
	count, err := r.client.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, &errors.RepositoryError{Operation: "GetClickCount", Err: err}
	}
	return count, nil
}
