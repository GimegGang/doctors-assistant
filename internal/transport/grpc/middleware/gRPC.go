package gRPCMiddleware

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log/slog"
	"time"
)

type traceIDKey struct{}

func GRPCLogger(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		var traceID string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("x-trace-id"); len(values) > 0 {
				traceID = values[0]
			}
		}
		if traceID == "" {
			traceID = uuid.New().String()
		}

		ctx = context.WithValue(ctx, traceIDKey{}, traceID)

		resp, err := handler(ctx, req)

		latency := time.Since(start)

		statusCode := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code()
			} else {
				statusCode = codes.Unknown
			}
		}

		attributes := []slog.Attr{
			slog.String("method", info.FullMethod),
			slog.String("latency", latency.String()),
			slog.String("trace-id", traceID),
			slog.String("grpc-code", statusCode.String()),
		}
		if err != nil {
			attributes = append(attributes, slog.String("error", err.Error()))
			log.LogAttrs(ctx, slog.LevelError, "gRPC request failed", attributes...)
		} else {
			log.LogAttrs(ctx, slog.LevelInfo, "gRPC request completed", attributes...)
		}

		return resp, err
	}
}
