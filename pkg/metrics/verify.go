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
	ErrIDRequired    = errors.New("'id': required")
	ErrDeltaRequired = errors.New("'delta': expected int64")
	ErrDeltaLessZero = errors.New("'delta': less then 0")
	ErrValueRequired = errors.New("'value': expected float64")
	ErrInvalidType   = errors.New("'type': ['gauge'|'counter']")
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

// VerifyDelta performs validation under Metrics.Delta field:
//   - If Metrics.MType not a counter, interrupt validation algorithm and return nil.
//   - If delta is nil, return [ErrDeltaRequired].
//   - If delta less then zero, return [ErrDeltaLessZero].
func VerifyDelta(m Metrics) error {
	if m.MType != MTypeCounter {
		return nil
	}

	if m.Delta == nil {
		return ErrDeltaRequired
	}

	if *m.Delta < 0 {
		return ErrDeltaLessZero
	}

	return nil
}

// VerifyValue performs validation under Metrics.Value field:
//   - If Metrics.MType not a gauge, interrupt validation algorithm and return nil.
//   - If value is nil, return [ErrValueRequired].
func VerifyValue(m Metrics) error {
	if m.MType != MTypeGauge {
		return nil
	}

	if m.Value == nil {
		return ErrValueRequired
	}

	return nil
}
