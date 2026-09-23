package grpcx

import (
	"context"
	"log/slog"
	"time"

	"backend/pkg/requestid"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Dial(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(requestIDClientInterceptor),
	)
}

func NewServer(opts ...grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(append([]grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			recoveryInterceptor,
			requestIDServerInterceptor,
			metricsServerInterceptor,
		),
	}, opts...)...)
}

func GracefulStop(srv *grpc.Server, timeout time.Duration) {
	stopped := make(chan struct{})
	go func() {
		srv.GracefulStop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(timeout):
		slog.Warn("grpc graceful stop timed out, forcing")
		srv.Stop()
	}
}

func recoveryInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic in grpc handler", "panic", r, "request_id", requestid.FromContext(ctx))
			err = status.Error(codes.Internal, "internal error")
		}
	}()
	return handler(ctx, req)
}

func requestIDClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if id := requestid.FromContext(ctx); id != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, requestid.Header, id)
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

func requestIDServerInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ids := md.Get(requestid.Header); len(ids) > 0 && ids[0] != "" {
			ctx = requestid.WithContext(ctx, ids[0])
		}
	}
	return handler(ctx, req)
}
