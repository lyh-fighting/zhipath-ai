package middleware

import "testing"

func TestRateLimiterAllowAndReject(t *testing.T) {
	rl := &rateLimiter{counters: make(map[string]*counter), limit: 3, window: 60_000_000_000}
	if !rl.allow("k1") {
		t.Error("第1次应允许")
	}
	if !rl.allow("k1") {
		t.Error("第2次应允许")
	}
	if !rl.allow("k1") {
		t.Error("第3次应允许")
	}
	if rl.allow("k1") {
		t.Error("第4次应拒绝（超限）")
	}
}

func TestRateLimiterDifferentKeys(t *testing.T) {
	rl := &rateLimiter{counters: make(map[string]*counter), limit: 2, window: 60_000_000_000}
	rl.allow("k1")
	rl.allow("k1")
	if !rl.allow("k2") {
		t.Error("不同 key 不应受影响")
	}
}
