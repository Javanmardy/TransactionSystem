package cache

import (
	"context"
	"fmt"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestSetAndGetTransaction(t *testing.T) {
	manager := GetManager()

	id := 12345
	data := "test-transaction-data"

	err := manager.SetTransaction(id, data)
	if err != nil {
		t.Fatalf("Failed to set transaction: %v", err)
	}

	result, err := manager.GetTransaction(id)
	if err != nil {
		t.Fatalf("Failed to get transaction: %v", err)
	}

	if result != data {
		t.Errorf("Expected %s, got %s", data, result)
	}
}

func TestSetGetAndExpireTransaction(t *testing.T) {
	manager := GetManager()
	id := 54321
	data := "expire-test"
	if err := manager.SetTransaction(id, data); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	got, err := manager.GetTransaction(id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != data {
		t.Errorf("Expected %q, got %q", data, got)
	}
	managerKey := fmt.Sprintf("transaction:%d", id)
	redisClient := getInternalRedisClientForTest()
	redisClient.Del(context.Background(), managerKey)
	_, err = manager.GetTransaction(id)
	if err == nil {
		t.Errorf("Expected error for missing key, got nil")
	}
}

func getInternalRedisClientForTest() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379", Password: "", DB: 0,
	})
}
