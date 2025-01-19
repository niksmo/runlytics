package middleware

import (
	"net/http"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"go.uber.org/zap"
)

type responseInfo struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseInfo *responseInfo
}

func (lrw loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(b)
	lrw.responseInfo.size += size
	return size, err

}

func (lrw loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.ResponseWriter.WriteHeader(statusCode)
	lrw.responseInfo.status = statusCode
}

func Logger(next http.Handler) http.Handler {
	logFunc := func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := loggingResponseWriter{
			ResponseWriter: rw,
			responseInfo:   new(responseInfo),
		}

		next.ServeHTTP(lrw, r)

		logger.Log.Info(
			"Handle request",
			zap.String("path", r.URL.EscapedPath()),
			zap.String("method", r.Method),
			zap.Int("status", lrw.responseInfo.status),
			zap.Duration("duration", time.Since(start)),
			zap.Int("size", lrw.responseInfo.size),
		)
	}

	return http.HandlerFunc(logFunc)
}
