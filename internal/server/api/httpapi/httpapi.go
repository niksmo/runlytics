// Package api provides REST API server handlers.
package httpapi

import (
	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/di"
	"go.uber.org/zap"
)

func debugLogRegister(endpoint string) {
	logger.Log.Debug("Register", zap.String("endpoint", endpoint))
}

type RegisterServices struct {
	di.HTMLService
	di.UpdateService
	di.BatchUpdateService
	di.ReadService
	di.HealthCheckService
}

func Register(mux *chi.Mux, s RegisterServices) {
	SetHTMLHandler(mux, s.HTMLService)
	SetUpdateHandler(mux, s.UpdateService)
	SetBatchUpdateHandler(mux, s.BatchUpdateService)
	SetValueHandler(mux, s.ReadService)
	SetHealthCheckHandler(mux, s.HealthCheckService)
}
