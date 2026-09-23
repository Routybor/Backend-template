package service

import (
	"context"
	"time"

	"backend/pkg/grpcx"
	pb "backend/pkg/pb"

	"google.golang.org/grpc"
)

const callTimeout = 10 * time.Second

type ItemsClient struct {
	conn   *grpc.ClientConn
	client pb.ItemServiceClient
}

func NewItemsClient(addr string) (*ItemsClient, error) {
	conn, err := grpcx.Dial(addr)
	if err != nil {
		return nil, err
	}

	return &ItemsClient{
		conn:   conn,
		client: pb.NewItemServiceClient(conn),
	}, nil
}

func (c *ItemsClient) GetAll(ctx context.Context) (*pb.ListItemsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	return c.client.GetAll(ctx, &pb.GetAllRequest{})
}

func (c *ItemsClient) GetByID(ctx context.Context, id string) (*pb.Item, error) {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	return c.client.GetByID(ctx, &pb.GetByIDRequest{Id: id})
}

func (c *ItemsClient) Create(ctx context.Context, name, description string) (*pb.Item, error) {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	return c.client.Create(ctx, &pb.CreateItemRequest{
		Name:        name,
		Description: description,
	})
}

func (c *ItemsClient) Delete(ctx context.Context, id string) (*pb.DeleteResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	return c.client.Delete(ctx, &pb.DeleteRequest{Id: id})
}

func (c *ItemsClient) Close() error {
	return c.conn.Close()
}
