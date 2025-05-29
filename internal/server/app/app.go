package app

import (
	"github.com/niksmo/runlytics/internal/logger"
	grpcapp "github.com/niksmo/runlytics/internal/server/app/grpc"
	httpapp "github.com/niksmo/runlytics/internal/server/app/http"
	"github.com/niksmo/runlytics/internal/server/config"
	"github.com/niksmo/runlytics/internal/server/service"
	"github.com/niksmo/runlytics/internal/server/storage"
	"github.com/niksmo/runlytics/pkg/cipher"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/fileoperator"
	"go.uber.org/zap"
)

type App struct {
	GRPCServer *grpcapp.App
	HTTPServer *httpapp.App
	Storage    di.Storage
}

func New(cfg *config.ServerConfig) *App {
	decrypter, err := cipher.NewDecrypterX509(cfg.Crypto.Data)
	if err != nil {
		logger.Log.Fatal("failed to init decrypter", zap.Error(err))
	}

	fileOperator := fileoperator.New(cfg.FileStorage.File)
	storage := storage.New(
		fileOperator,
		cfg.DB.DSN,
		cfg.FileStorage.SaveInterval,
		cfg.FileStorage.Restore,
	)

	htmlS := service.NewHTMLService(storage)
	updateS := service.NewUpdateService(storage)
	readS := service.NewReadService(storage)
	healthCheckS := service.NewHealthCheckService(storage)
	batchUpdateService := service.NewBatchUpdateService(storage)

	gRPCApp := grpcapp.New(
		grpcapp.AppParams{
			BatchUpdateService: batchUpdateService,
			Addr:               cfg.GRPCAddr.TCPAddr,
			Decrypter:          decrypter,
			HashKey:            cfg.HashKey.Key,
			TrustedNed:         cfg.TrustedNet.IPNet,
		},
	)

	httpApp := httpapp.New(
		httpapp.AppParams{
			HTMLService:        htmlS,
			UpdateService:      updateS,
			ReadService:        readS,
			HealthCheckService: healthCheckS,
			BatchUpdateService: batchUpdateService,
			Addr:               cfg.HTTPAddr.TCPAddr,
			Decrypter:          decrypter,
			HashKey:            cfg.HashKey.Key,
			TrustedNet:         cfg.TrustedNet.IPNet,
		},
	)

	return &App{
		GRPCServer: gRPCApp,
		HTTPServer: httpApp,
		Storage:    storage,
	}
}
