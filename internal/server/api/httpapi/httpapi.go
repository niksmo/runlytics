// Package api provides REST API server handlers.
package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/pkg/di"
	"go.uber.org/zap"
)

// Headers
const (
	ContentType     = "Content-Type"
	ContentEncoding = "Content-Encoding"
	AcceptEncoding  = "Accept-Encoding"

	XRealIP = "X-Real-IP"
)

// Content types
const (
	JSON = "application/json"
	HTML = "text/html"
	TEXT = "text/plain"
)

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

// ReadRequest decode request body from JSON objects.
func ReadJSONRequest(r *http.Request, scheme any) error {
	if err := json.NewDecoder(r.Body).Decode(scheme); err != nil {
		return fmt.Errorf("decode request body error: %w", err)
	}
	return nil
}

// WriteResponse encode data and write response with passed code.
func WriteJSONResponse(w http.ResponseWriter, statusCode int, scheme any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(scheme); err != nil {
		return fmt.Errorf("encode response body error: %w", err)
	}
	return nil
}

func debugLogRegister(endpoint string) {
	logger.Log.Debug("Register", zap.String("endpoint", endpoint))
}
