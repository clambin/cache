package cache

import (
	"fmt"
	"time"
)

type scrubber[K comparable, V any] struct {
	period time.Duration
	halt   chan struct{}
	cache  *Cache[K, V]
}

func (s *scrubber[K, V]) run() {
	ticker := time.NewTicker(s.period)
	for running := true; running; {
		select {
		case <-s.halt:
			fmt.Println("stopping the scrubber")
			running = false
		case <-ticker.C:
			s.cache.scrub()
		}
	}
	ticker.Stop()
}

func stopScrubber[K comparable, V any](c *Cache[K, V]) {
	c.scrubber.halt <- struct{}{}
}
