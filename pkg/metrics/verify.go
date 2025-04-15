package metrics

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrRequired = errors.New("required")
)

type VerifyErrors []error

func (ev VerifyErrors) Unwrap() []error {
	if len(ev) == 0 {
		return nil
	}
	return ev
}

func (ev VerifyErrors) Error() string {
	errStrings := make([]string, 0, len(ev))
	for _, err := range ev {
		errStrings = append(errStrings, err.Error())
	}
	return strings.Join(errStrings, "; ")
}

func verifyFieldID(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("'ID':%w", ErrRequired)
	}
	return nil
}

func verifyFiledMType(mType string) error {
	allowed := map[string]struct{}{MTypeCounter: {}, MTypeGauge: {}}
	if _, ok := allowed[mType]; !ok {
		return fmt.Errorf("'MType':allowed 'gauge'|'counter'")
	}

	return nil
}

func verifyFieldDelta(value *int64) error {
	field := "'Delta'"
	if value == nil {
		return fmt.Errorf("%s:%w expect int64", field, ErrRequired)
	}

	if *value < 0 {
		return fmt.Errorf("%s:less then 0", field)
	}
	return nil
}

func verifyFieldValue(value *float64) error {
	if value == nil {
		return fmt.Errorf("'Value':%w expect float64", ErrRequired)
	}
	return nil
}
