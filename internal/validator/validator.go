package validator

type Verifier interface {
	Verify() error
}
