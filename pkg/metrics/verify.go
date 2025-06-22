package metrics

import (
	"errors"
	"strings"
)

var (
	allowedTypes = map[string]struct{}{
		MTypeCounter: {},
		MTypeGauge:   {},
	}
)

// Metrics field validation errors.
var (
	ErrIDRequired  = errors.New("'id': required")
	ErrInvalidType = errors.New("'type': ['gauge'|'counter']")
)

// VerifyErrors implements error interface, used in Metrics.Verify method.
type VerifyErrors []error

// Unwrap is builtin error interface method implementation.
func (ve VerifyErrors) Unwrap() []error {
	if len(ve) == 0 {
		return nil
	}
	return ve
}

// Error is builtin error interface method implementation.
func (ve VerifyErrors) Error() string {
	errStrings := make([]string, 0, len(ve))
	for _, err := range ve {
		if err != nil {
			errStrings = append(errStrings, err.Error())
		}
	}
	return strings.Join(errStrings, "; ")
}

// VerifyID performs validation under Metrics.ID field :
//   - If id is zero string, [ErrIDRequired] is occur.
func VerifyID(m Metrics) error {
	if strings.TrimSpace(m.ID) == "" {
		return ErrIDRequired
	}
	return nil
}

// VerifyType performs validation under Metrics.MType field:
//   - If type is not counter or gauge, [ErrInvalidType] is occur.
func VerifyType(m Metrics) error {
	if _, ok := allowedTypes[m.MType]; !ok {
		return ErrInvalidType
	}

	return nil
}
