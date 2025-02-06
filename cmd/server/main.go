package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server/api"
	"github.com/niksmo/runlytics/internal/server/config"
	"github.com/niksmo/runlytics/internal/server/middleware"
	"github.com/niksmo/runlytics/internal/server/service"
	"github.com/niksmo/runlytics/internal/server/storage"
	"github.com/niksmo/runlytics/internal/server/validator"
	"go.uber.org/zap"
)

func main() {
	var wg sync.WaitGroup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	config := config.Load()

	if err := logger.Init(config.LogLvl()); err != nil {
		panic(err)
	}

	logger.Log.Info(
		"Bootstrap server with flags",
		zap.String("ADDRESS", config.Addr()),
		zap.String("LOG_LVL", config.LogLvl()),
		zap.Float64("STORE_INTERVAL", config.SaveInterval().Seconds()),
		zap.String("FILE_STORAGE_PATH", config.FileName()),
		zap.Bool("RESTORE", config.Restore()),
		zap.String("DATABASE_DSN", config.DatabaseDSN()),
	)

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux.Use(middleware.AllowContentEncoding("gzip"))
	mux.Use(middleware.Gzip)

	var repository storage.Storage
	if config.IsDatabase() {
		repository = storage.NewPSQL(config.DatabaseDSN())
	} else {
		repository = storage.NewMemory(
			config.File(),
			config.SaveInterval(),
			config.Restore(),
		)
	}
	repository.Run(ctx, &wg)

	api.SetHTMLHandler(mux, service.NewHTMLService(repository))

	api.SetUpdateHandler(
		mux,
		service.NewUpdateService(repository),
		validator.NewUpdateValidator(),
	)

	api.SetValueHandler(
		mux,
		service.NewValueService(repository),
		validator.NewValueValidator(),
	)

	api.SetHealthCheckHandler(mux, service.NewHealthCheckService(repository))

	server := http.Server{
		Addr:    config.Addr(),
		Handler: mux,
	}
	logger.Log.Info("Listen", zap.String("host", server.Addr))

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Log.Error("Server shutdown", zap.Error(err))
		}
	}()

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Log.Error("Server closed with errors", zap.Error(err))
	}

	wg.Wait()
}
