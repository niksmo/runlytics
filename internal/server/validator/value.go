package validator

import "github.com/niksmo/runlytics/pkg/di"

type ValueValidator struct{}

func NewValueValidator() *ValueValidator {
	return &ValueValidator{}
}

func (v *ValueValidator) VerifyScheme(object di.Verifier) error {
	return object.Verify()
}
