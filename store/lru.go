package store

import "container/list"

type entry struct {
	key   string
	value Value
}

// LRU implements a non-thread-safe LRU cache.
type LRU struct {
	max_bytes  int64
	used_bytes int64
	list       *list.List
	entry_map  map[string]*list.Element
	on_evicted func(key string, value Value)
}

// NewLRU creates an LRU with maxBytes (0 means no limit).
func NewLRU(max_bytes int64, on_evicted func(string, Value)) *LRU {
	return &LRU{
		max_bytes:  max_bytes,
		list:       list.New(),
		entry_map:  make(map[string]*list.Element),
		on_evicted: on_evicted,
	}
}

func (cache *LRU) Get(key string) (Value, bool) {
	if element, ok := cache.entry_map[key]; ok {
		cache.list.MoveToFront(element)
		cache_entry := element.Value.(*entry)
		return cache_entry.value, true
	}
	return nil, false
}

func (cache *LRU) Add(key string, value Value) {
	if element, ok := cache.entry_map[key]; ok {
		cache.list.MoveToFront(element)
		cache_entry := element.Value.(*entry)
		cache.used_bytes += int64(value.Len()) - int64(cache_entry.value.Len())
		cache_entry.value = value
		return
	}
	cache_entry := &entry{key: key, value: value}
	element := cache.list.PushFront(cache_entry)
	cache.entry_map[key] = element
	cache.used_bytes += int64(len(key)) + int64(value.Len())

	for cache.max_bytes != 0 && cache.used_bytes > cache.max_bytes {
		cache.remove_oldest(true)
	}
}

func (cache *LRU) Remove(key string) {
	if element, ok := cache.entry_map[key]; ok {
		cache.remove_element(element, true)
	}
}

func (cache *LRU) Len() int {
	return cache.list.Len()
}

func (cache *LRU) Bytes() int64 {
	return cache.used_bytes
}

func (cache *LRU) remove_oldest(call_evicted bool) {
	element := cache.list.Back()
	if element != nil {
		cache.remove_element(element, call_evicted)
	}
}

func (cache *LRU) remove_element(element *list.Element, call_evicted bool) {
	cache.list.Remove(element)
	cache_entry := element.Value.(*entry)
	delete(cache.entry_map, cache_entry.key)
	cache.used_bytes -= int64(len(cache_entry.key)) + int64(cache_entry.value.Len())
	if call_evicted && cache.on_evicted != nil {
		cache.on_evicted(cache_entry.key, cache_entry.value)
	}
}
