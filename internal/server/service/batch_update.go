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
	gaugeMap := make(map[string]float64)
	counterMap := make(map[string]int64)

	for _, item := range batch {
		switch item.MType {
		case metrics.MTypeGauge:
			gaugeMap[item.ID] = *item.Value
		case metrics.MTypeCounter:
			counterMap[item.ID] = *item.Delta
		default:
			return server.ErrInternal
		}
	}

	if len(gaugeMap) != 0 {
		err := service.repository.UpdateGaugeList(ctx, gaugeMap)
		if err != nil {
			return err
		}
	}

	if len(counterMap) != 0 {
		err := service.repository.UpdateCounterList(ctx, counterMap)
		if err != nil {
			return err
		}
	}

	return nil
}
