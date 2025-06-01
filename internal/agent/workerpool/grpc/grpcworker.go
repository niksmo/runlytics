package grpcworker

import (
	"context"

	"github.com/niksmo/runlytics/pkg/metrics"
)

func SendMetrics(ctx context.Context, m metrics.MetricsList) error {
	return nil
}
