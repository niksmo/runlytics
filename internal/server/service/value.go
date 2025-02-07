package service

import (
	"context"

	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
)

type ReadService struct {
	repository di.ReadByNameRepository
}

func NewValueService(repository di.ReadByNameRepository) *ReadService {
	return &ReadService{repository}
}

func (service *ReadService) Read(
	ctx context.Context, scheme *metrics.MetricsRead,
) (metrics.Metrics, error) {
	switch scheme.MType {
	case metrics.MTypeGauge:
		value, err := service.repository.ReadGaugeByName(ctx, scheme.ID)
		if err != nil {
			return nil, err
		}
		mGauge := metrics.MetricsGauge{
			ID: scheme.ID, MType: scheme.MType, Value: value,
		}
		return mGauge, nil

	case metrics.MTypeCounter:
		delta, err := service.repository.ReadCounterByName(ctx, scheme.ID)
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
