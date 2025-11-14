package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient обертка для Redis клиента
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
	ttl    time.Duration
}

// RedisConfig конфигурация Redis
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	TTL      time.Duration
}

// NewRedisClient создает новый Redis клиент
func NewRedisClient(cfg *RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()

	// Проверка подключения
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("✅ Connected to Redis at %s:%d", cfg.Host, cfg.Port)

	return &RedisClient{
		client: client,
		ctx:    ctx,
		ttl:    cfg.TTL,
	}, nil
}

// Close закрывает соединение
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Set сохраняет значение с TTL
func (r *RedisClient) Set(key string, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = r.ttl
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.Set(r.ctx, key, data, ttl).Err()
}

// Get получает значение из кэша
func (r *RedisClient) Get(key string, dest interface{}) error {
	data, err := r.client.Get(r.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get value: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete удаляет ключ
func (r *RedisClient) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// Exists проверяет существование ключа
func (r *RedisClient) Exists(key string) (bool, error) {
	result, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Increment увеличивает счетчик
func (r *RedisClient) Increment(key string) (int64, error) {
	return r.client.Incr(r.ctx, key).Result()
}

// Expire устанавливает TTL для ключа
func (r *RedisClient) Expire(key string, ttl time.Duration) error {
	return r.client.Expire(r.ctx, key, ttl).Err()
}

// GetTTL возвращает оставшееся время жизни ключа
func (r *RedisClient) GetTTL(key string) (time.Duration, error) {
	return r.client.TTL(r.ctx, key).Result()
}

// InvalidatePattern удаляет все ключи по паттерну
func (r *RedisClient) InvalidatePattern(pattern string) error {
	iter := r.client.Scan(r.ctx, 0, pattern, 0).Iterator()
	for iter.Next(r.ctx) {
		if err := r.client.Del(r.ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// SetNX устанавливает значение только если ключа не существует
func (r *RedisClient) SetNX(key string, value interface{}, ttl time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.SetNX(r.ctx, key, data, ttl).Result()
}

// ErrCacheMiss ошибка если ключ не найден
var ErrCacheMiss = fmt.Errorf("cache miss")