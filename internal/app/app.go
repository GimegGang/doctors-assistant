package app

import (
	"fmt"
	"google.golang.org/grpc"
	"kode/internal/grpc/medicine"
	"kode/internal/service/medService"
	"kode/internal/storage/sqlite"
	"log/slog"
	"net"
	"time"
)

type App struct {
	log        *slog.Logger
	period     time.Duration
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, period time.Duration, port int, storage *sqlite.Storage) *App {
	gRPCServer := grpc.NewServer()

	service := medService.New(log, storage, period)

	medicine.Register(gRPCServer, service)

	return &App{
		log:        log,
		period:     period,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) Start() error {
	const fun = "grpcapp.Run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", fun, err)
	}
	a.log.With(slog.String("op", fun)).Info("starting gRPC server", slog.String("address", l.Addr().String()))
	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", fun, err)
	}
	return nil
}

func (a *App) Stop() {
	const fun = "grpcapp.Stop"

	a.log.With(slog.String("op", fun)).Info("stopping gRPC server")
	a.gRPCServer.GracefulStop()
}
