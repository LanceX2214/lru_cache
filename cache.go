package lru_cache

import (
	"sync"
	"sync/atomic"
	"time"

	"lru_cache/store"
)

// CacheOptions configures the local cache store.
type CacheOptions struct {
	Store_type string // "lru" or "lru2"
	Max_bytes  int64
}

// Cache holds a local in-memory cache.
type Cache struct {
	mutex      sync.Mutex
	store      store.Store
	options    CacheOptions
	hit_count  uint64
	miss_count uint64
}

type cache_value struct {
	value     ByteView
	expire_at int64
}

func (value *cache_value) Len() int {
	return value.value.Len()
}

func (value *cache_value) expired(current_time int64) bool {
	return value.expire_at > 0 && current_time >= value.expire_at
}

// NewCache creates a cache with options.
func NewCache(options CacheOptions) *Cache {
	return &Cache{
		store:   new_store(options),
		options: options,
	}
}

func new_store(options CacheOptions) store.Store {
	switch options.Store_type {
	case "lru2":
		return store.NewLRU2(options.Max_bytes, nil)
	default:
		return store.NewLRU(options.Max_bytes, nil)
	}
}

// Get returns a value from cache.
func (cache *Cache) Get(key string) (ByteView, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	stored_value, ok := cache.store.Get(key)
	if !ok {
		atomic.AddUint64(&cache.miss_count, 1)
		return ByteView{}, false
	}
	cache_value := stored_value.(*cache_value)
	if cache_value.expired(time.Now().UnixNano()) {
		cache.store.Remove(key)
		atomic.AddUint64(&cache.miss_count, 1)
		return ByteView{}, false
	}
	atomic.AddUint64(&cache.hit_count, 1)
	return cache_value.value, true
}

// Set stores a value with optional ttl (0 means no expiration).
func (cache *Cache) Set(key string, value ByteView, ttl time.Duration) {
	var expire_at int64
	if ttl > 0 {
		expire_at = time.Now().Add(ttl).UnixNano()
	}
	stored_value := &cache_value{value: value, expire_at: expire_at}

	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.store.Add(key, stored_value)
}

// Remove deletes a key.
func (cache *Cache) Remove(key string) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.store.Remove(key)
}

// Len returns the number of entries in the cache.
func (cache *Cache) Len() int {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	return cache.store.Len()
}

// Bytes returns the total bytes tracked by the cache.
func (cache *Cache) Bytes() int64 {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	return cache.store.Bytes()
}

// Stats returns cache hit/miss stats.
func (cache *Cache) Stats() (hits uint64, misses uint64) {
	return atomic.LoadUint64(&cache.hit_count), atomic.LoadUint64(&cache.miss_count)
}
