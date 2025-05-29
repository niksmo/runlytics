package app

import (
	"github.com/niksmo/runlytics/internal/logger"
	grpcapp "github.com/niksmo/runlytics/internal/server/app/grpc"
	httpapp "github.com/niksmo/runlytics/internal/server/app/http"
	"github.com/niksmo/runlytics/internal/server/config"
	"github.com/niksmo/runlytics/internal/server/service"
	"github.com/niksmo/runlytics/internal/server/storage"
	"github.com/niksmo/runlytics/pkg/fileoperator"
	"github.com/niksmo/runlytics/pkg/sqldb"
	"go.uber.org/zap"
)

type App struct {
	GRPCServer *grpcapp.App
	HTTPServer *httpapp.App
}

func New(cfg *config.ServerConfig) *App {
	pgDB := sqldb.New("pgx", cfg.DB.DSN, func(err error) {
		logger.Log.Warn("failed connect to database", zap.Error(err))
	})
	fileOperator := fileoperator.New(cfg.FileStorage.File)
	repository := storage.New(pgDB, fileOperator, cfg)

	batchUpdateService := service.NewBatchUpdateService(repository)
	gRPCApp := grpcapp.New(batchUpdateService, cfg.GRPCAddr.TCPAddr)
	return &App{GRPCServer: gRPCApp}
}
