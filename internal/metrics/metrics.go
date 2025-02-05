package metrics

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	MTypeGauge   = "gauge"
	MTypeCounter = "counter"
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

type ValueStringer interface {
	StrconvValue() string
}

type Metrics interface {
	ValueStringer
}

type MetricsCounter struct {
	ID    string `json:"id"`
	MType string `json:"type"`
	Delta int64  `json:"delta"`
}

func (mc MetricsCounter) StrconvValue() string {
	return strconvDelta(mc.Delta)
}

type MetricsGauge struct {
	ID    string  `json:"id"`
	MType string  `json:"type"`
	Value float64 `json:"value"`
}

func (mg MetricsGauge) StrconvValue() string {
	return strconvValue(mg.Value)
}

func strconvValue(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func strconvDelta(d int64) string {
	return strconv.FormatInt(d, 10)
}

func verifyFieldID(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("'ID': %w", ErrRequired)
	}
	return nil
}

func verifyFiledMType(mType string) error {
	field := "'MType'"
	if strings.TrimSpace(mType) == "" {
		return fmt.Errorf("%s: %w", field, ErrRequired)
	}

	allowed := map[string]struct{}{MTypeCounter: {}, MTypeGauge: {}}
	if _, ok := allowed[mType]; !ok {
		return fmt.Errorf("%s: allowed 'gauge', 'counter'", field)
	}

	return nil
}

func verifyFieldDelta(value *int64) error {
	field := "'Delta'"
	if value == nil {
		return fmt.Errorf("%s: %w", field, ErrRequired)
	}

	if *value < 0 {
		return fmt.Errorf("%s: less then 0", field)
	}
	return nil
}

func verifyFieldValue(value *float64) error {
	if value == nil {
		return fmt.Errorf("'Value': %w", ErrRequired)
	}
	return nil
}
