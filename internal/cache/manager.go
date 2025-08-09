package cache

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	manager     *CacheManager
	once        sync.Once
	redisClient *redis.Client
	cacheTTL    = ttlFromEnv()
)

func initClientFromEnv() *redis.Client {
	addr := getenv("REDIS_ADDR", "localhost:6379")
	pass := os.Getenv("REDIS_PASSWORD")
	db := atoi(getenv("REDIS_DB", "0"))
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       db,
	})
}

func ensureClient() error {
	ctx := context.Background()
	if redisClient == nil {
		redisClient = initClientFromEnv()
	}
	if err := redisClient.Ping(ctx).Err(); err != nil {
		redisClient = initClientFromEnv()
		return redisClient.Ping(ctx).Err()
	}
	return nil
}

func GetManager() *CacheManager {
	once.Do(func() {
		redisClient = initClientFromEnv()
		manager = &CacheManager{}
	})
	return manager
}

type CacheManager struct{}

func ttlFromEnv() time.Duration {
	if v := os.Getenv("CACHE_TTL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 10 * time.Minute
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func (c *CacheManager) SetTransaction(id int, data string) error {
	if err := ensureClient(); err != nil {
		return err
	}
	ctx := context.Background()
	key := fmt.Sprintf("transaction:%d", id)
	return redisClient.Set(ctx, key, data, cacheTTL).Err()
}

func (c *CacheManager) GetTransaction(id int) (string, error) {
	if err := ensureClient(); err != nil {
		return "", err
	}
	ctx := context.Background()
	key := fmt.Sprintf("transaction:%d", id)
	return redisClient.Get(ctx, key).Result()
}

func (c *CacheManager) PushRecent(userID int, jsonData string, max int) error {
	if err := ensureClient(); err != nil {
		return err
	}
	if max <= 0 {
		max = 20
	}
	ctx := context.Background()
	key := fmt.Sprintf("recent:%d", userID)
	pipe := redisClient.TxPipeline()
	pipe.LPush(ctx, key, jsonData)
	pipe.LTrim(ctx, key, 0, int64(max-1))
	pipe.Expire(ctx, key, cacheTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *CacheManager) GetRecent(userID, limit int) ([]string, error) {
	if err := ensureClient(); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 20
	}
	ctx := context.Background()
	key := fmt.Sprintf("recent:%d", userID)
	return redisClient.LRange(ctx, key, 0, int64(limit-1)).Result()
}

func (c *CacheManager) DeleteTransactionKey(id int) {
	if ensureClient() != nil {
		return
	}
	_ = redisClient.Del(context.Background(), fmt.Sprintf("transaction:%d", id)).Err()
}
