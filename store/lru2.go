package store

// LRU2 is a simple 2-queue cache. Entries are inserted into history on first add;
// on second access they are promoted into the main cache.
type LRU2 struct {
	main_cache    *LRU
	history_cache *LRU
}

// NewLRU2 creates an LRU2 with a split of maxBytes (0 means no limit).
// History uses 1/4 of the space by default.
func NewLRU2(max_bytes int64, on_evicted func(string, Value)) *LRU2 {
	var history_max_bytes int64
	var main_max_bytes int64
	if max_bytes > 0 {
		history_max_bytes = max_bytes / 4
		if history_max_bytes < 1 {
			history_max_bytes = 1
		}
		main_max_bytes = max_bytes - history_max_bytes
	}
	return &LRU2{
		main_cache:    NewLRU(main_max_bytes, on_evicted),
		history_cache: NewLRU(history_max_bytes, nil),
	}
}

func (cache *LRU2) Get(key string) (Value, bool) {
	if value, ok := cache.main_cache.Get(key); ok {
		return value, true
	}
	if value, ok := cache.history_cache.Get(key); ok {
		cache.history_cache.Remove(key)
		cache.main_cache.Add(key, value)
		return value, true
	}
	return nil, false
}

func (cache *LRU2) Add(key string, value Value) {
	if _, ok := cache.main_cache.Get(key); ok {
		cache.main_cache.Add(key, value)
		return
	}
	if _, ok := cache.history_cache.Get(key); ok {
		cache.history_cache.Remove(key)
		cache.main_cache.Add(key, value)
		return
	}
	cache.history_cache.Add(key, value)
}

func (cache *LRU2) Remove(key string) {
	cache.main_cache.Remove(key)
	cache.history_cache.Remove(key)
}

func (cache *LRU2) Len() int {
	return cache.main_cache.Len() + cache.history_cache.Len()
}

func (cache *LRU2) Bytes() int64 {
	return cache.main_cache.Bytes() + cache.history_cache.Bytes()
}
