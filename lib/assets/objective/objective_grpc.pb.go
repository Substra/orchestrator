// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package objective

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// ObjectiveServiceClient is the client API for ObjectiveService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ObjectiveServiceClient interface {
	RegisterObjective(ctx context.Context, in *Objective, opts ...grpc.CallOption) (*Objective, error)
	QueryObjective(ctx context.Context, in *ObjectiveQuery, opts ...grpc.CallOption) (*Objective, error)
}

type objectiveServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewObjectiveServiceClient(cc grpc.ClientConnInterface) ObjectiveServiceClient {
	return &objectiveServiceClient{cc}
}

func (c *objectiveServiceClient) RegisterObjective(ctx context.Context, in *Objective, opts ...grpc.CallOption) (*Objective, error) {
	out := new(Objective)
	err := c.cc.Invoke(ctx, "/objective.ObjectiveService/RegisterObjective", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectiveServiceClient) QueryObjective(ctx context.Context, in *ObjectiveQuery, opts ...grpc.CallOption) (*Objective, error) {
	out := new(Objective)
	err := c.cc.Invoke(ctx, "/objective.ObjectiveService/QueryObjective", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ObjectiveServiceServer is the server API for ObjectiveService service.
// All implementations must embed UnimplementedObjectiveServiceServer
// for forward compatibility
type ObjectiveServiceServer interface {
	RegisterObjective(context.Context, *Objective) (*Objective, error)
	QueryObjective(context.Context, *ObjectiveQuery) (*Objective, error)
	mustEmbedUnimplementedObjectiveServiceServer()
}

// UnimplementedObjectiveServiceServer must be embedded to have forward compatible implementations.
type UnimplementedObjectiveServiceServer struct {
}

func (UnimplementedObjectiveServiceServer) RegisterObjective(context.Context, *Objective) (*Objective, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterObjective not implemented")
}
func (UnimplementedObjectiveServiceServer) QueryObjective(context.Context, *ObjectiveQuery) (*Objective, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryObjective not implemented")
}
func (UnimplementedObjectiveServiceServer) mustEmbedUnimplementedObjectiveServiceServer() {}

// UnsafeObjectiveServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ObjectiveServiceServer will
// result in compilation errors.
type UnsafeObjectiveServiceServer interface {
	mustEmbedUnimplementedObjectiveServiceServer()
}

func RegisterObjectiveServiceServer(s grpc.ServiceRegistrar, srv ObjectiveServiceServer) {
	s.RegisterService(&_ObjectiveService_serviceDesc, srv)
}

func _ObjectiveService_RegisterObjective_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Objective)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectiveServiceServer).RegisterObjective(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/objective.ObjectiveService/RegisterObjective",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectiveServiceServer).RegisterObjective(ctx, req.(*Objective))
	}
	return interceptor(ctx, in, info, handler)
}

func _ObjectiveService_QueryObjective_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ObjectiveQuery)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectiveServiceServer).QueryObjective(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/objective.ObjectiveService/QueryObjective",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectiveServiceServer).QueryObjective(ctx, req.(*ObjectiveQuery))
	}
	return interceptor(ctx, in, info, handler)
}

var _ObjectiveService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "objective.ObjectiveService",
	HandlerType: (*ObjectiveServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterObjective",
			Handler:    _ObjectiveService_RegisterObjective_Handler,
		},
		{
			MethodName: "QueryObjective",
			Handler:    _ObjectiveService_QueryObjective_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "objective.proto",
}
