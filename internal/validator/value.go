package validator

import "github.com/niksmo/runlytics/internal/metrics"

type ValueValidator struct{}

func NewValueValidator() *ValueValidator {
	return &ValueValidator{}
}

func (v *ValueValidator) VerifyScheme(s *metrics.MetricsRead) error {
	return s.Verify()
}
