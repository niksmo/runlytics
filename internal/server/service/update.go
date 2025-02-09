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
	ctx context.Context, scheme *metrics.MetricsUpdate,
) (di.Metrics, error) {
	if scheme == nil {
		return nil, server.ErrInternal
	}

	switch scheme.MType {
	case metrics.MTypeGauge:
		return s.updateGauge(ctx, scheme)
	case metrics.MTypeCounter:
		return s.updateCounter(ctx, scheme)
	}
	return nil, server.ErrInternal
}

func (s *UpdateService) updateGauge(
	ctx context.Context, scheme *metrics.MetricsUpdate,
) (metrics.MetricsGauge, error) {
	if scheme.Value == nil {
		return metrics.MetricsGauge{}, server.ErrInternal
	}
	value, err := s.repository.UpdateGaugeByName(ctx, scheme.ID, *scheme.Value)
	if err != nil {
		return metrics.MetricsGauge{}, err
	}
	mGauge := metrics.MetricsGauge{
		ID: scheme.ID, MType: scheme.MType, Value: value,
	}
	return mGauge, nil
}

func (s *UpdateService) updateCounter(
	ctx context.Context, scheme *metrics.MetricsUpdate,
) (metrics.MetricsCounter, error) {
	if scheme.Delta == nil {
		return metrics.MetricsCounter{}, server.ErrInternal
	}
	delta, err := s.repository.UpdateCounterByName(ctx, scheme.ID, *scheme.Delta)
	if err != nil {
		return metrics.MetricsCounter{}, err
	}
	mCounter := metrics.MetricsCounter{
		ID: scheme.ID, MType: scheme.MType, Delta: delta,
	}
	return mCounter, nil
}
