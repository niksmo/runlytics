package service

import (
	"fmt"

	"github.com/niksmo/runlytics/internal/schemas"
	"github.com/niksmo/runlytics/internal/server"
)

type ReadService struct {
	repository ReadByNameRepository
}

type ReadByNameRepository interface {
	ReadCounterByName(name string) (int64, error)
	ReadGaugeByName(name string) (float64, error)
}

func NewReadService(repository ReadByNameRepository) *ReadService {
	return &ReadService{repository}
}

func (service *ReadService) Read(metrics *schemas.Metrics) error {
	switch metrics.MType {
	case server.MTypeGauge:
		v, err := service.repository.ReadGaugeByName(metrics.ID)
		if err != nil {
			return err
		}
		metrics.Value = &v
	case server.MTypeCounter:
		vInt, err := service.repository.ReadCounterByName(metrics.ID)
		if err != nil {
			return err
		}
		vFloat := float64(vInt)
		metrics.Value = &vFloat
	default:
		return fmt.Errorf("wrong type value: '%s'. Expect 'counter' or 'gauge'", metrics.MType)
	}

	return nil
}
