package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server/api"
	"github.com/niksmo/runlytics/internal/server/config"
	"github.com/niksmo/runlytics/internal/server/middleware"
	"github.com/niksmo/runlytics/internal/server/service"
	"github.com/niksmo/runlytics/internal/server/storage"
	"github.com/niksmo/runlytics/internal/server/validator"
	"github.com/niksmo/runlytics/pkg/httpserver"
	"github.com/niksmo/runlytics/pkg/sqldb"
	"go.uber.org/zap"
)

func main() {
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
	)

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux.Use(middleware.AllowContentEncoding("gzip"))
	mux.Use(middleware.Gzip)

	pgDB := sqldb.New("pgx", config.DatabaseDSN(), logger.Log.Sugar())
	repository := storage.New(pgDB, config)

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

	api.SetHealthCheckHandler(mux, service.NewHealthCheckService(pgDB))

	HTTPServer := httpserver.New(config.Addr(), mux, logger.Log.Sugar())

	var wg sync.WaitGroup
	interruptCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	repository.Run(interruptCtx, &wg)
	HTTPServer.Run(interruptCtx, &wg)
	wg.Wait()
}
