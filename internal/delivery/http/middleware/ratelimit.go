package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type clientBucket struct {
	tokens     float64
	lastRefill time.Time
}

type IPRateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientBucket
	rate     float64 // токенов в секунду
	capacity float64 // максимальный объем
}

func NewIPRateLimiter(rate float64, capacity float64) *IPRateLimiter {
	limiter := &IPRateLimiter{
		clients:  make(map[string]*clientBucket),
		rate:     rate,
		capacity: capacity,
	}

	// Очистка устаревших IP каждые 10 минут
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			limiter.mu.Lock()
			for ip, b := range limiter.clients {
				if time.Since(b.lastRefill) > 15*time.Minute {
					delete(limiter.clients, ip)
				}
			}
			limiter.mu.Unlock()
		}
	}()

	return limiter
}

func (l *IPRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, exists := l.clients[ip]
	if !exists {
		l.clients[ip] = &clientBucket{
			tokens:     l.capacity - 1,
			lastRefill: now,
		}
		return true
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * l.rate
	if b.tokens > l.capacity {
		b.tokens = l.capacity
	}
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}

	return false
}

func GetIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (l *IPRateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := GetIP(r)
		if !l.Allow(ip) {
			http.Error(w, "Слишком много попыток входа. Пожалуйста, подождите минуту перед повторной попыткой.", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}
