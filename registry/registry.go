package registry

import (
	"context"
	"fmt"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Registry provides service registration and discovery.
type Registry struct {
	client *clientv3.Client
}

// Registration tracks a lease-backed service registration.
type Registration struct {
	client       *clientv3.Client
	lease_id     clientv3.LeaseID
	key_name     string
	stop_channel chan struct{}
}

// New creates a Registry.
func New(client *clientv3.Client) *Registry {
	return &Registry{client: client}
}

// Register registers a service address with TTL.
func (registry *Registry) Register(request_context context.Context, service_name, address string, ttl time.Duration) (*Registration, error) {
	lease, error_value := registry.client.Grant(request_context, int64(ttl.Seconds()))
	if error_value != nil {
		return nil, error_value
	}
	key_name := fmt.Sprintf("%s/%s", service_name, address)
	if _, error_value := registry.client.Put(request_context, key_name, address, clientv3.WithLease(lease.ID)); error_value != nil {
		return nil, error_value
	}
	keep_alive_channel, error_value := registry.client.KeepAlive(request_context, lease.ID)
	if error_value != nil {
		return nil, error_value
	}
	registration := &Registration{client: registry.client, lease_id: lease.ID, key_name: key_name, stop_channel: make(chan struct{})}
	go registration.keep_alive(keep_alive_channel)
	return registration, nil
}

func (registration *Registration) keep_alive(keep_alive_channel <-chan *clientv3.LeaseKeepAliveResponse) {
	for {
		select {
		case <-registration.stop_channel:
			return
		case _, ok := <-keep_alive_channel:
			if !ok {
				return
			}
		}
	}
}

// Close revokes the lease.
func (registration *Registration) Close(request_context context.Context) error {
	close(registration.stop_channel)
	_, error_value := registration.client.Revoke(request_context, registration.lease_id)
	return error_value
}

// List returns addresses for a service.
func (registry *Registry) List(request_context context.Context, service_name string) ([]string, error) {
	prefix := service_name + "/"
	response, error_value := registry.client.Get(request_context, prefix, clientv3.WithPrefix())
	if error_value != nil {
		return nil, error_value
	}
	addresses := make([]string, 0, len(response.Kvs))
	for _, key_value := range response.Kvs {
		address := strings.TrimPrefix(string(key_value.Key), prefix)
		if address == "" {
			continue
		}
		addresses = append(addresses, address)
	}
	return addresses, nil
}

// Watch watches service changes and emits full address lists.
func (registry *Registry) Watch(request_context context.Context, service_name string) <-chan []string {
	change_channel := make(chan []string, 1)
	prefix := service_name + "/"
	go func() {
		defer close(change_channel)
		send_snapshot := func() {
			addresses, error_value := registry.List(request_context, service_name)
			if error_value != nil {
				return
			}
			select {
			case change_channel <- addresses:
			default:
			}
		}
		send_snapshot()
		watch_channel := registry.client.Watch(request_context, prefix, clientv3.WithPrefix())
		for {
			select {
			case <-request_context.Done():
				return
			case _, ok := <-watch_channel:
				if !ok {
					return
				}
				send_snapshot()
			}
		}
	}()
	return change_channel
}
