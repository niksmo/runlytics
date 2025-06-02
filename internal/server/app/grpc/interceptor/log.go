package interceptor

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/niksmo/runlytics/internal/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

func WithLog() grpc.UnaryServerInterceptor {
	return logging.UnaryServerInterceptor(InterceptorLogger(logger.Log))
}

func InterceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		logger.Log.Sugar().Logw(zapcore.Level(lvl), msg, fields...)
	})
}
