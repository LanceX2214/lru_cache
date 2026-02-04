package lru_cache

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestGroupGetCacheHit(t *testing.T) {
	var load_count int32
	getter := GetterFunc(func(key string) ([]byte, error) {
		atomic.AddInt32(&load_count, 1)
		return []byte("value"), nil
	})

	group := NewGroup("test_group_hit", 1<<20, getter)

	value, error_value := group.Get("k1")
	if error_value != nil {
		t.Fatalf("unexpected error: %v", error_value)
	}
	if value.String() != "value" {
		t.Fatalf("unexpected value: %s", value.String())
	}

	value, error_value = group.Get("k1")
	if error_value != nil {
		t.Fatalf("unexpected error: %v", error_value)
	}
	if atomic.LoadInt32(&load_count) != 1 {
		t.Fatalf("expected getter to be called once, got %d", load_count)
	}
}

func TestGroupExpiration(t *testing.T) {
	var load_count int32
	getter := GetterFunc(func(key string) ([]byte, error) {
		atomic.AddInt32(&load_count, 1)
		return []byte("value"), nil
	})

	group := NewGroup("test_group_expire", 1<<20, getter, WithExpiration(10*time.Millisecond))

	_, error_value := group.Get("k1")
	if error_value != nil {
		t.Fatalf("unexpected error: %v", error_value)
	}

	time.Sleep(20 * time.Millisecond)

	_, error_value = group.Get("k1")
	if error_value != nil {
		t.Fatalf("unexpected error: %v", error_value)
	}

	if atomic.LoadInt32(&load_count) < 2 {
		t.Fatalf("expected getter to be called at least twice after expiration")
	}
}
