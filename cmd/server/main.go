package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server/api"
	"github.com/niksmo/runlytics/internal/server/middleware"
	"github.com/niksmo/runlytics/internal/server/repository"
	"github.com/niksmo/runlytics/internal/server/service"
	"go.uber.org/zap"
)

func main() {
	parseFlags()

	if err := logger.Initialize(flagLog); err != nil {
		panic(err)
	}

	logger.Log.Debug("Bootstrap server")

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)

	repository := repository.New()
	HTMLService := service.NewHTMLService(repository)
	updateService := service.NewUpdateService(repository)
	readService := service.NewReadService(repository)

	api.SetHTMLHandler(mux, HTMLService)
	api.SetUpdateHandler(mux, updateService)
	api.SetReadHandler(mux, readService)

	server := http.Server{
		Addr:    flagAddr.String(),
		Handler: mux,
	}

	logger.Log.Info("Listen", zap.String("host", server.Addr))
	logger.Log.Info("Stop server", zap.Error(server.ListenAndServe()))
}
