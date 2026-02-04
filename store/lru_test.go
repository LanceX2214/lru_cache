package store

import "testing"

type test_value string

func (value test_value) Len() int {
	return len(value)
}

func TestLRUEviction(t *testing.T) {
	cache := NewLRU(10, nil)
	cache.Add("k1", test_value("v1"))
	cache.Add("k2", test_value("v2"))
	cache.Add("k3", test_value("v3"))

	if _, ok := cache.Get("k1"); ok {
		t.Fatalf("expected k1 to be evicted")
	}
	if _, ok := cache.Get("k2"); !ok {
		t.Fatalf("expected k2 to remain")
	}
	if _, ok := cache.Get("k3"); !ok {
		t.Fatalf("expected k3 to remain")
	}

	// Access k2 to make it most recent, then add k4 to evict k3.
	cache.Get("k2")
	cache.Add("k4", test_value("v4"))

	if _, ok := cache.Get("k3"); ok {
		t.Fatalf("expected k3 to be evicted after adding k4")
	}
	if _, ok := cache.Get("k2"); !ok {
		t.Fatalf("expected k2 to remain after adding k4")
	}
	if _, ok := cache.Get("k4"); !ok {
		t.Fatalf("expected k4 to remain after adding k4")
	}
}
