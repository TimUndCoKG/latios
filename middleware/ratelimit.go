package middleware

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// IPRateLimiter holds a map of limiters for each IP address
type IPRateLimiter struct {
	ips                 map[string]*rate.Limiter
	mutex               sync.RWMutex
	requests_per_second rate.Limit
	burst_size          int
}

// NewIPRateLimiter creates a new limiter
func NewIPRateLimiter(requests_per_second rate.Limit, burst_size int) *IPRateLimiter {
	return &IPRateLimiter{
		ips:                 make(map[string]*rate.Limiter),
		requests_per_second: requests_per_second,
		burst_size:          burst_size,
	}
}

func (rateLimiter *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	rateLimiter.mutex.Lock()
	defer rateLimiter.mutex.Unlock()

	limiter, exists := rateLimiter.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(rateLimiter.requests_per_second, rateLimiter.burst_size)
		rateLimiter.ips[ip] = limiter
	}

	return limiter
}

// RateLimitMiddleware wraps a handler to enforce the limit
func (rateLimiter *IPRateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Get IP from header if proxied by cloudflare or nginx
		clientIP := r.Header.Get("X-Forwarded-For")
		if clientIP == "" {
			var err error
			clientIP, _, err = net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				clientIP = r.RemoteAddr
			}
		}

		limiter := rateLimiter.getLimiter(clientIP)
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
