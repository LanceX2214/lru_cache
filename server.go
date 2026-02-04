package lru_cache

import (
	"context"
	"net"
	"time"

	"lru_cache/pb"
	"lru_cache/registry"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

// Server hosts the cache service.
type Server struct {
	address      string
	service_name string
	grpc_server  *grpc.Server
	listener     net.Listener
	etcd_client  *clientv3.Client
	registration *registry.Registration
}

// NewServer creates a new server.
func NewServer(address, service_name string) *Server {
	return &Server{address: address, service_name: service_name}
}

// Start starts the gRPC server.
func (server *Server) Start() error {
	listener, error_value := net.Listen("tcp", server.address)
	if error_value != nil {
		return error_value
	}
	server.listener = listener
	server.grpc_server = grpc.NewServer()
	pb.RegisterLCacheServer(server.grpc_server, server)
	go server.grpc_server.Serve(listener)
	return nil
}

// Stop stops the server.
func (server *Server) Stop() {
	if server.registration != nil {
		_ = server.registration.Close(context.Background())
		server.registration = nil
	}
	if server.grpc_server != nil {
		server.grpc_server.GracefulStop()
	}
	if server.etcd_client != nil {
		_ = server.etcd_client.Close()
	}
}

// RegisterEtcd registers service info in etcd.
func (server *Server) RegisterEtcd(request_context context.Context, endpoints []string, ttl time.Duration) error {
	client, error_value := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 3 * time.Second,
	})
	if error_value != nil {
		return error_value
	}
	server.etcd_client = client
	registry_client := registry.New(client)
	registration, error_value := registry_client.Register(request_context, server.service_name, server.address, ttl)
	if error_value != nil {
		return error_value
	}
	server.registration = registration
	return nil
}

// Get handles peer cache requests.
func (server *Server) Get(request_context context.Context, request *pb.GetRequest) (*pb.GetResponse, error) {
	group := GetGroup(request.Group)
	if group == nil {
		return &pb.GetResponse{Err: ErrNotFound.Error()}, nil
	}
	view, error_value := group.Get(request.Key)
	if error_value != nil {
		return &pb.GetResponse{Err: error_value.Error()}, nil
	}
	return &pb.GetResponse{Value: view.ByteSlice()}, nil
}
