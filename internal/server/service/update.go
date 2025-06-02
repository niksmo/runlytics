package service

import (
	"context"

	"github.com/niksmo/runlytics/internal/server"
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
)

// UpdateService works with repository and provides Update method.
type UpdateService struct {
	repository di.IUpdateByNameStorage
}

// NewUpdateService returns UpdateService pointer.
func NewUpdateService(repository di.IUpdateByNameStorage) *UpdateService {
	return &UpdateService{repository}
}

// Update defines metrics type and updates value. Returns error if occured.
func (s *UpdateService) Update(
	ctx context.Context, m *metrics.Metrics,
) error {
	if m == nil {
		return server.ErrInternal
	}

	switch m.MType {
	case metrics.MTypeGauge:
		return s.updateGauge(ctx, m)
	case metrics.MTypeCounter:
		return s.updateCounter(ctx, m)
	default:
		return server.ErrInternal
	}
}

func (s *UpdateService) updateGauge(
	ctx context.Context, m *metrics.Metrics,
) error {
	v, err := s.repository.UpdateGaugeByName(
		ctx, m.ID, m.Value,
	)
	if err != nil {
		return err
	}
	m.Value = v
	return nil
}

func (s *UpdateService) updateCounter(
	ctx context.Context, m *metrics.Metrics,
) error {

	d, err := s.repository.UpdateCounterByName(
		ctx, m.ID, m.Delta,
	)
	if err != nil {
		return err
	}
	m.Delta = d
	return nil
}
