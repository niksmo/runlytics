package service

import (
	"fmt"
	"strings"

	"github.com/niksmo/runlytics/internal/logger"
	"github.com/niksmo/runlytics/internal/schemas"
	"go.uber.org/zap"
)

const emptyFieldStatus = "field is empty"

type ErrMetricsField []error

func (e ErrMetricsField) Error() string {
	var s []string
	for _, err := range e {
		s = append(s, err.Error())
	}
	return strings.Join(s, "; ")
}

type UpdateService struct {
	repository UpdateRepository
}

type UpdateRepository interface {
	UpdateCounterByName(name string, value int64) int64
	UpdateGaugeByName(name string, value float64) float64
}

func NewUpdateService(repository UpdateRepository) *UpdateService {
	return &UpdateService{repository}
}

func (service *UpdateService) Update(metrics *schemas.Metrics) error {
	var errs ErrMetricsField
	if metrics.ID == "" {
		errs = append(errs, fmt.Errorf("'id' %s", emptyFieldStatus))
	}

	switch metrics.MType {
	case "gauge":
		if metrics.Value == nil {
			errs = append(errs, fmt.Errorf("'value' %s", emptyFieldStatus))
		} else {
			v := service.repository.UpdateGaugeByName(
				metrics.ID,
				*metrics.Value,
			)
			metrics.Value = &v
		}
	case "counter":
		if metrics.Delta == nil {
			errs = append(errs, fmt.Errorf("'delta' %s", emptyFieldStatus))
		} else {
			v := float64(service.repository.UpdateCounterByName(
				metrics.ID,
				*metrics.Delta,
			))
			metrics.Value = &v
		}
	default:
		errs = append(
			errs,
			fmt.Errorf("wrong type value: '%s'. Expect 'counter' or 'gauge'", metrics.MType),
		)
	}

	if len(errs) != 0 {
		logger.Log.Debug(
			"The metrics have not been updated",
			zap.Error(errs),
		)
		return errs
	}

	return nil
}
