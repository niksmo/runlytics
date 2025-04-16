package service

import (
	"context"

	"github.com/niksmo/runlytics/internal/server/errs"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
)

type ReadService struct {
	repository di.ReadByNameRepository
}

func NewValueService(repository di.ReadByNameRepository) *ReadService {
	return &ReadService{repository}
}

func (s *ReadService) Read(
	ctx context.Context, m *metrics.Metrics,
) error {
	if m == nil {
		return errs.ErrInternal
	}

	switch m.MType {
	case metrics.MTypeGauge:
		return s.readGauge(ctx, m)
	case metrics.MTypeCounter:
		return s.readCounter(ctx, m)
	default:
		return errs.ErrInternal
	}
}

func (s *ReadService) readGauge(
	ctx context.Context, m *metrics.Metrics,
) error {
	v, err := s.repository.ReadGaugeByName(ctx, m.ID)
	if err != nil {
		return err
	}
	m.Value = &v
	return nil
}

func (s *ReadService) readCounter(
	ctx context.Context, m *metrics.Metrics,
) error {
	d, err := s.repository.ReadCounterByName(ctx, m.ID)
	if err != nil {
		return err
	}
	m.Delta = &d
	return nil
}
