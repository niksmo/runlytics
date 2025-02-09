package repeat

import (
	"context"
	"errors"
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"go.uber.org/zap"
)

func WithTries(logPrefix string, tries []time.Duration, fn func() error) {
	err := fn()

	if err == nil {
		return
	}

	for _, duration := range tries {
		logger.Log.Debug(logPrefix, zap.Duration("try_after", duration), zap.Error(err))
		time.Sleep(duration)
		err = fn()
		if err == nil || errors.Is(err, context.Canceled) {
			break
		}
	}
}
