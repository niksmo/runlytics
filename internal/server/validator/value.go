package validator

import (
	"github.com/niksmo/runlytics/pkg/di"
	"github.com/niksmo/runlytics/pkg/metrics"
)

type ValueValidator struct{}

func NewValueValidator() *ValueValidator {
	return &ValueValidator{}
}

func (v *ValueValidator) VerifyScheme(object di.Verifier) error {
	return object.Verify()
}

func (v *ValueValidator) VerifyParams(
	id, mType, _ string,
) (metrics.Metrics, error) {
	var scheme metrics.MetricsRead
	scheme.ID = id
	scheme.MType = mType
	if err := scheme.Verify(); err != nil {
		return scheme.Metrics, err
	}

	return scheme.Metrics, nil
}
