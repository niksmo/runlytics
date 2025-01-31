package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server/api"
	"github.com/niksmo/runlytics/internal/server/db"
	"github.com/niksmo/runlytics/internal/server/middleware"
	"github.com/niksmo/runlytics/internal/server/repository"
	"github.com/niksmo/runlytics/internal/server/service"
	"github.com/niksmo/runlytics/internal/server/storage"
	"go.uber.org/zap"
)

func main() {
	parseFlags()

	if err := logger.Initialize(flagLog); err != nil {
		panic(err)
	}

	logger.Log.Info(
		"Bootstrap server with flags",
		zap.String("ADDRESS", flagAddr.String()),
		zap.String("LOG_LVL", flagLog),
		zap.Float64("STORE_INTERVAL", flagInterval.Seconds()),
		zap.String("FILE_STORAGE_PATH", flagStoragePath.Name()),
		zap.Bool("RESTORE", flagRestore),
		zap.String("DATABASE_DSN", flagDSN),
	)

	db := db.Init(flagDSN)
	defer db.Close()

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux.Use(middleware.AllowContentEncoding("gzip"))
	mux.Use(middleware.Gzip)

	repository := repository.New()
	fileStorage := storage.NewFileStorage(
		repository,
		flagInterval,
		flagStoragePath,
		flagRestore,
	)
	HTMLService := service.NewHTMLService(repository)
	updateService := service.NewUpdateService(fileStorage)
	readService := service.NewReadService(repository)
	healthCheckService := service.NewHealthCheckService(db)

	api.SetHTMLHandler(mux, HTMLService)
	api.SetUpdateHandler(mux, updateService)
	api.SetReadHandler(mux, readService)
	api.SetHealthCheckHandler(mux, healthCheckService)

	fileStorage.Run()

	server := http.Server{
		Addr:    flagAddr.String(),
		Handler: mux,
	}

	logger.Log.Info("Listen", zap.String("host", server.Addr))
	logger.Log.Info("Stop server", zap.Error(server.ListenAndServe()))
}
