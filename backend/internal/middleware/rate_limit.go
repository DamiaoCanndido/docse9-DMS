package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type ipBucket struct {
	tokens     int
	lastRefill time.Time
}

// IPRateLimiter gerencia baldes de tokens em memória indexados por endereço IP.
type IPRateLimiter struct {
	mu          sync.Mutex
	buckets     map[string]*ipBucket
	rate        int
	window      time.Duration
	stopCleanup chan struct{}
}

// NewIPRateLimiter instancia um novo limitador de taxa por IP.
func NewIPRateLimiter(rate int, window time.Duration) *IPRateLimiter {
	limiter := &IPRateLimiter{
		buckets:     make(map[string]*ipBucket),
		rate:        rate,
		window:      window,
		stopCleanup: make(chan struct{}),
	}

	go limiter.cleanupLoop()

	return limiter
}

// Allow verifica se uma requisição originada do IP pode prosseguir.
func (l *IPRateLimiter) Allow(ip string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, exists := l.buckets[ip]
	if !exists {
		l.buckets[ip] = &ipBucket{
			tokens:     l.rate - 1,
			lastRefill: now,
		}
		return true, 0
	}

	// Se a janela de tempo já passou, reseta os tokens
	if now.Sub(b.lastRefill) >= l.window {
		b.tokens = l.rate - 1
		b.lastRefill = now
		return true, 0
	}

	if b.tokens > 0 {
		b.tokens--
		return true, 0
	}

	retryAfter := l.window - now.Sub(b.lastRefill)
	return false, retryAfter
}

func (l *IPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.mu.Lock()
			now := time.Now()
			for ip, b := range l.buckets {
				if now.Sub(b.lastRefill) > l.window*2 {
					delete(l.buckets, ip)
				}
			}
			l.mu.Unlock()
		case <-l.stopCleanup:
			return
		}
	}
}

// Stop finaliza a goroutine de limpeza.
func (l *IPRateLimiter) Stop() {
	close(l.stopCleanup)
}

// RateLimiterMiddleware retorna um middleware Gin que limita requisições por IP.
func RateLimiterMiddleware(rate int, window time.Duration) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rate, window)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		allowed, retryAfter := limiter.Allow(ip)
		if !allowed {
			retrySeconds := int(retryAfter.Seconds()) + 1
			c.Header("Retry-After", fmt.Sprintf("%d", retrySeconds))
			response.Error(c, http.StatusTooManyRequests, "muitas requisições. Por favor, aguarde antes de tentar novamente.")
			c.Abort()
			return
		}
		c.Next()
	}
}
