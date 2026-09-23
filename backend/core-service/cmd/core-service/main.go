package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/pkg/grpcx"
	pb "backend/pkg/pb"

	"core-service/internal/data"
	"core-service/internal/infrastructure"
	"core-service/internal/service"
)

func main() {
	cfg, err := infrastructure.LoadConfig()
	if err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}

	store := data.NewMemoryItemStore()
	svc := service.NewItemService(store)

	grpcSrv := grpcx.NewServer()
	pb.RegisterItemServiceServer(grpcSrv, infrastructure.NewGRPCHandler(svc))

	lis, err := net.Listen("tcp", ":"+cfg.GrpcPort)
	if err != nil {
		slog.Error("failed to listen for grpc", "error", err, "port", cfg.GrpcPort)
		os.Exit(1)
	}

	errCh := make(chan error, 2)
	go func() {
		slog.Info("grpc server listening", "port", cfg.GrpcPort)
		if err := grpcSrv.Serve(lis); err != nil {
			errCh <- fmt.Errorf("grpc server: %w", err)
		}
	}()

	checks := map[string]func(context.Context) error{
		"store": svc.Ping,
	}
	httpSrv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      infrastructure.NewHTTPRouter(svc, checks),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		slog.Info("http server listening", "port", cfg.Port)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http server: %w", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		slog.Error("server error", "error", err)
		_ = httpSrv.Close()
		grpcSrv.Stop()
		os.Exit(1)
	case <-ctx.Done():
	}

	slog.Info("shutting down core-service")

	go grpcx.GracefulStop(grpcSrv, 15*time.Second)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		slog.Error("http shutdown failed", "error", err)
	}

	slog.Info("core-service stopped")
}
