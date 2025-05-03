// Package metrics provides general type of metrics object.
package metrics

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Metrics type constants.
const (
	MTypeGauge   = "gauge"
	MTypeCounter = "counter"
)

// The VerifyOp type is validation check function signature.
type VerifyOp func(m Metrics) error

// A Metrics describes metrics object.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`            // use gauge or counter constants
	Delta *int64   `json:"delta,omitempty"` // for counter
	Value *float64 `json:"value,omitempty"` // for gauge
}

// NewFromStrArgs constructor returns a new metrics.
//
// If the value is not an empty string,
// constructor try to parse string according by MType
// and then define Delta or Value field.
func NewFromStrArgs(id, mType, value string) Metrics {
	var m Metrics

	m.ID = id
	m.MType = mType

	if value == "" {
		return m
	}

	switch m.MType {
	case MTypeGauge:
		v, err := strconv.ParseFloat(value, 64)
		if err == nil {
			m.Value = &v
		}
	case MTypeCounter:
		v, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			m.Delta = &v
		}
	}
	return m
}

// Verify returns metrics object validation result.
// Verify operations order may be important.
func (m Metrics) Verify(ops ...VerifyOp) error {
	var errs VerifyErrors
	for _, op := range ops {
		if err := op(m); err != nil {
			errs = append(errs, op(m))
		}
	}

	if len(errs) != 0 {
		return errs
	}
	return nil
}

// GetValue returns a converted Delta or Value to string.
//
// Zero string returns, if:
//   - metrics type not counter of gauge
//   - delta or value (according by metrics type) is nil
func (m Metrics) GetValue() string {
	switch m.MType {
	case MTypeGauge:
		return valueToStr(m.Value)
	case MTypeCounter:
		return deltaToStr(m.Delta)
	default:
		return ""
	}
}

// MetricsList represents slice of metrics objects.
type MetricsList []Metrics

// Verify behaves as Metrics.Verify method, but errors joins in one.
func (ml MetricsList) Verify(ops ...VerifyOp) error {
	var s []string
	for i, item := range ml {
		if err := item.Verify(ops...); err != nil {
			s = append(s, fmt.Sprintf("%d: %s", i, err.Error()))
		}
	}

	if len(s) != 0 {
		return errors.New("[" + strings.Join(s, ", ") + "]")
	}

	return nil
}

// GetValue build all converted values in one string.
// Values are separated by a newline charecter.
func (ml MetricsList) GetValue() string {
	var b strings.Builder
	for _, m := range ml {
		b.WriteString(m.GetValue())
	}
	return b.String()
}

func valueToStr(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', -1, 64)
}

func deltaToStr(d *int64) string {
	if d == nil {
		return ""
	}
	return strconv.FormatInt(*d, 10)
}
