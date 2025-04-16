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

var (
	ErrIDRequired    = errors.New("'id': required")
	ErrDeltaRequired = errors.New("'delta': expected int64")
	ErrDeltaLessZero = errors.New("'delta': less then 0")
	ErrValueRequired = errors.New("'value': expected float64")
	ErrInvalidType   = errors.New("'type': ['gauge'|'counter']")
)

type VerifyErrors []error

func (ve VerifyErrors) Unwrap() []error {
	if len(ve) == 0 {
		return nil
	}
	return ve
}

func (ve VerifyErrors) Error() string {
	errStrings := make([]string, 0, len(ve))
	for _, err := range ve {
		if err != nil {
			errStrings = append(errStrings, err.Error())
		}
	}
	return strings.Join(errStrings, "; ")
}

func VerifyID(m Metrics) error {
	if strings.TrimSpace(m.ID) == "" {
		return ErrIDRequired
	}
	return nil
}

func VerifyType(m Metrics) error {
	if _, ok := allowedTypes[m.MType]; !ok {
		return ErrInvalidType
	}

	return nil
}

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

func VerifyValue(m Metrics) error {
	if m.MType != MTypeGauge {
		return nil
	}

	if m.Value == nil {
		return ErrValueRequired
	}

	return nil
}
