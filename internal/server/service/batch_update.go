package service

import (
	"context"

	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
)

// BatchUpdateService works with repository and provides BatchUpdate method.
type BatchUpdateService struct {
	repository di.IBatchUpdateStorage
}

// NewBatchUpdateService returns BatchUpdateService pointer.
func NewBatchUpdateService(
	repository di.IBatchUpdateStorage,
) *BatchUpdateService {
	return &BatchUpdateService{repository}
}

// BatchUpdate accept slice of metrics and returns error if occured.
//
// Update metrics steps:
//  1. split metrics on two different slices: counter and gauge
//  2. update two slices in order: gauge -> counter
//
// If error occur on gauge update step, returns that error immediately.
//
// TODO(niksmo): update gauge and counter slices in separate goroutines.
func (s *BatchUpdateService) BatchUpdate(
	ctx context.Context, ml metrics.MetricsList,
) error {
	var gl []metrics.Metrics
	var cl []metrics.Metrics

	for _, m := range ml {
		switch m.MType {
		case metrics.MTypeGauge:
			if m.Value == nil {
				return server.ErrInternal
			}
			gl = append(gl, m)
		case metrics.MTypeCounter:
			if m.Delta == nil {
				return server.ErrInternal
			}
			cl = append(cl, m)
		default:
			return server.ErrInternal
		}
	}

	if len(gl) != 0 {
		err := s.repository.UpdateGaugeList(ctx, gl)
		if err != nil {
			return err
		}
	}

	if len(cl) != 0 {
		err := s.repository.UpdateCounterList(ctx, cl)
		if err != nil {
			return err
		}
	}

	return nil
}
