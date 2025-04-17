// Package api provides REST API handlers for server.
package api

import (
	"github.com/niksmo/runlytics/internal/logger"
	"go.uber.org/zap"
)

func debugLogRegister(endpoint string) {
	logger.Log.Debug("Register", zap.String("endpoint", endpoint))
}
