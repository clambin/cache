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
	assert.Zero(t, c.Size())

	_, found := c.Get("foo")
	require.False(t, found)

	c.Add("foo", "bar")
	var value string
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
	assert.Equal(t, 1, c.Len())
	assert.Equal(t, 1, c.Size())

	assert.Eventually(t, func() bool {
		_, found = c.Get("foo")
		return found == false
	}, 200*time.Millisecond, 50*time.Millisecond)

	assert.Equal(t, 0, c.Len())
	assert.Equal(t, 1, c.Size())

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
		return c.Size() == 0
	}, 200*time.Millisecond, 50*time.Millisecond)

	c = nil
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
}

func TestCache_Types(t *testing.T) {
	c1 := cache.New[int, string](0, 0)
	c1.Add(1, "hello")
	v1, found := c1.Get(1)
	assert.True(t, found)
	assert.Equal(t, "hello", v1)

	c2 := cache.New[string, string](0, 0)
	c2.Add("foo", "bar")
	v2, found2 := c2.Get("foo")
	assert.True(t, found2)
	assert.Equal(t, "bar", v2)

	type testStruct struct {
		value int
	}
	key := testStruct{value: 1}

	c3 := cache.New[testStruct, string](0, 0)
	c3.Add(key, "snafu")
	v3, found3 := c3.Get(key)
	assert.True(t, found3)
	assert.Equal(t, "snafu", v3)
}
