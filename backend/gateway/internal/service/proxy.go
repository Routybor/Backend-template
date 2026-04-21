package service

import (
	"context"

	"gateway/internal/config"
	pb "gateway/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ItemGrpcClient struct {
	conn   *grpc.ClientConn
	client pb.ItemServiceClient
}

func NewItemGrpcClient(grpcAddr string) (*ItemGrpcClient, error) {
	conn, err := grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &ItemGrpcClient{
		conn:   conn,
		client: pb.NewItemServiceClient(conn),
	}, nil
}

func (c *ItemGrpcClient) GetAll(ctx context.Context) (*pb.ListItemsResponse, error) {
	return c.client.GetAll(ctx, &pb.GetAllRequest{})
}

func (c *ItemGrpcClient) GetByID(ctx context.Context, id string) (*pb.Item, error) {
	return c.client.GetByID(ctx, &pb.GetByIDRequest{Id: id})
}

func (c *ItemGrpcClient) Create(ctx context.Context, name, description string) (*pb.Item, error) {
	return c.client.Create(ctx, &pb.CreateItemRequest{
		Name:        name,
		Description: description,
	})
}

func (c *ItemGrpcClient) Delete(ctx context.Context, id string) (*pb.DeleteResponse, error) {
	return c.client.Delete(ctx, &pb.DeleteRequest{Id: id})
}

func (c *ItemGrpcClient) Close() error {
	return c.conn.Close()
}

func (c *ItemGrpcClient) Health(ctx context.Context) (*pb.HealthResponse, error) {
	return c.client.Health(ctx, &pb.HealthRequest{})
}

var NewReverseProxy = NewItemGrpcClient

func LoadProxyConfig(cfg *config.Config) (*ItemGrpcClient, error) {
	return NewItemGrpcClient(cfg.CoreServiceGrpc)
}
