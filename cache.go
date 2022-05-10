package cache

import (
	"runtime"
	"sync"
	"time"
)

type Cache[K comparable, V any] struct {
	*cache[K, V]
	*scrubber[K, V]
}

func (c Cache[K, V]) Stop() {
	if c.scrubber != nil {
		c.scrubber.halt <- struct{}{}
	}
}

type cache[K comparable, V any] struct {
	values     map[K]entry[V]
	expiration time.Duration
	lock       sync.RWMutex
}

type entry[V any] struct {
	value  V
	expiry time.Time
}

func New[K comparable, V any](expiration, cleanup time.Duration) (c *Cache[K, V]) {
	rc := &cache[K, V]{
		values:     make(map[K]entry[V]),
		expiration: expiration,
	}
	c = &Cache[K, V]{cache: rc}
	if cleanup > 0 {
		c.scrubber = &scrubber[K, V]{
			period: cleanup,
			halt:   make(chan struct{}),
			cache:  c,
		}
		go c.scrubber.run()
		runtime.SetFinalizer(c, stopScrubber[K, V])
	}
	return
}

func (c *Cache[K, V]) Add(key K, value V) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.values[key] = entry[V]{
		value:  value,
		expiry: time.Now().Add(c.expiration),
	}
}

func (c *Cache[K, V]) Get(key K) (result V, found bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var e entry[V]
	e, found = c.values[key]

	if found == false {
		return
	}

	if time.Now().After(e.expiry) {
		found = false
		return
	}

	result = e.value
	return
}

func (c *Cache[K, V]) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.values)
}

func (c *Cache[K, V]) scrub() {
	c.lock.Lock()
	defer c.lock.Unlock()

	for key, e := range c.values {
		if time.Now().After(e.expiry) {
			delete(c.values, key)
		}
	}
}
