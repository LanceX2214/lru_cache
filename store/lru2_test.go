package store

import "testing"

func TestLRU2Promotion(t *testing.T) {
	cache := NewLRU2(20, nil)
	cache.Add("k1", test_value("v1"))

	if _, ok := cache.Get("k1"); !ok {
		t.Fatalf("expected k1 to be returned from history and promoted")
	}

	cache.Add("k2", test_value("v2"))
	cache.Add("k3", test_value("v3"))
	cache.Add("k4", test_value("v4"))
	cache.Add("k5", test_value("v5"))

	if _, ok := cache.Get("k1"); !ok {
		t.Fatalf("expected k1 to remain in main cache after history churn")
	}
}
