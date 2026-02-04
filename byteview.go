package lru_cache

// ByteView is an immutable view of cached bytes.
type ByteView struct {
	bytes []byte
}

func (view ByteView) Len() int {
	return len(view.bytes)
}

func (view ByteView) ByteSlice() []byte {
	return clone_bytes(view.bytes)
}

func (view ByteView) String() string {
	return string(view.bytes)
}
