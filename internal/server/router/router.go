package router

import (
	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/server"
	"go.uber.org/zap"
)

const (
	gauge   = server.Gauge
	counter = server.Counter
)

func logRegister(endpoint string) {
	logger.Log.Debug("Register", zap.String("endpoint", endpoint))
}
