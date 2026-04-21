package main

import (
	"context"
	"core-service/internal/config"
	"core-service/internal/dto"
	"core-service/internal/handler"
	"core-service/internal/repository"
	"core-service/internal/service"
	pb "core-service/proto"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type gRPCServer struct {
	pb.UnimplementedItemServiceServer
	itemService *service.ItemService
}

func (s *gRPCServer) GetAll(ctx context.Context, req *pb.GetAllRequest) (*pb.ListItemsResponse, error) {
	items := s.itemService.GetAllItems()
	pbItems := make([]*pb.Item, len(items))
	for i, item := range items {
		pbItems[i] = &pb.Item{
			Id:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			CreatedAt:   0,
			UpdatedAt:   0,
		}
	}
	return &pb.ListItemsResponse{Items: pbItems, Total: int64(len(items))}, nil
}

func (s *gRPCServer) GetByID(ctx context.Context, req *pb.GetByIDRequest) (*pb.Item, error) {
	item, ok := s.itemService.GetItem(req.Id)
	if !ok {
		return nil, nil
	}
	return &pb.Item{
		Id:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		CreatedAt:   0,
		UpdatedAt:   0,
	}, nil
}

func (s *gRPCServer) Create(ctx context.Context, req *pb.CreateItemRequest) (*pb.Item, error) {
	dtoReq := dto.CreateItemRequest{
		Name:        req.Name,
		Description: req.Description,
	}
	item := s.itemService.CreateItem(dtoReq)
	return &pb.Item{
		Id:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		CreatedAt:   0,
		UpdatedAt:   0,
	}, nil
}

func (s *gRPCServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	success := s.itemService.DeleteItem(req.Id)
	return &pb.DeleteResponse{Success: success}, nil
}

func (s *gRPCServer) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{Status: "healthy"}, nil
}

func main() {
	slog.Info("starting core-service")

	cfg := config.Load()

	itemRepo := repository.NewItemRepository()
	itemService := service.NewItemService(itemRepo)

	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GrpcPort)
		if err != nil {
			slog.Error("failed to listen for gRPC", "error", err)
			os.Exit(1)
		}
		grpcServer := grpc.NewServer()
		pb.RegisterItemServiceServer(grpcServer, &gRPCServer{itemService: itemService})
		slog.Info("gRPC server listening", "port", cfg.GrpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("gRPC server failed", "error", err)
			os.Exit(1)
		}
	}()

	itemHandler := handler.NewItemHandler(itemService)
	healthHandler := handler.NewHealthHandler()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger())
	router.Use(securityHeaders())

	router.GET("/health", healthHandler.Check)
	router.GET("/health/live", healthHandler.Check)
	router.GET("/health/ready", healthHandler.Check)

	router.GET("/items", itemHandler.GetAll)
	router.POST("/items", itemHandler.Create)
	router.GET("/items/:id", itemHandler.GetByID)
	router.DELETE("/items/:id", itemHandler.Delete)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("core-service HTTP listening", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		slog.Info("request completed",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", time.Since(start),
			"client_ip", c.ClientIP(),
			"user_id", c.GetHeader("X-User-ID"),
		)
	}
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}
