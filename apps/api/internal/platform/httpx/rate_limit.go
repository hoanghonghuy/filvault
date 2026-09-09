package httpx

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// bucket is a token bucket for one client IP.
type bucket struct {
	tokens   float64
	lastSeen time.Time
}

// IPRateLimiter is an in-process token-bucket limiter keyed by client IP
// (ADR 0004). Sufficient for single-instance deployments.
type IPRateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	perMin   float64 // refill rate, tokens per second
	capacity float64
	lastSweep time.Time
}

// NewIPRateLimiter allows perMinute requests per minute with a burst of
// twice the per-minute allowance.
func NewIPRateLimiter(perMinute int) *IPRateLimiter {
	if perMinute <= 0 {
		perMinute = 30
	}
	capacity := float64(perMinute) * 2
	return &IPRateLimiter{
		buckets:   make(map[string]*bucket),
		perMin:    float64(perMinute) / 60,
		capacity:  capacity,
		lastSweep: time.Now(),
	}
}

// Middleware returns a gin middleware enforcing the limit with 429 replies.
func (l *IPRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.allow(c.ClientIP()) {
			c.AbortWithStatus(429)
			return
		}
		c.Next()
	}
}

func (l *IPRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.buckets[ip]
	if !ok {
		b = &bucket{tokens: l.capacity, lastSeen: now}
		l.buckets[ip] = b
	}
	elapsed := now.Sub(b.lastSeen).Seconds()
	b.lastSeen = now
	b.tokens += elapsed * l.perMin
	if b.tokens > l.capacity {
		b.tokens = l.capacity
	}
	if b.tokens < 1 {
		l.sweep(now)
		return false
	}
	b.tokens--
	l.sweep(now)
	return true
}

// sweep drops idle buckets at most once per minute to bound memory.
func (l *IPRateLimiter) sweep(now time.Time) {
	if now.Sub(l.lastSweep) < time.Minute {
		return
	}
	l.lastSweep = now
	for ip, b := range l.buckets {
		if now.Sub(b.lastSeen) > 10*time.Minute {
			delete(l.buckets, ip)
		}
	}
}
