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

type VerifyOp func(m Metrics) error

// Mtype is 'gauge' or 'counter'.
// Delta is not nil for 'counter',
// otherwise Value is not nil for 'gauge'.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

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

type MetricsList []Metrics

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
