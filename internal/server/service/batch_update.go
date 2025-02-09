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

func NewBatchUpdateService(repository di.UpdateListRepository) *BatchUpdateService {
	return &BatchUpdateService{repository}
}

func (service *BatchUpdateService) BatchUpdate(ctx context.Context, batch metrics.MetricsBatchUpdate) error {
	var gaugeSlice []metrics.MetricsGauge
	var counterSlice []metrics.MetricsCounter

	for _, item := range batch {
		switch item.MType {
		case metrics.MTypeGauge:
			if item.Value == nil {
				return server.ErrInternal
			}
			gaugeSlice = append(gaugeSlice, metrics.MetricsGauge{
				ID:    item.ID,
				MType: item.MType,
				Value: *item.Value,
			})
		case metrics.MTypeCounter:
			if item.Delta == nil {
				return server.ErrInternal
			}
			counterSlice = append(counterSlice, metrics.MetricsCounter{
				ID:    item.ID,
				MType: item.MType,
				Delta: *item.Delta,
			})
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
