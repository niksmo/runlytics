package service

import (
	"context"

	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
)

type UpdateService struct {
	repository di.UpdateRepository
}

func NewUpdateService(repository di.UpdateRepository) *UpdateService {
	return &UpdateService{repository}
}

func (s *UpdateService) Update(
	ctx context.Context, scheme *metrics.Metrics,
) (err error) {
	if scheme == nil {
		return server.ErrInternal
	}

	switch scheme.MType {
	case metrics.MTypeGauge:
		return s.updateGauge(ctx, scheme)
	case metrics.MTypeCounter:
		return s.updateCounter(ctx, scheme)
	default:
		return server.ErrInternal
	}
}

func (s *UpdateService) updateGauge(
	ctx context.Context, scheme *metrics.Metrics,
) error {
	if scheme.Value == nil {
		return server.ErrInternal
	}
	value, err := s.repository.UpdateGaugeByName(
		ctx, scheme.ID, *scheme.Value,
	)
	if err != nil {
		return err
	}
	scheme.Value = &value
	return nil
}

func (s *UpdateService) updateCounter(
	ctx context.Context, scheme *metrics.Metrics,
) error {
	if scheme.Delta == nil {
		return server.ErrInternal
	}
	delta, err := s.repository.UpdateCounterByName(
		ctx, scheme.ID, *scheme.Delta,
	)
	if err != nil {
		return err
	}
	scheme.Delta = &delta
	return nil
}
