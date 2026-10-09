// Package cache provides a small in-process TTL cache, used to keep repeated
// searches from re-fetching the same per-place enrichment.
package cache

import (
	"sync"
	"time"
)

type entry[V any] struct {
	value   V
	expires time.Time
}

// TTL is a concurrency-safe map with per-entry expiry and a size cap.
// It is deliberately simple: enrichment data is cheap to re-fetch if evicted.
type TTL[V any] struct {
	mu      sync.Mutex
	ttl     time.Duration
	max     int
	entries map[string]entry[V]
	hits    uint64
	misses  uint64
}

// New returns a cache holding at most max entries for ttl each.
func New[V any](ttl time.Duration, max int) *TTL[V] {
	if max <= 0 {
		max = 10000
	}
	return &TTL[V]{ttl: ttl, max: max, entries: make(map[string]entry[V])}
}

// Get returns the cached value if it is present and unexpired.
func (c *TTL[V]) Get(key string) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.entries[key]
	if !ok {
		c.misses++
		var zero V
		return zero, false
	}
	if time.Now().After(e.expires) {
		delete(c.entries, key)
		c.misses++
		var zero V
		return zero, false
	}
	c.hits++
	return e.value, true
}

// Set stores a value, evicting expired entries first if the cache is full.
func (c *TTL[V]) Set(key string, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= c.max {
		c.evictLocked()
	}
	c.entries[key] = entry[V]{value: value, expires: time.Now().Add(c.ttl)}
}

// evictLocked drops expired entries, then falls back to dropping arbitrary
// ones if everything is still live. Callers must hold the lock.
func (c *TTL[V]) evictLocked() {
	now := time.Now()
	for k, e := range c.entries {
		if now.After(e.expires) {
			delete(c.entries, k)
		}
	}
	// Still full: shed a tenth of the entries so this is not run per write.
	if len(c.entries) >= c.max {
		drop := c.max / 10
		if drop == 0 {
			drop = 1
		}
		for k := range c.entries {
			delete(c.entries, k)
			drop--
			if drop <= 0 {
				break
			}
		}
	}
}

// Stats reports cumulative hits and misses, for logging.
func (c *TTL[V]) Stats() (hits, misses uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.hits, c.misses
}
