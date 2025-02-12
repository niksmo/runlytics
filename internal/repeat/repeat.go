package repeat

import (
	"context"
	"errors"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"go.uber.org/zap"
)

func WithTries(logPrefix string, waitIntervals []time.Duration, fn func() error) {
	err := fn()

	if err == nil {
		return
	}

	for _, interval := range waitIntervals {
		logger.Log.Debug(logPrefix, zap.Duration("try_after", interval), zap.Error(err))
		time.Sleep(interval)
		err = fn()
		if err == nil || errors.Is(err, context.Canceled) {
			break
		}
	}
}
