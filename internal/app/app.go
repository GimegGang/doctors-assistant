package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"kode/docs/openAPI"
	"kode/internal/config"
	"kode/internal/service/medService"
	"kode/internal/transport/grpc/grpcServer"
	gRPCMiddleware "kode/internal/transport/grpc/middleware"
	"kode/internal/transport/rest/handlers"
	"kode/internal/transport/rest/restMiddleware"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	restServer *http.Server
	gRPCPort   int
}

func New(log *slog.Logger, config *config.Config, service *medService.MedService) *App {
	gRPCServer := grpc.NewServer(grpc.UnaryInterceptor(gRPCMiddleware.GRPCLogger(log)))
	grpcServer.Register(gRPCServer, service)

	if config.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(restMiddleware.RestLogger(log))
	router.Use(gin.Recovery())

	handler := handlers.New(log, service)
	openAPI.RegisterHandlers(router, handler)

	restServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.RestAddress),
		Handler:      router,
		ReadTimeout:  config.Timeout,
		WriteTimeout: config.Timeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		restServer: restServer,
		gRPCPort:   config.GrpcAddress,
	}
}

func (a *App) Start() error {
	const fun = "grpcapp.Run"

	go func() {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.gRPCPort))
		if err != nil {
			a.log.Error(fmt.Sprintf("%s: %v", fun, err))
			return
		}

		a.log.Info("gRPC server starting", slog.String("addr", l.Addr().String()))
		if err := a.gRPCServer.Serve(l); err != nil {
			a.log.Error(fmt.Sprintf("%s: %v", fun, err))
			return
		}
	}()

	a.log.Info("REST server starting", slog.String("addr", a.restServer.Addr))
	if err := a.restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("%s: %w", fun, err)
	}

	return nil
}

func (a *App) Stop() {
	const fun = "app.Stop"

	a.log.Info("stopping servers")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a.gRPCServer.GracefulStop()

	if err := a.restServer.Shutdown(ctx); err != nil {
		a.log.Error("failed to stop REST server", slog.String("fun", fun), slog.Any("err", err))
	}
}
