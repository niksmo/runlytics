package service

import (
	"context"

	"github.com/niksmo/runlytics/internal/metrics"
	"github.com/niksmo/runlytics/internal/server"
)

type ReaderByNameRepository interface {
	ReadCounterByName(ctx context.Context, name string) (int64, error)
	ReadGaugeByName(ctx context.Context, name string) (float64, error)
}

type ReadService struct {
	repository ReaderByNameRepository
}

func NewValueService(repository ReaderByNameRepository) *ReadService {
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
