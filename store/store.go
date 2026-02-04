package store

// Value is the interface values must implement to be cached.
type Value interface {
	Len() int
}

// Store is the cache storage interface.
type Store interface {
	Get(key string) (Value, bool)
	Add(key string, value Value)
	Remove(key string)
	Len() int
	Bytes() int64
}
