package limiter

import (
	"sync"
	"time"
)

// TokenBucket 令牌桶限流器
type TokenBucket struct {
	rate     float64 // 每秒补充令牌数
	capacity float64 // 桶容量
	tokens   float64
	lastTime time.Time
	mu       sync.Mutex
}

func NewTokenBucket(rate, capacity float64) *TokenBucket {
	return &TokenBucket{
		rate:     rate,
		capacity: capacity,
		tokens:   capacity,
		lastTime: time.Now(),
	}
}

// Allow 尝试消耗一个令牌，返回是否允许
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastTime).Seconds()
	tb.tokens = min(tb.capacity, tb.tokens+elapsed*tb.rate)
	tb.lastTime = now

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// UserLimiterManager 管理每个用户的限流器
type UserLimiterManager struct {
	mu       sync.RWMutex
	limiters map[int64]*TokenBucket
	rate     float64
	capacity float64
}

func NewUserLimiterManager(rate, capacity float64) *UserLimiterManager {
	return &UserLimiterManager{
		limiters: make(map[int64]*TokenBucket),
		rate:     rate,
		capacity: capacity,
	}
}

func (m *UserLimiterManager) Allow(uid int64) bool {
	m.mu.RLock()
	lb, ok := m.limiters[uid]
	m.mu.RUnlock()

	if !ok {
		m.mu.Lock()
		lb = NewTokenBucket(m.rate, m.capacity)
		m.limiters[uid] = lb
		m.mu.Unlock()
	}
	return lb.Allow()
}
