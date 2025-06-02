package grpcworker

import (
	"context"

	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
)

func SendMetrics(
	ctx context.Context,
	m metrics.MetricsList,
	enc di.Encrypter,
	url, hk, ip string,
) error {
	return nil
}
