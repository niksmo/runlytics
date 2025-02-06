package validator

type ValueValidator struct{}

func NewValueValidator() *ValueValidator {
	return &ValueValidator{}
}

func (v *ValueValidator) VerifyScheme(object Verifier) error {
	return object.Verify()
}
