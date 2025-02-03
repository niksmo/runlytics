package service

import (
	"fmt"

	"github.com/niksmo/runlytics/internal/metrics"
)

type ReadByNameRepository interface {
	ReadCounterByName(name string) (int64, error)
	ReadGaugeByName(name string) (float64, error)
}

type ReadService struct {
	repository ReadByNameRepository
}

func NewReadService(repository ReadByNameRepository) *ReadService {
	return &ReadService{repository}
}

func (service *ReadService) Read(mData *metrics.Metrics) error {
	switch mData.MType {
	case metrics.MTypeGauge:
		v, err := service.repository.ReadGaugeByName(mData.ID)
		if err != nil {
			return err
		}
		mData.Value = &v
	case metrics.MTypeCounter:
		v, err := service.repository.ReadCounterByName(mData.ID)
		if err != nil {
			return err
		}
		mData.Delta = &v
	default:
		return fmt.Errorf("wrong type value: '%s'. Expect 'counter' or 'gauge'", mData.MType)
	}

	return nil
}
