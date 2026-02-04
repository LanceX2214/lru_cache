package pb

import (
	"context"

	"google.golang.org/grpc"
)

// GetRequest is the cache fetch request.
type GetRequest struct {
	Group string `json:"group"`
	Key   string `json:"key"`
}

// GetResponse is the cache fetch response.
type GetResponse struct {
	Value []byte `json:"value,omitempty"`
	Err   string `json:"err,omitempty"`
}

// LCacheClient is the client API for LCache service.
type LCacheClient interface {
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
}

type lCacheClient struct {
	cc grpc.ClientConnInterface
}

// NewLCacheClient creates a new client.
func NewLCacheClient(cc grpc.ClientConnInterface) LCacheClient {
	return &lCacheClient{cc}
}

func (c *lCacheClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, "/lcache.LCache/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LCacheServer is the server API for LCache service.
type LCacheServer interface {
	Get(context.Context, *GetRequest) (*GetResponse, error)
}

// RegisterLCacheServer registers the server.
func RegisterLCacheServer(s *grpc.Server, srv LCacheServer) {
	s.RegisterService(&LCache_ServiceDesc, srv)
}

// LCache_ServiceDesc is the grpc.ServiceDesc for LCache service.
var LCache_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "lcache.LCache",
	HandlerType: (*LCacheServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _LCache_Get_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "lcache.proto",
}

func _LCache_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LCacheServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lcache.LCache/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LCacheServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}
