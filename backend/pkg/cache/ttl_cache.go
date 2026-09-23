package cache

import (
	"sync"
	"time"
)

type entry[V any] struct {
	value   V
	expires time.Time
}

type TTLCache[K comparable, V any] struct {
	mu      sync.RWMutex
	items   map[K]entry[V]
	ttl     time.Duration
	stop    chan struct{}
	stopped sync.Once
}

func NewTTLCache[K comparable, V any](ttl, cleanupInterval time.Duration) *TTLCache[K, V] {
	c := &TTLCache[K, V]{
		items: make(map[K]entry[V]),
		ttl:   ttl,
		stop:  make(chan struct{}),
	}
	go c.cleanup(cleanupInterval)
	return c
}

func (c *TTLCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	e, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expires) {
		var zero V
		return zero, false
	}
	return e.value, true
}

func (c *TTLCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	c.items[key] = entry[V]{value: value, expires: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

func (c *TTLCache[K, V]) Delete(key K) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

func (c *TTLCache[K, V]) Stop() {
	c.stopped.Do(func() { close(c.stop) })
}

func (c *TTLCache[K, V]) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for k, e := range c.items {
				if now.After(e.expires) {
					delete(c.items, k)
				}
			}
			c.mu.Unlock()
		case <-c.stop:
			return
		}
	}
}
