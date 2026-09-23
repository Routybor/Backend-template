package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	StateClosed   = 0
	StateOpen     = 1
	StateHalfOpen = 2
)

type CircuitBreaker struct {
	state       int
	failures    int
	mu          sync.RWMutex
	threshold   int
	timeout     time.Duration
	lastFailure time.Time
}

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

	if cb.failures >= cb.threshold {
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

		c.Next()

		if c.Writer.Status() >= 500 {
			cb.RecordFailure()
		} else {
			cb.RecordSuccess()
		}
	}
}
