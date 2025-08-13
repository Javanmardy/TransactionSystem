package cache

import (
	"context"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func resetSingletons() {
	manager = nil
	redisClient = nil
	once = sync.Once{}
	_ = os.Unsetenv("REDIS_ADDR")
	_ = os.Unsetenv("REDIS_PASSWORD")
	_ = os.Unsetenv("REDIS_DB")
	_ = os.Unsetenv("CACHE_TTL_SECONDS")
}

func withMiniRedis(t *testing.T, opts ...string) *miniredis.Miniredis {
	t.Helper()
	s, err := miniredis.Run()
	if err != nil {
		t.Skipf("could not start miniredis: %v", err)
	}
	ttl := "120"
	if len(opts) > 0 && opts[0] != "" {
		ttl = opts[0]
	}

	os.Setenv("REDIS_ADDR", s.Addr())
	os.Setenv("CACHE_TTL_SECONDS", ttl)
	resetManager()

	return s
}

func resetManager() {
	manager = nil
	redisClient = nil
	once = sync.Once{}
}

func testRedisClient() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	return redis.NewClient(&redis.Options{Addr: addr})
}

func TestSetAndGetTransaction(t *testing.T) {
	resetSingletons()
	s := withMiniRedis(t)
	defer s.Close()

	m := GetManager()

	id := 12345
	data := "test-transaction-data"

	if err := m.SetTransaction(id, data); err != nil {
		t.Fatalf("SetTransaction failed: %v", err)
	}
	got, err := m.GetTransaction(id)
	if err != nil {
		t.Fatalf("GetTransaction failed: %v", err)
	}
	if got != data {
		t.Errorf("expected %q, got %q", data, got)
	}

	ttl := s.TTL("transaction:12345")
	if ttl <= 0 {
		t.Errorf("expected positive TTL, got %v", ttl)
	}
}

func TestSetGetAndMissingKey(t *testing.T) {
	resetSingletons()
	s := withMiniRedis(t)
	defer s.Close()

	m := GetManager()
	id := 54321
	data := "expire-test"

	if err := m.SetTransaction(id, data); err != nil {
		t.Fatalf("SetTransaction failed: %v", err)
	}
	got, err := m.GetTransaction(id)
	if err != nil {
		t.Fatalf("GetTransaction failed: %v", err)
	}
	if got != data {
		t.Errorf("expected %q, got %q", data, got)
	}

	// حذف کلید و انتظار خطا
	rc := testRedisClient()
	rc.Del(context.Background(), "transaction:54321")
	if _, err := m.GetTransaction(id); err == nil {
		t.Errorf("expected error for missing key, got nil")
	}
}

func TestPushRecent_BasicTrimAndOrder(t *testing.T) {
	resetSingletons()
	s := withMiniRedis(t)
	defer s.Close()

	m := GetManager()
	userID := 9

	if err := m.PushRecent(userID, `{"id":1}`, 2); err != nil {
		t.Fatalf("PushRecent failed: %v", err)
	}
	if err := m.PushRecent(userID, `{"id":2}`, 2); err != nil {
		t.Fatalf("PushRecent failed: %v", err)
	}
	if err := m.PushRecent(userID, `{"id":3}`, 2); err != nil {
		t.Fatalf("PushRecent failed: %v", err)
	}

	items, err := m.GetRecent(userID, 10)
	if err != nil {
		t.Fatalf("GetRecent failed: %v", err)
	}
	want := []string{`{"id":3}`, `{"id":2}`}
	if !reflect.DeepEqual(items, want) {
		t.Errorf("unexpected recent list.\nwant: %+v\ngot:  %+v", want, items)
	}

	key := "recent:9"
	ttl := s.TTL(key)
	if ttl <= 0 {
		t.Errorf("expected positive TTL for %s, got %v", key, ttl)
	}

	if ttl <= 0 {
		t.Errorf("expected positive TTL for %s, got %v", key, ttl)
	}
}

func TestPushRecent_DefaultMaxAndNegativeInput(t *testing.T) {
	resetSingletons()
	s := withMiniRedis(t)
	defer s.Close()

	m := GetManager()
	userID := 21

	if err := m.PushRecent(userID, `{"id":0}`, 0); err != nil {
		t.Fatalf("PushRecent failed: %v", err)
	}

	for i := 1; i <= 25; i++ {
		if err := m.PushRecent(userID, `{"id":x}`, -5); err != nil {
			t.Fatalf("PushRecent failed at %d: %v", i, err)
		}
	}

	items, err := m.GetRecent(userID, 100)
	if err != nil {
		t.Fatalf("GetRecent failed: %v", err)
	}
	if len(items) != 20 {
		t.Errorf("expected 20 items after trim, got %d", len(items))
	}
}

func TestGetRecent_LimitHandling(t *testing.T) {
	resetSingletons()
	s := withMiniRedis(t)
	defer s.Close()

	m := GetManager()
	userID := 77

	for i := 1; i <= 5; i++ {
		if err := m.PushRecent(userID, `{"id":x}`, 50); err != nil {
			t.Fatalf("PushRecent failed: %v", err)
		}
	}

	items, err := m.GetRecent(userID, 0)
	if err != nil {
		t.Fatalf("GetRecent failed: %v", err)
	}
	if len(items) != 5 {
		t.Errorf("expected 5 items, got %d", len(items))
	}
}

func TestDeleteTransactionKey(t *testing.T) {
	resetSingletons()
	s := withMiniRedis(t)
	defer s.Close()

	m := GetManager()
	id := 999

	if err := m.SetTransaction(id, "xx"); err != nil {
		t.Fatalf("SetTransaction failed: %v", err)
	}

	m.DeleteTransactionKey(id)

	if _, err := m.GetTransaction(id); err == nil {
		t.Errorf("expected error after DeleteTransactionKey, got nil")
	}
}

func TestTTLFromEnvParsing(t *testing.T) {
	resetSingletons()
	os.Setenv("CACHE_TTL_SECONDS", "7")
	if got := ttlFromEnv(); got != 7*time.Second {
		t.Errorf("expected 7s, got %v", got)
	}
	os.Setenv("CACHE_TTL_SECONDS", "not-a-number")
	if got := ttlFromEnv(); got != 10*time.Minute {
		t.Errorf("expected 10m default, got %v", got)
	}
	os.Unsetenv("CACHE_TTL_SECONDS")
	if got := ttlFromEnv(); got != 10*time.Minute {
		t.Errorf("expected 10m default, got %v", got)
	}
}

func TestEnsureClientReconnectsIfPingFails(t *testing.T) {
	resetSingletons()
	os.Setenv("REDIS_ADDR", "127.0.0.1:6389")
	resetManager()

	_ = ensureClient()

	s, err := miniredis.Run()
	if err != nil {
		t.Skipf("could not start miniredis: %v", err)
	}
	defer s.Close()

	os.Setenv("REDIS_ADDR", s.Addr())
	if err := ensureClient(); err != nil {
		t.Fatalf("ensureClient should reconnect successfully, got err: %v", err)
	}

	m := GetManager()
	if err := m.SetTransaction(1, "ok"); err != nil {
		t.Fatalf("SetTransaction after reconnect failed: %v", err)
	}
	val, err := m.GetTransaction(1)
	if err != nil || val != "ok" {
		t.Fatalf("GetTransaction after reconnect failed: got %q, err %v", val, err)
	}
}
