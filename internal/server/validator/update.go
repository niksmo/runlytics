package validator

import (
	"strconv"

	"github.com/niksmo/runlytics/internal/metrics"
)

type UpdateValidator struct{}

func NewUpdateValidator() *UpdateValidator {
	return &UpdateValidator{}
}

func (v *UpdateValidator) VerifyScheme(object Verifier) error {
	return object.Verify()
}

func (v *UpdateValidator) VerifyParams(
	id, mType, value string,
) (*metrics.MetricsUpdate, error) {
	scheme := &metrics.MetricsUpdate{ID: id, MType: mType}
	switch scheme.MType {
	case metrics.MTypeCounter:
		delta, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			scheme.Delta = &delta
		}
	case metrics.MTypeGauge:
		value, err := strconv.ParseFloat(value, 64)
		if err == nil {
			scheme.Value = &value
		}
	}

	if err := scheme.Verify(); err != nil {
		return nil, err
	}

	return scheme, nil
}
