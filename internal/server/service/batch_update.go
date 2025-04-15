package service

import (
	"context"

	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
)

type BatchUpdateService struct {
	repository di.UpdateListRepository
}

func NewBatchUpdateService(
	repository di.UpdateListRepository,
) *BatchUpdateService {
	return &BatchUpdateService{repository}
}

func (service *BatchUpdateService) BatchUpdate(
	ctx context.Context, batch metrics.MetricsBatchUpdate,
) error {
	var gaugeSlice []metrics.Metrics
	var counterSlice []metrics.Metrics

	for _, metricsUpdate := range batch {
		switch metricsUpdate.MType {
		case metrics.MTypeGauge:
			if metricsUpdate.Value == nil {
				return server.ErrInternal
			}
			gaugeSlice = append(gaugeSlice, metricsUpdate.Metrics)
		case metrics.MTypeCounter:
			if metricsUpdate.Delta == nil {
				return server.ErrInternal
			}
			counterSlice = append(counterSlice, metricsUpdate.Metrics)
		default:
			return server.ErrInternal
		}
	}

	if len(gaugeSlice) != 0 {
		err := service.repository.UpdateGaugeList(ctx, gaugeSlice)
		if err != nil {
			return err
		}
	}

	if len(counterSlice) != 0 {
		err := service.repository.UpdateCounterList(ctx, counterSlice)
		if err != nil {
			return err
		}
	}

	return nil
}
