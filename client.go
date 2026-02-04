package lru_cache

import (
	"context"
	"errors"
	"sync"

	"lru_cache/consistenthash"
	"lru_cache/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is a gRPC peer client.
type Client struct {
	address     string
	connection  *grpc.ClientConn
	grpc_client pb.LCacheClient
}

// NewClient connects to a peer address.
func NewClient(address string) (*Client, error) {
	connection, error_value := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype(pb.SubtypeJSON)),
	)
	if error_value != nil {
		return nil, error_value
	}
	return &Client{
		address:     address,
		connection:  connection,
		grpc_client: pb.NewLCacheClient(connection),
	}, nil
}

// Get fetches data from remote peer.
func (client *Client) Get(request_context context.Context, group_name string, key string) ([]byte, error) {
	response, error_value := client.grpc_client.Get(request_context, &pb.GetRequest{Group: group_name, Key: key})
	if error_value != nil {
		return nil, error_value
	}
	if response.Err != "" {
		if response.Err == ErrNotFound.Error() {
			return nil, ErrNotFound
		}
		return nil, errors.New(response.Err)
	}
	return response.Value, nil
}

// Close closes the client connection.
func (client *Client) Close() error {
	if client.connection != nil {
		return client.connection.Close()
	}
	return nil
}

// ClientPicker picks peers using consistent hashing.
type ClientPicker struct {
	self_address  string
	replica_count int

	mutex        sync.RWMutex
	hash_ring    *consistenthash.Map
	peer_clients map[string]*Client
}

// NewClientPicker creates a picker with default replicas.
func NewClientPicker(self_address string) *ClientPicker {
	return &ClientPicker{self_address: self_address, replica_count: 50}
}

// SetReplicas sets virtual node replicas.
func (picker *ClientPicker) SetReplicas(replica_count int) {
	picker.replica_count = replica_count
}

// SetPeers replaces the peer list.
func (picker *ClientPicker) SetPeers(peer_addresses ...string) {
	picker.mutex.Lock()
	defer picker.mutex.Unlock()

	hash_ring := consistenthash.New(picker.replica_count, nil)
	hash_ring.Add(peer_addresses...)

	new_peer_clients := make(map[string]*Client)
	for _, address := range peer_addresses {
		if address == "" {
			continue
		}
		if existing_client, ok := picker.peer_clients[address]; ok {
			new_peer_clients[address] = existing_client
			continue
		}
		client, error_value := NewClient(address)
		if error_value != nil {
			continue
		}
		new_peer_clients[address] = client
	}
	for address, client := range picker.peer_clients {
		if _, ok := new_peer_clients[address]; !ok {
			_ = client.Close()
		}
	}
	picker.hash_ring = hash_ring
	picker.peer_clients = new_peer_clients
}

// PickPeer returns a peer for the given key.
func (picker *ClientPicker) PickPeer(key string) (PeerGetter, bool) {
	picker.mutex.RLock()
	defer picker.mutex.RUnlock()
	if picker.hash_ring == nil || len(picker.peer_clients) == 0 {
		return nil, false
	}
	peer_address := picker.hash_ring.Get(key)
	if peer_address == "" || peer_address == picker.self_address {
		return nil, false
	}
	client, ok := picker.peer_clients[peer_address]
	return client, ok
}

// Close closes all clients.
func (picker *ClientPicker) Close() {
	picker.mutex.Lock()
	defer picker.mutex.Unlock()
	for _, client := range picker.peer_clients {
		_ = client.Close()
	}
	picker.peer_clients = nil
	picker.hash_ring = nil
}
