package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

// Map implements consistent hashing.
type Map struct {
	hash_function Hash
	replica_count int
	sorted_keys   []int
	hash_map      map[int]string
}

// New creates a new Map.
func New(replica_count int, hash_function Hash) *Map {
	hash_map := &Map{
		replica_count: replica_count,
		hash_function: hash_function,
		hash_map:      make(map[int]string),
	}
	if hash_map.hash_function == nil {
		hash_map.hash_function = crc32.ChecksumIEEE
	}
	return hash_map
}

// Add adds keys to the hash.
func (hash_map *Map) Add(keys ...string) {
	for _, key := range keys {
		for replica_index := 0; replica_index < hash_map.replica_count; replica_index++ {
			hash_value := int(hash_map.hash_function([]byte(strconv.Itoa(replica_index) + key)))
			hash_map.sorted_keys = append(hash_map.sorted_keys, hash_value)
			hash_map.hash_map[hash_value] = key
		}
	}
	sort.Ints(hash_map.sorted_keys)
}

// Remove removes keys from the hash.
func (hash_map *Map) Remove(keys ...string) {
	if len(keys) == 0 || len(hash_map.sorted_keys) == 0 {
		return
	}
	remove := make(map[int]struct{})
	for _, key := range keys {
		for replica_index := 0; replica_index < hash_map.replica_count; replica_index++ {
			hash_value := int(hash_map.hash_function([]byte(strconv.Itoa(replica_index) + key)))
			remove[hash_value] = struct{}{}
			delete(hash_map.hash_map, hash_value)
		}
	}
	filtered := hash_map.sorted_keys[:0]
	for _, hash_value := range hash_map.sorted_keys {
		if _, ok := remove[hash_value]; !ok {
			filtered = append(filtered, hash_value)
		}
	}
	hash_map.sorted_keys = filtered
}

// Set resets and adds the provided keys.
func (hash_map *Map) Set(keys []string) {
	hash_map.sorted_keys = nil
	hash_map.hash_map = make(map[int]string)
	hash_map.Add(keys...)
}

// Get returns the closest item in the hash to the provided key.
func (hash_map *Map) Get(key string) string {
	if len(hash_map.sorted_keys) == 0 {
		return ""
	}
	hash_value := int(hash_map.hash_function([]byte(key)))
	index := sort.Search(len(hash_map.sorted_keys), func(i int) bool { return hash_map.sorted_keys[i] >= hash_value })
	if index == len(hash_map.sorted_keys) {
		index = 0
	}
	return hash_map.hash_map[hash_map.sorted_keys[index]]
}
