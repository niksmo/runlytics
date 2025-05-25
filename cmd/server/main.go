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
	logger.Init(config.LogLvl())

	logger.Log.Info(
		"Bootstrap server with flags",
		zap.String("ADDRESS", config.Addr()),
		zap.String("LOG_LVL", config.LogLvl()),
		zap.Float64("STORE_INTERVAL", config.SaveInterval().Seconds()),
		zap.String("FILE_STORAGE_PATH", config.FileName()),
		zap.Bool("RESTORE", config.Restore()),
		zap.String("DATABASE_DSN", config.DatabaseDSN()),
		zap.String("KEY", config.Key()),
		zap.String("CRYPTO_KEY", config.CryptoKeyPath()),
	)

	decrypter, err := cipher.NewDecrypter(config.CryptoKeyData())
	if err != nil {
		logger.Log.Fatal("failed to init decrypter", zap.Error(err))
	}

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux.Use(middleware.Decrypt(decrypter))
	mux.Use(middleware.AllowContentEncoding("gzip"))
	mux.Use(middleware.Gzip)
	mux.Use(middleware.VerifyAndWriteSHA256(config.Key(), http.MethodPost))

	pgDB := sqldb.New("pgx", config.DatabaseDSN(), logger.Log.Sugar())
	fileOperator := fileoperator.New(config.File())
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

	HTTPServer := httpserver.New(config.Addr(), mux, logger.Log.Sugar())

	var wg sync.WaitGroup
	stopCtx, stopFn := signal.NotifyContext(
		context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT,
	)
	defer stopFn()
	repository.Run(stopCtx, &wg)
	HTTPServer.Run(stopCtx, &wg)
	wg.Wait()
}
