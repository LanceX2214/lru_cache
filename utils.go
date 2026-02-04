package lru_cache

import "errors"

var (
	ErrNotFound = errors.New("lru_cache: key not found")
	ErrEmptyKey = errors.New("lru_cache: empty key")
)

func clone_bytes(bytes []byte) []byte {
	if bytes == nil {
		return nil
	}
	copied_bytes := make([]byte, len(bytes))
	copy(copied_bytes, bytes)
	return copied_bytes
}
