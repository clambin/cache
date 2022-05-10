package cache

import (
	"runtime"
	"sync"
	"time"
)

// Cacher implements a cache interface
type Cacher[K comparable, V any] interface {
	Add(key K, value V)
	Get(key K) (value V, found bool)
}

// Cache implements the Cacher interface
type Cache[K comparable, V any] struct {
	*realCache[K, V]
	*scrubber[K, V]
}

var _ Cacher[int, string] = &Cache[int, string]{}

type realCache[K comparable, V any] struct {
	values     map[K]entry[V]
	expiration time.Duration
	lock       sync.RWMutex
}

type entry[V any] struct {
	value  V
	expiry time.Time
}

func (e entry[K]) isExpired() bool {
	return !e.expiry.IsZero() && time.Now().After(e.expiry)
}

// New creates a new Cache for the specified key and value types.  expiration specifies how long an entry can live in the cache
// before expiring.  cleanup specifies how often the cache should remove expired items from the cache.
func New[K comparable, V any](expiration, cleanup time.Duration) (c *Cache[K, V]) {
	c = &Cache[K, V]{
		realCache: &realCache[K, V]{
			values:     make(map[K]entry[V]),
			expiration: expiration,
		},
	}
	if cleanup > 0 {
		c.scrubber = &scrubber[K, V]{
			period: cleanup,
			halt:   make(chan struct{}),
			cache:  c.realCache,
		}
		go c.scrubber.run()
		runtime.SetFinalizer(c, stopScrubber[K, V])
	}
	return
}

// Add adds a value/value pair to the cache
func (c *Cache[K, V]) Add(key K, value V) {
	c.lock.Lock()
	defer c.lock.Unlock()

	expiry := time.Time{}
	if c.expiration != 0 {
		expiry = time.Now().Add(c.expiration)
	}

	c.values[key] = entry[V]{
		value:  value,
		expiry: expiry,
	}
}

// Get retrieves the value from the cache. If the item is not found, or expired, found will be false
func (c *Cache[K, V]) Get(key K) (result V, found bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var e entry[V]
	e, found = c.values[key]

	if !found || e.isExpired() {
		return result, false
	}

	return e.value, true
}

// Size returns the current size of the cache. Expired items are counted
func (c *Cache[K, V]) Size() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.values)
}

// Len returns the number of non-expired items in the case
func (c *Cache[K, V]) Len() (count int) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	for _, e := range c.values {
		if !e.isExpired() {
			count++
		}
	}
	return count
}

func (c *realCache[K, V]) scrub() {
	c.lock.Lock()
	defer c.lock.Unlock()

	for key, e := range c.values {
		if time.Now().After(e.expiry) {
			delete(c.values, key)
		}
	}
}
