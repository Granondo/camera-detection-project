package cache

import (
	"fmt"
	"time"
)

// CacheService сервис для работы с кэшем
type CacheService struct {
	redis *RedisClient
}

// NewCacheService создает новый сервис кэша
func NewCacheService(redis *RedisClient) *CacheService {
	return &CacheService{
		redis: redis,
	}
}

// Cache Keys
const (
	KeySystemStatus    = "system:status"
	KeySystemStats     = "system:stats"
	KeyCameraInfo      = "camera:info:%d"
	KeyRecordings      = "recordings:list:%d"
	KeyRecording       = "recording:%d"
	KeyFrames          = "frames:list:%d:%t"
	KeyFrame           = "frame:%d"
	KeyEvents          = "events:list:%d"
	KeyRateLimit       = "rate:limit:%s"
)

// GetSystemStatus получает статус системы из кэша
func (c *CacheService) GetSystemStatus() (interface{}, error) {
	var data interface{}
	err := c.redis.Get(KeySystemStatus, &data)
	return data, err
}

// SetSystemStatus кэширует статус системы
func (c *CacheService) SetSystemStatus(data interface{}) error {
	return c.redis.Set(KeySystemStatus, data, 30*time.Second)
}

// InvalidateSystemStatus сбрасывает кэш статуса
func (c *CacheService) InvalidateSystemStatus() error {
	return c.redis.Delete(KeySystemStatus)
}

// GetStats получает статистику из кэша
func (c *CacheService) GetStats() (interface{}, error) {
	var data interface{}
	err := c.redis.Get(KeySystemStats, &data)
	return data, err
}

// SetStats кэширует статистику
func (c *CacheService) SetStats(data interface{}) error {
	return c.redis.Set(KeySystemStats, data, 60*time.Second)
}

// InvalidateStats сбрасывает кэш статистики
func (c *CacheService) InvalidateStats() error {
	return c.redis.Delete(KeySystemStats)
}

// CheckRateLimit проверяет лимит запросов
func (c *CacheService) CheckRateLimit(identifier string, limit int64, window time.Duration) (bool, error) {
	key := fmt.Sprintf(KeyRateLimit, identifier)
	
	count, err := c.redis.Increment(key)
	if err != nil {
		return false, err
	}

	// Установить TTL при первом запросе
	if count == 1 {
		if err := c.redis.Expire(key, window); err != nil {
			return false, err
		}
	}

	return count <= limit, nil
}

// GetCachedRecordings получает список записей из кэша
func (c *CacheService) GetCachedRecordings(cameraID int) (interface{}, error) {
	var data interface{}
	key := fmt.Sprintf(KeyRecordings, cameraID)
	err := c.redis.Get(key, &data)
	return data, err
}

// SetCachedRecordings кэширует список записей
func (c *CacheService) SetCachedRecordings(cameraID int, data interface{}) error {
	key := fmt.Sprintf(KeyRecordings, cameraID)
	return c.redis.Set(key, data, 5*time.Minute)
}

// InvalidateRecordings сбрасывает кэш записей
func (c *CacheService) InvalidateRecordings() error {
	return c.redis.InvalidatePattern("recordings:*")
}

// InvalidateAll сбрасывает весь кэш
func (c *CacheService) InvalidateAll() error {
	patterns := []string{
		"system:*",
		"camera:*",
		"recordings:*",
		"frames:*",
		"events:*",
	}

	for _, pattern := range patterns {
		if err := c.redis.InvalidatePattern(pattern); err != nil {
			return err
		}
	}

	return nil
}