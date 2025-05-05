package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

type traceIDKey struct{}

func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path += "?" + c.Request.URL.RawQuery
		}

		traceID := c.GetHeader("X-TRACE-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		c.Header("X-TRACE-ID", traceID)

		ctx := context.WithValue(c.Request.Context(), traceIDKey{}, traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		latency := time.Now().Sub(start)

		attributes := []slog.Attr{
			slog.Int("status", c.Writer.Status()),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("ip", c.ClientIP()),
			slog.String("latency", latency.String()),
			slog.String("user-agent", c.Request.UserAgent()),
			slog.String("trace-id", traceID),
		}

		log.LogAttrs(c, slog.LevelInfo, "HTTP response", attributes...)
	}
}

func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey{}).(string); ok {
		return id
	}
	return ""
}

//TODO сделать логирование для gRPC
