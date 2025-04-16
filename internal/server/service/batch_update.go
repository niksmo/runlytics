package service

import (
	"context"

	"github.com/niksmo/runlytics/internal/server/errs"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
)

type BatchUpdateService struct {
	repository di.BatchUpdate
}

func NewBatchUpdateService(
	repository di.BatchUpdate,
) *BatchUpdateService {
	return &BatchUpdateService{repository}
}

func (s *BatchUpdateService) BatchUpdate(
	ctx context.Context, ml metrics.MetricsList,
) error {
	var gl []metrics.Metrics
	var cl []metrics.Metrics

	for _, m := range ml {
		switch m.MType {
		case metrics.MTypeGauge:
			if m.Value == nil {
				return errs.ErrInternal
			}
			gl = append(gl, m)
		case metrics.MTypeCounter:
			if m.Delta == nil {
				return errs.ErrInternal
			}
			cl = append(cl, m)
		default:
			return errs.ErrInternal
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
