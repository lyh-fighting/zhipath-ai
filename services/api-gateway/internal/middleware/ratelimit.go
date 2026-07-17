package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// 用户/IP/接口三级限流。
type rateLimiter struct {
	mu       sync.Mutex
	counters map[string]*counter
	limit    int
	window   time.Duration
}

type counter struct {
	count    int
	expireAt time.Time
}

var defaultLimiter = &rateLimiter{
	counters: make(map[string]*counter),
	limit:    60, // 每分钟 60 次
	window:   time.Minute,
}

// RateLimit 用户/IP/接口三级限流。
func RateLimit() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		userID := UserFromContext(ctx)
		ip := c.ClientIP()
		path := string(c.Request.URI().Path())
		key := "u:" + userID + "|ip:" + string(ip) + "|p:" + path
		if !defaultLimiter.allow(key) {
			WriteError(ctx, c, 429, "RATE_LIMITED", "请求过于频繁，请稍后再试", true)
			c.Abort()
			return
		}
		c.Next(ctx)
	}
}

func (r *rateLimiter) allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	c, ok := r.counters[key]
	if !ok || now.After(c.expireAt) {
		r.counters[key] = &counter{count: 1, expireAt: now.Add(r.window)}
		return true
	}
	if c.count >= r.limit {
		return false
	}
	c.count++
	return true
}
