package consistenthash

import "testing"

func TestConsistentHashBasic(t *testing.T) {
	hash_ring := New(3, nil)
	hash_ring.Add("nodeA", "nodeB")

	value := hash_ring.Get("key1")
	if value != "nodeA" && value != "nodeB" {
		t.Fatalf("unexpected node for key1: %s", value)
	}

	hash_ring.Remove("nodeA")
	value = hash_ring.Get("key1")
	if value != "nodeB" {
		t.Fatalf("expected key1 to map to nodeB after removing nodeA, got %s", value)
	}
}
