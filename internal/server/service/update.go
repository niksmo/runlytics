package service

import (
	"fmt"
	"strings"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/metrics"
	"go.uber.org/zap"
)

const emptyField = "field is empty"

type ErrMetricsField []error

func (e ErrMetricsField) Error() string {
	var s []string
	for _, err := range e {
		s = append(s, err.Error())
	}
	return strings.Join(s, "; ")
}

type UpdateRepository interface {
	UpdateCounterByName(name string, value int64) (int64, error)
	UpdateGaugeByName(name string, value float64) (float64, error)
}

type UpdateService struct {
	repository UpdateRepository
}

func NewUpdateService(repository UpdateRepository) *UpdateService {
	return &UpdateService{repository}
}

func (service *UpdateService) Update(mData *metrics.Metrics) error {
	var errs ErrMetricsField
	if mData.ID == "" {
		errs = append(errs, fmt.Errorf("'id' %s", emptyField))
	}

	//should check empty MType too

	switch mData.MType {
	case metrics.MTypeGauge:
		if mData.Value == nil {
			errs = append(errs, fmt.Errorf("'value' %s", emptyField))
		} else {
			v := service.repository.UpdateGaugeByName(
				mData.ID,
				*mData.Value,
			)
			mData.Value = &v
		}
	case metrics.MTypeCounter:
		if mData.Delta == nil {
			errs = append(errs, fmt.Errorf("'delta' %s", emptyField))
		} else {
			v := service.repository.UpdateCounterByName(
				mData.ID,
				*mData.Delta,
			)
			mData.Delta = &v
		}
	default:
		errs = append(
			errs,
			fmt.Errorf("wrong type value: '%s'. Expect 'counter' or 'gauge'", mData.MType),
		)
	}

	if len(errs) != 0 {
		logger.Log.Debug(
			"The metrics have not been updated",
			zap.String("metricsID", mData.ID),
			zap.Error(errs),
		)
		return errs
	}

	return nil
}
