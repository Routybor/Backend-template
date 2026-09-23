package infrastructure

import (
	"context"
	"errors"
	"log/slog"

	"backend/pkg/pb"
	"backend/pkg/requestid"
	"core-service/internal/domain"
	"core-service/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	pb.UnimplementedItemServiceServer
	svc *service.ItemService
}

func NewGRPCHandler(svc *service.ItemService) *GRPCHandler {
	return &GRPCHandler{svc: svc}
}

func (h *GRPCHandler) GetAll(ctx context.Context, req *pb.GetAllRequest) (*pb.ListItemsResponse, error) {
	items := h.svc.GetAll(ctx)
	out := make([]*pb.Item, len(items))
	for i, item := range items {
		out[i] = toItemPb(item)
	}
	return &pb.ListItemsResponse{Items: out, Total: int64(len(out))}, nil
}

func (h *GRPCHandler) GetByID(ctx context.Context, req *pb.GetByIDRequest) (*pb.Item, error) {
	item, err := h.svc.Get(ctx, req.Id)
	if err != nil {
		return nil, mapDomainError(ctx, err)
	}
	return toItemPb(item), nil
}

func (h *GRPCHandler) Create(ctx context.Context, req *pb.CreateItemRequest) (*pb.Item, error) {
	item, err := h.svc.Create(ctx, req.Name, req.Description)
	if err != nil {
		return nil, mapDomainError(ctx, err)
	}
	return toItemPb(item), nil
}

func (h *GRPCHandler) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if err := h.svc.Delete(ctx, req.Id); err != nil {
		return nil, mapDomainError(ctx, err)
	}
	return &pb.DeleteResponse{Success: true}, nil
}

func (h *GRPCHandler) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{Status: "healthy"}, nil
}

func toItemPb(item domain.Item) *pb.Item {
	return &pb.Item{
		Id:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		CreatedAt:   item.CreatedAt.Unix(),
		UpdatedAt:   item.UpdatedAt.Unix(),
	}
}

func mapDomainError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, "item not found")
	case errors.Is(err, domain.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		slog.Error("unexpected service error", "error", err, "request_id", requestid.FromContext(ctx))
		return status.Error(codes.Internal, "internal error")
	}
}
