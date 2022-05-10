package cache_test

import (
	"github.com/clambin/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"runtime"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	c := cache.New[string, string](time.Hour, 0)
	require.NotNil(t, c)
	assert.Zero(t, c.Len())

	value, found := c.Get("foo")
	require.False(t, found)

	c.Add("foo", "bar")
	value, found = c.Get("foo")
	require.True(t, found)
	assert.Equal(t, "bar", value)

	c.Add("foo", "foo")
	value, found = c.Get("foo")
	require.True(t, found)
	assert.Equal(t, "foo", value)
}

func TestCacheExpiry(t *testing.T) {
	c := cache.New[string, string](100*time.Millisecond, 0)
	require.NotNil(t, c)

	c.Add("foo", "bar")
	value, found := c.Get("foo")
	require.True(t, found)
	assert.Equal(t, "bar", value)

	assert.Eventually(t, func() bool {
		_, found = c.Get("foo")
		return found == false
	}, 200*time.Millisecond, 50*time.Millisecond)

}

func TestCacheScrubber(t *testing.T) {
	c := cache.New[string, string](100*time.Millisecond, 150*time.Millisecond)
	require.NotNil(t, c)

	c.Add("foo", "bar")
	value, found := c.Get("foo")
	require.True(t, found)
	assert.Equal(t, "bar", value)

	assert.Eventually(t, func() bool {
		_, found = c.Get("foo")
		return found == false
	}, 200*time.Millisecond, 50*time.Millisecond)

	assert.Eventually(t, func() bool {
		return c.Len() == 0
	}, 200*time.Millisecond, 50*time.Millisecond)

	c = nil
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
}
