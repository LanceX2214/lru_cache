package lru_cache

import (
	"context"
	"sync"
	"time"

	"lru_cache/singleflight"
)

// Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc adapts a function to Getter.
type GetterFunc func(key string) ([]byte, error)

func (getter GetterFunc) Get(key string) ([]byte, error) {
	return getter(key)
}

// Group is a cache namespace.
type Group struct {
	group_name         string
	data_getter        Getter
	main_cache         *Cache
	cache_options      CacheOptions
	peer_picker        PeerPicker
	load_group         *singleflight.Group
	default_expiration time.Duration
}

var (
	groups_mutex sync.RWMutex
	group_map    = make(map[string]*Group)
)

// GroupOption configures Group.
type GroupOption func(*Group)

// WithExpiration sets default expiration for entries in the group.
func WithExpiration(expiration_duration time.Duration) GroupOption {
	return func(group *Group) { group.default_expiration = expiration_duration }
}

// WithStoreType selects the local store implementation.
func WithStoreType(store_type string) GroupOption {
	return func(group *Group) { group.cache_options.Store_type = store_type }
}

// WithPeers registers a peer picker for distributed cache.
func WithPeers(peer_picker PeerPicker) GroupOption {
	return func(group *Group) { group.peer_picker = peer_picker }
}

// NewGroup creates a new cache group.
func NewGroup(group_name string, cache_bytes int64, data_getter Getter, options ...GroupOption) *Group {
	if data_getter == nil {
		panic("nil Getter")
	}
	group := &Group{
		group_name:    group_name,
		data_getter:   data_getter,
		cache_options: CacheOptions{Max_bytes: cache_bytes},
		load_group:    &singleflight.Group{},
	}
	for _, option := range options {
		option(group)
	}
	if group.main_cache == nil {
		group.main_cache = NewCache(group.cache_options)
	}
	groups_mutex.Lock()
	group_map[group_name] = group
	groups_mutex.Unlock()
	return group
}

// GetGroup retrieves a group by name.
func GetGroup(group_name string) *Group {
	groups_mutex.RLock()
	group := group_map[group_name]
	groups_mutex.RUnlock()
	return group
}

// Name returns the group name.
func (group *Group) Name() string {
	return group.group_name
}

// RegisterPeers sets the peer picker.
func (group *Group) RegisterPeers(peer_picker PeerPicker) {
	group.peer_picker = peer_picker
}

// Get retrieves a value for a key.
func (group *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, ErrEmptyKey
	}
	if value, ok := group.main_cache.Get(key); ok {
		return value, nil
	}
	return group.load(key)
}

// Set manually populates the cache.
func (group *Group) Set(key string, value []byte) {
	group.main_cache.Set(key, ByteView{bytes: clone_bytes(value)}, group.default_expiration)
}

func (group *Group) load(key string) (ByteView, error) {
	value_interface, error_value, _ := group.load_group.Do(key, func() (interface{}, error) {
		if group.peer_picker != nil {
			if peer_getter, ok := group.peer_picker.PickPeer(key); ok {
				if value, error_value := group.get_from_peer(peer_getter, key); error_value == nil {
					return value, nil
				}
			}
		}
		return group.get_locally(key)
	})
	if error_value != nil {
		return ByteView{}, error_value
	}
	return value_interface.(ByteView), nil
}

func (group *Group) get_locally(key string) (ByteView, error) {
	bytes, error_value := group.data_getter.Get(key)
	if error_value != nil {
		return ByteView{}, error_value
	}
	value := ByteView{bytes: clone_bytes(bytes)}
	group.populate_cache(key, value)
	return value, nil
}

func (group *Group) populate_cache(key string, value ByteView) {
	group.main_cache.Set(key, value, group.default_expiration)
}

func (group *Group) get_from_peer(peer_getter PeerGetter, key string) (ByteView, error) {
	request_context, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	bytes, error_value := peer_getter.Get(request_context, group.group_name, key)
	if error_value != nil {
		return ByteView{}, error_value
	}
	return ByteView{bytes: bytes}, nil
}
