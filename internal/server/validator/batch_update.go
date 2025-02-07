package validator

import "github.com/niksmo/runlytics/pkg/di"

type BatchUpdateValidator struct{}

func NewBatchUpdateValidator() *BatchUpdateValidator {
	return &BatchUpdateValidator{}
}

func (v *BatchUpdateValidator) VerifyScheme(object di.Verifier) error {
	return object.Verify()
}
