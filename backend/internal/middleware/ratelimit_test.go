package middleware_test

import (
	"testing"
	"time"

	"negative-ion-respirator/backend/internal/middleware"
)

func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	rl := middleware.NewRateLimiter(100, 200) // 100 req/s, burst 200

	key := "test-ip"
	for i := 0; i < 200; i++ {
		if !rl.Allow(key) {
			t.Errorf("request %d should have been allowed", i+1)
		}
	}

	// 201st request should be denied
	if rl.Allow(key) {
		t.Error("request 201 should have been denied")
	}
}

func TestRateLimiter_DifferentKeys(t *testing.T) {
	rl := middleware.NewRateLimiter(1, 1)

	if !rl.Allow("ip-1") {
		t.Error("ip-1 first request should be allowed")
	}
	if !rl.Allow("ip-2") {
		t.Error("ip-2 should have its own bucket")
	}
	if rl.Allow("ip-1") {
		t.Error("ip-1 second request should be denied")
	}
}

func TestRateLimiter_RefillOverTime(t *testing.T) {
	// 1 token per second, capacity of 2
	rl := middleware.NewRateLimiter(1, 2)

	key := "refill-test"

	// Use all tokens immediately (2 burst)
	if !rl.Allow(key) {
		t.Error("first request should be allowed")
	}
	if !rl.Allow(key) {
		t.Error("second request should be allowed")
	}
	if rl.Allow(key) {
		t.Error("third request should be denied (bucket empty)")
	}

	// Wait for token refill
	time.Sleep(1100 * time.Millisecond)

	if !rl.Allow(key) {
		t.Error("request after refill should be allowed")
	}
}

func TestRateLimiter_ManyKeys(t *testing.T) {
	rl := middleware.NewRateLimiter(10, 10)

	// Each key should get its own bucket
	for i := 0; i < 100; i++ {
		key := "ip-" + string(rune('0'+i%10))
		if !rl.Allow(key) {
			t.Errorf("key %s first request should be allowed", key)
		}
	}
}

func TestRateLimiter_RateLimitHandler(t *testing.T) {
	// Verify the handler factory returns a valid function
	handler := middleware.RateLimit(10, 20)
	if handler == nil {
		t.Error("RateLimit should return a non-nil handler function")
	}
}
