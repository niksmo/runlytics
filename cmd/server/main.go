package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server/api"
	"github.com/niksmo/runlytics/internal/server/middleware"
	"github.com/niksmo/runlytics/internal/server/repository"
	"github.com/niksmo/runlytics/internal/server/router"
	"github.com/niksmo/runlytics/internal/server/service"
	"github.com/niksmo/runlytics/internal/storage"
	"go.uber.org/zap"
)

func main() {
	parseFlags()

	if err := logger.Initialize(flagLog); err != nil {
		panic(err)
	}

	logger.Log.Debug("Bootstrap server")

	storage := storage.NewMemStorage()
	mux := chi.NewRouter()
	mux.Use(middleware.Logger)

	repository := repository.New()
	HTMLService := service.NewHTMLService(repository)
	updateService := service.NewUpdateService(repository)

	api.SetHTMLHandler(mux, HTMLService)
	api.SetUpdateHandler(mux, updateService)

	// router.SetMainRoute(mux, storage)
	// router.SetUpdateRoute(mux, storage)
	router.SetValueRoute(mux, storage)

	server := http.Server{
		Addr:    flagAddr.String(),
		Handler: mux,
	}

	logger.Log.Info("Listen", zap.String("host", server.Addr))
	logger.Log.Info("Stop listening and serve", zap.Error(server.ListenAndServe()))
}
