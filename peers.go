package lru_cache

import "context"

// PeerPicker selects a peer for a given key.
type PeerPicker interface {
	PickPeer(key string) (PeerGetter, bool)
}

// PeerGetter fetches data from a peer.
type PeerGetter interface {
	Get(request_context context.Context, group_name string, key string) ([]byte, error)
}
