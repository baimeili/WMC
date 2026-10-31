// Package redisx 提供 Redis 与内存两种键值存储实现。
// 开发/测试环境允许使用内存实现；Redis 不可用时自动降级并告警。
package redisx

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store 键值存储抽象，用于验证码、限流、缓存与分布式锁。
type Store interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error)
	Incr(ctx context.Context, key string, ttl time.Duration) (int64, error)
	Del(ctx context.Context, keys ...string) error
	Ping(ctx context.Context) error
	Kind() string
	Close() error
}

// RedisStore 基于 go-redis 的实现。
type RedisStore struct{ client *redis.Client }

// NewRedis 依据 URL 创建 Redis 存储。
func NewRedis(rawURL string) (*RedisStore, error) {
	opt, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("解析 REDIS_URL 失败: %w", err)
	}
	return &RedisStore{client: redis.NewClient(opt)}, nil
}

// Kind 实现 Store。
func (s *RedisStore) Kind() string { return "redis" }

// Ping 实现 Store。
func (s *RedisStore) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return s.client.Ping(ctx).Err()
}

// Get 实现 Store。
func (s *RedisStore) Get(ctx context.Context, key string) (string, bool, error) {
	v, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return v, true, nil
}

// Set 实现 Store。
func (s *RedisStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

// SetNX 实现 Store。
func (s *RedisStore) SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	return s.client.SetNX(ctx, key, value, ttl).Result()
}

// Incr 实现 Store：自增并设置过期时间。
func (s *RedisStore) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	pipe := s.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// Del 实现 Store。
func (s *RedisStore) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return s.client.Del(ctx, keys...).Err()
}

// Close 实现 Store。
func (s *RedisStore) Close() error { return s.client.Close() }

// MemoryStore 进程内实现，用于测试与 Redis 不可用时的降级。
type MemoryStore struct {
	mu   sync.Mutex
	data map[string]memoryEntry
	now  func() time.Time
}

type memoryEntry struct {
	value     string
	expiresAt time.Time
}

// NewMemory 创建内存存储。
func NewMemory() *MemoryStore {
	return &MemoryStore{data: make(map[string]memoryEntry), now: time.Now}
}

// Kind 实现 Store。
func (s *MemoryStore) Kind() string { return "memory" }

func (s *MemoryStore) get(key string) (memoryEntry, bool) {
	e, ok := s.data[key]
	if !ok {
		return memoryEntry{}, false
	}
	if !e.expiresAt.IsZero() && s.now().After(e.expiresAt) {
		delete(s.data, key)
		return memoryEntry{}, false
	}
	return e, true
}

// Get 实现 Store。
func (s *MemoryStore) Get(_ context.Context, key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.get(key)
	if !ok {
		return "", false, nil
	}
	return e.value, true, nil
}

// Set 实现 Store。
func (s *MemoryStore) Set(_ context.Context, key, value string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := memoryEntry{value: value}
	if ttl > 0 {
		e.expiresAt = s.now().Add(ttl)
	}
	s.data[key] = e
	return nil
}

// SetNX 实现 Store。
func (s *MemoryStore) SetNX(_ context.Context, key, value string, ttl time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.get(key); ok {
		return false, nil
	}
	e := memoryEntry{value: value}
	if ttl > 0 {
		e.expiresAt = s.now().Add(ttl)
	}
	s.data[key] = e
	return true, nil
}

// Incr 实现 Store。
func (s *MemoryStore) Incr(_ context.Context, key string, ttl time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.get(key)
	var n int64
	if ok {
		fmt.Sscanf(e.value, "%d", &n)
	}
	n++
	entry := memoryEntry{value: fmt.Sprintf("%d", n)}
	if ok && !e.expiresAt.IsZero() {
		entry.expiresAt = e.expiresAt
	} else if ttl > 0 {
		entry.expiresAt = s.now().Add(ttl)
	}
	s.data[key] = entry
	return n, nil
}

// Del 实现 Store。
func (s *MemoryStore) Del(_ context.Context, keys ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, k := range keys {
		delete(s.data, k)
	}
	return nil
}

// Ping 实现 Store。
func (s *MemoryStore) Ping(context.Context) error { return nil }

// Close 实现 Store。
func (s *MemoryStore) Close() error { return nil }

// New 按配置创建存储：mode=redis 必须成功；auto 时失败降级为内存。
func New(ctx context.Context, mode, rawURL string) (Store, error) {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = "auto"
	}
	if mode == "memory" || rawURL == "" {
		if mode == "redis" {
			return nil, errors.New("RATE_LIMIT_STORE=redis 但未配置 REDIS_URL")
		}
		return NewMemory(), nil
	}
	store, err := NewRedis(rawURL)
	if err != nil {
		if mode == "redis" {
			return nil, err
		}
		return NewMemory(), nil
	}
	if err := store.Ping(ctx); err != nil {
		_ = store.Close()
		if mode == "redis" {
			return nil, fmt.Errorf("Redis 不可用: %w", err)
		}
		return NewMemory(), nil
	}
	return store, nil
}
