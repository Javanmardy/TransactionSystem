package cache

import (
	"context"
	"fmt"
	"os"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func withMiniRedis(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	s, err := miniredis.Run()
	if err != nil {
		t.Skipf("could not start miniredis: %v", err)
	}
	os.Setenv("REDIS_ADDR", s.Addr())
	os.Setenv("CACHE_TTL_SECONDS", "120")
	return s
}

func TestSetAndGetTransaction(t *testing.T) {
	s := withMiniRedis(t)
	defer s.Close()

	m := GetManager()

	id := 12345
	data := "test-transaction-data"

	if err := m.SetTransaction(id, data); err != nil {
		t.Fatalf("Failed to set transaction: %v", err)
	}
	result, err := m.GetTransaction(id)
	if err != nil {
		t.Fatalf("Failed to get transaction: %v", err)
	}
	if result != data {
		t.Errorf("Expected %s, got %s", data, result)
	}
}

func TestSetGetAndExpireTransaction(t *testing.T) {
	s := withMiniRedis(t)
	defer s.Close()

	m := GetManager()
	id := 54321
	data := "expire-test"
	if err := m.SetTransaction(id, data); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	got, err := m.GetTransaction(id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != data {
		t.Errorf("Expected %q, got %q", data, got)
	}
	managerKey := fmt.Sprintf("transaction:%d", id)
	redisClient := getInternalRedisClientForTest()
	redisClient.Del(context.Background(), managerKey)
	_, err = m.GetTransaction(id)
	if err == nil {
		t.Errorf("Expected error for missing key, got nil")
	}
}

func getInternalRedisClientForTest() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	return redis.NewClient(&redis.Options{Addr: addr})
}
