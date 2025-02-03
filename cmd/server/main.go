package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server/api"
	"github.com/niksmo/runlytics/internal/server/config"
	"github.com/niksmo/runlytics/internal/server/db"
	"github.com/niksmo/runlytics/internal/server/middleware"
	"github.com/niksmo/runlytics/internal/server/service"
	"github.com/niksmo/runlytics/internal/server/storage"
	"go.uber.org/zap"
)

func main() {
	config := config.Load()

	if err := logger.Init(config.LogLvl()); err != nil {
		panic(err)
	}

	logger.Log.Info(
		"Bootstrap server with flags",
		zap.String("ADDRESS", config.Addr()),
		zap.String("LOG_LVL", config.LogLvl()),
		zap.Float64("STORE_INTERVAL", config.SaveInterval().Seconds()),
		zap.String("FILE_STORAGE_PATH", config.File().Name()),
		zap.Bool("RESTORE", config.Restore()),
		zap.String("DATABASE_DSN", config.DatabaseDSN()),
	)

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux.Use(middleware.AllowContentEncoding("gzip"))
	mux.Use(middleware.Gzip)

	db := db.Init(config.DatabaseDSN())
	defer db.Close()

	storage := storage.NewMemory(config.File(), config.SaveInterval(), config.Restore())
	storage.Run()

	HTMLService := service.NewHTMLService(storage)
	updateService := service.NewUpdateService(storage)
	readService := service.NewReadService(storage)
	healthCheckService := service.NewHealthCheckService(db)

	api.SetHTMLHandler(mux, HTMLService)
	api.SetUpdateHandler(mux, updateService)
	api.SetReadHandler(mux, readService)
	api.SetHealthCheckHandler(mux, healthCheckService)

	server := http.Server{
		Addr:    config.Addr(),
		Handler: mux,
	}

	logger.Log.Info("Listen", zap.String("host", server.Addr))
	logger.Log.Info("Stop server", zap.Error(server.ListenAndServe()))
}
