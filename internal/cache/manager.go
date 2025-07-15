package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	manager     *CacheManager
	once        sync.Once
	redisClient *redis.Client
)

type CacheManager struct{}

func GetManager() *CacheManager {
	once.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})
		manager = &CacheManager{}
	})
	return manager
}

func (c *CacheManager) SetTransaction(id int, data string) error {
	ctx := context.Background()
	key := fmt.Sprintf("transaction:%d", id)
	return redisClient.Set(ctx, key, data, 10*time.Minute).Err()
}

func (c *CacheManager) GetTransaction(id int) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("transaction:%d", id)
	return redisClient.Get(ctx, key).Result()
}
