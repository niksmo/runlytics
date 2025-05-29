package grpcapp

import (
	"fmt"
	"net"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server/api/grpcapi"
	"github.com/niksmo/runlytics/internal/server/app/grpc/interceptor"
	"github.com/niksmo/runlytics/pkg/di"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App struct {
	gRPCServer *grpc.Server
	addr       *net.TCPAddr
}

func New(batchUpdateService di.BatchUpdateService, addr *net.TCPAddr) *App {
	gRPCServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.WithRecovery(),
			interceptor.WithLog(),
		),
	)
	grpcapi.Register(gRPCServer, batchUpdateService)

	return &App{gRPCServer: gRPCServer, addr: addr}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	lis, err := net.ListenTCP("tcp", a.addr)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	logger.Log.Info(
		"gRPC server started",
		zap.String("op", op),
		zap.String("addr", a.addr.String()),
	)

	err = a.gRPCServer.Serve(lis)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	log := logger.Log.With(
		zap.String("op", op),
		zap.String("addr", a.addr.String()),
	)

	log.Info("gRPC server stopping gracefully")

	a.gRPCServer.GracefulStop()

	log.Info("grpc server stopped")
}
