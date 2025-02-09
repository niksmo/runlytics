package repeater

import (
	"time"

	"github.com/niksmo/runlytics/internal/logger"
	"go.uber.org/zap"
)

type Repeater struct {
	Wait      []time.Duration
	Fn        func() error
	LogPrefix string
}

func New(logPrefix string, wait []time.Duration, fn func() error) *Repeater {
	return &Repeater{wait, fn, logPrefix}
}

func (r *Repeater) DoFn() {
	err := r.Fn()
	if err != nil {
		for _, duration := range r.Wait {
			logger.Log.Debug(r.LogPrefix, zap.Duration("try_after", duration), zap.Error(err))
			time.Sleep(duration)
			err = r.Fn()
			if err == nil {
				break
			}
		}
	}
}
