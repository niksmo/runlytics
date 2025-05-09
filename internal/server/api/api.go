// Package api provides REST API server handlers.
package api

import (
	"github.com/niksmo/runlytics/internal/logger"
	"go.uber.org/zap"
)

func debugLogRegister(endpoint string) {
	logger.Log.Debug("Register", zap.String("endpoint", endpoint))
}
