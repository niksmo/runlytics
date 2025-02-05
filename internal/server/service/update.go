package service

import (
	"context"

	"github.com/niksmo/runlytics/internal/metrics"
	"github.com/niksmo/runlytics/internal/server"
)

type UpdaterRepository interface {
	UpdateCounterByName(ctx context.Context, name string, value int64) (int64, error)
	UpdateGaugeByName(ctx context.Context, name string, value float64) (float64, error)
}

type UpdateService struct {
	repository UpdaterRepository
}

func NewUpdateService(repository UpdaterRepository) *UpdateService {
	return &UpdateService{repository}
}

func (s *UpdateService) Update(
	ctx context.Context, scheme *metrics.MetricsUpdate,
) (metrics.Metrics, error) {
	switch scheme.MType {
	case metrics.MTypeGauge:
		value, err := s.repository.UpdateGaugeByName(ctx, scheme.ID, *scheme.Value)
		if err != nil {
			return nil, err
		}
		mGauge := metrics.MetricsGauge{
			ID: scheme.ID, MType: scheme.MType, Value: value,
		}
		return mGauge, nil

	case metrics.MTypeCounter:
		delta, err := s.repository.UpdateCounterByName(ctx, scheme.ID, *scheme.Delta)
		if err != nil {
			return nil, err
		}
		mCounter := metrics.MetricsCounter{
			ID: scheme.ID, MType: scheme.MType, Delta: delta,
		}
		return mCounter, nil
	}
	return nil, server.ErrInternal
}
