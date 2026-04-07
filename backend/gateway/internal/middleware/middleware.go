package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type GzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *GzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !shouldCompress(c.Request) {
			c.Next()
			return
		}

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
		if err != nil {
			c.Next()
			return
		}

		c.Writer = &GzipWriter{
			ResponseWriter: c.Writer,
			writer:         gz,
		}

		c.Next()

		gz.Close()
	}
}

func shouldCompress(r *http.Request) bool {
	encoding := r.Header.Get("Accept-Encoding")
	return len(encoding) > 0 && contains(encoding, "gzip")
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, times := range rl.requests {
			var valid []time.Time
			for _, t := range times {
				if now.Sub(t) < rl.window {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = valid
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	times := rl.requests[ip]

	var valid []time.Time
	for _, t := range times {
		if now.Sub(t) < rl.window {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[ip] = valid
		return false
	}

	valid = append(valid, now)
	rl.requests[ip] = valid
	return true
}

func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.Allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}

type CircuitBreaker struct {
	state       int32
	failures    int32
	mu          sync.RWMutex
	threshold   int
	timeout     time.Duration
	lastFailure time.Time
}

const (
	StateClosed   = 0
	StateOpen     = 1
	StateHalfOpen = 2
)

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:     StateClosed,
		threshold: threshold,
		timeout:   timeout,
	}
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	state := cb.state
	lastFailure := cb.lastFailure
	cb.mu.RUnlock()

	if state == StateOpen {
		if time.Since(lastFailure) > cb.timeout {
			cb.mu.Lock()
			if cb.state == StateOpen {
				cb.state = StateHalfOpen
			}
			cb.mu.Unlock()
			return true
		}
		return false
	}

	return true
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailure = time.Now()
	cb.failures++

	if cb.failures >= int32(cb.threshold) {
		cb.state = StateOpen
	}
}

func CircuitBreakerMiddleware(threshold int, timeout time.Duration) gin.HandlerFunc {
	cb := NewCircuitBreaker(threshold, timeout)

	return func(c *gin.Context) {
		if !cb.Allow() {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "service temporarily unavailable",
			})
			return
		}

		bl := &bodyLogger{
			ResponseWriter: c.Writer,
			body:           []byte{},
			breaker:        cb,
		}
		c.Writer = bl

		c.Next()

		if c.Writer.Status() >= 500 {
			cb.RecordFailure()
		} else {
			cb.RecordSuccess()
		}
	}
}

type bodyLogger struct {
	gin.ResponseWriter
	body    []byte
	breaker *CircuitBreaker
}

func (bl *bodyLogger) Write(b []byte) (int, error) {
	bl.body = append(bl.body, b...)
	return bl.ResponseWriter.Write(b)
}

type circuitBreakerRoundTripper struct {
	breaker *CircuitBreaker
	inner   http.RoundTripper
}

func NewCircuitBreakerRoundTripper(cb *CircuitBreaker) http.RoundTripper {
	return &circuitBreakerRoundTripper{
		breaker: cb,
		inner:   http.DefaultTransport,
	}
}

func (cbrt *circuitBreakerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if !cbrt.breaker.Allow() {
		return nil, io.EOF
	}

	resp, err := cbrt.inner.RoundTrip(req)
	if err != nil {
		cbrt.breaker.RecordFailure()
		return resp, err
	}

	if resp.StatusCode >= 500 {
		cbrt.breaker.RecordFailure()
	} else {
		cbrt.breaker.RecordSuccess()
	}

	return resp, nil
}
