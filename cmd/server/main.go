package main

import (
	"context"
	"net/http"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/niksmo/runlytics/internal/buildinfo"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server/api"
	"github.com/niksmo/runlytics/internal/server/config"
	"github.com/niksmo/runlytics/internal/server/middleware"
	"github.com/niksmo/runlytics/internal/server/service"
	"github.com/niksmo/runlytics/internal/server/storage"
	"github.com/niksmo/runlytics/pkg/cipher"
	"github.com/niksmo/runlytics/pkg/fileoperator"
	"github.com/niksmo/runlytics/pkg/httpserver"
	"github.com/niksmo/runlytics/pkg/sqldb"
)

func main() {
	buildinfo.Print()
	config := config.Load()
	logger.Init(config.Log.Level)

	printServerConfig(config, logger.Log)

	mux := chi.NewRouter()
	setupMiddlewares(mux, config)

	pgDB := sqldb.New("pgx", config.DB.DSN, func(err error) {
		logger.Log.Warn("failed to connect to database", zap.Error(err))
	})
	fileOperator := fileoperator.New(config.FileStorage.File)
	repository := storage.New(pgDB, fileOperator, config)

	api.SetHTMLHandler(mux, service.NewHTMLService(repository))

	api.SetUpdateHandler(
		mux,
		service.NewUpdateService(repository),
	)

	api.SetBatchUpdateHandler(mux,
		service.NewBatchUpdateService(repository),
	)

	api.SetValueHandler(
		mux,
		service.NewValueService(repository),
	)

	api.SetHealthCheckHandler(mux, service.NewHealthCheckService(pgDB))

	HTTPServer := httpserver.New(config.HTTPAddr.TCPAddr.String(), mux)

	stopCtx, stopFn := signal.NotifyContext(
		context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT,
	)
	defer stopFn()
	var wg sync.WaitGroup
	repository.Run(stopCtx, &wg)
	HTTPServer.Run(stopCtx, &wg)
	wg.Wait()
}

func printServerConfig(config *config.ServerConfig, logger *zap.Logger) {
	logger.Info(
		"Bootstrap server with flags",
		zap.String("ADDRESS", config.HTTPAddr.TCPAddr.String()),
		zap.String("LOG_LVL", config.Log.Level),
		zap.Float64(
			"STORE_INTERVAL", config.FileStorage.SaveInterval.Seconds(),
		),
		zap.String("FILE_STORAGE_PATH", config.FileStorage.FileName()),
		zap.Bool("RESTORE", config.FileStorage.Restore),
		zap.String("DATABASE_DSN", config.DB.DSN),
		zap.String("KEY", config.HashKey.Key),
		zap.String("CRYPTO_KEY", config.Crypto.Path),
		zap.String("TRUSTED_SUBNET", config.TrustedNet.IPNet.String()),
	)
}

func setupMiddlewares(mux *chi.Mux, config *config.ServerConfig) {
	decrypter, err := cipher.NewDecrypterX509(config.Crypto.Data)
	if err != nil {
		logger.Log.Fatal("failed to init decrypter", zap.Error(err))
	}

	mux.Use(middleware.Logger)
	mux.Use(middleware.Decrypt(decrypter))
	mux.Use(middleware.AllowContentEncoding("gzip"))
	mux.Use(middleware.Gzip)

	if config.HashKey.IsSet() {
		mux.Use(
			middleware.VerifyAndWriteSHA256(config.HashKey.Key, http.MethodPost),
		)
	}

	if config.TrustedNet.IsSet() {
		mux.Use(middleware.TrustedNet(config.TrustedNet.IPNet))
	}
}
