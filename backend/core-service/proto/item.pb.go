package proto

import (
	"context"

	"google.golang.org/grpc"
)

type Item struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type GetAllRequest struct{}

type GetByIDRequest struct {
	Id string `json:"id"`
}

type CreateItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ListItemsResponse struct {
	Items []*Item `json:"items"`
	Total int64   `json:"total"`
}

type DeleteRequest struct {
	Id string `json:"id"`
}

type DeleteResponse struct {
	Success bool `json:"success"`
}

type HealthRequest struct{}

type HealthResponse struct {
	Status string `json:"status"`
}

type ItemServiceClient interface {
	GetAll(ctx context.Context, in *GetAllRequest, opts ...grpc.CallOption) (*ListItemsResponse, error)
	GetByID(ctx context.Context, in *GetByIDRequest, opts ...grpc.CallOption) (*Item, error)
	Create(ctx context.Context, in *CreateItemRequest, opts ...grpc.CallOption) (*Item, error)
	Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error)
	Health(ctx context.Context, in *HealthRequest, opts ...grpc.CallOption) (*HealthResponse, error)
}

type itemServiceClient struct {
	cc *grpc.ClientConn
}

func NewItemServiceClient(cc *grpc.ClientConn) ItemServiceClient {
	return &itemServiceClient{cc: cc}
}

func (c *itemServiceClient) GetAll(ctx context.Context, in *GetAllRequest, opts ...grpc.CallOption) (*ListItemsResponse, error) {
	out := new(ListItemsResponse)
	err := c.cc.Invoke(ctx, "/core.ItemService/GetAll", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *itemServiceClient) GetByID(ctx context.Context, in *GetByIDRequest, opts ...grpc.CallOption) (*Item, error) {
	out := new(Item)
	err := c.cc.Invoke(ctx, "/core.ItemService/GetByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *itemServiceClient) Create(ctx context.Context, in *CreateItemRequest, opts ...grpc.CallOption) (*Item, error) {
	out := new(Item)
	err := c.cc.Invoke(ctx, "/core.ItemService/Create", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *itemServiceClient) Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error) {
	out := new(DeleteResponse)
	err := c.cc.Invoke(ctx, "/core.ItemService/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *itemServiceClient) Health(ctx context.Context, in *HealthRequest, opts ...grpc.CallOption) (*HealthResponse, error) {
	out := new(HealthResponse)
	err := c.cc.Invoke(ctx, "/core.ItemService/Health", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func RegisterItemServiceServer(s *grpc.Server, srv ItemServiceServer) {
	s.RegisterService(&_ItemService_serviceDesc, srv)
}

var _ItemService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "core.ItemService",
	HandlerType: (*ItemServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAll",
			Handler:    _ItemService_GetAll_Handler,
		},
		{
			MethodName: "GetByID",
			Handler:    _ItemService_GetByID_Handler,
		},
		{
			MethodName: "Create",
			Handler:    _ItemService_Create_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _ItemService_Delete_Handler,
		},
		{
			MethodName: "Health",
			Handler:    _ItemService_Health_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "item.proto",
}

func _ItemService_GetAll_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAllRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ItemServiceServer).GetAll(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/core.ItemService/GetAll",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ItemServiceServer).GetAll(ctx, req.(*GetAllRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ItemService_GetByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetByIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ItemServiceServer).GetByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/core.ItemService/GetByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ItemServiceServer).GetByID(ctx, req.(*GetByIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ItemService_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateItemRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ItemServiceServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/core.ItemService/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ItemServiceServer).Create(ctx, req.(*CreateItemRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ItemService_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ItemServiceServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/core.ItemService/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ItemServiceServer).Delete(ctx, req.(*DeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ItemService_Health_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ItemServiceServer).Health(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/core.ItemService/Health",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ItemServiceServer).Health(ctx, req.(*HealthRequest))
	}
	return interceptor(ctx, in, info, handler)
}

type ItemServiceServer interface {
	GetAll(context.Context, *GetAllRequest) (*ListItemsResponse, error)
	GetByID(context.Context, *GetByIDRequest) (*Item, error)
	Create(context.Context, *CreateItemRequest) (*Item, error)
	Delete(context.Context, *DeleteRequest) (*DeleteResponse, error)
	Health(context.Context, *HealthRequest) (*HealthResponse, error)
}

type UnimplementedItemServiceServer struct{}

func (UnimplementedItemServiceServer) GetAll(context.Context, *GetAllRequest) (*ListItemsResponse, error) {
	return nil, nil
}
func (UnimplementedItemServiceServer) GetByID(context.Context, *GetByIDRequest) (*Item, error) {
	return nil, nil
}
func (UnimplementedItemServiceServer) Create(context.Context, *CreateItemRequest) (*Item, error) {
	return nil, nil
}
func (UnimplementedItemServiceServer) Delete(context.Context, *DeleteRequest) (*DeleteResponse, error) {
	return nil, nil
}
func (UnimplementedItemServiceServer) Health(context.Context, *HealthRequest) (*HealthResponse, error) {
	return nil, nil
}
