package metrics

import (
	"strconv"
)

const (
	MTypeGauge   = "gauge"
	MTypeCounter = "counter"
)

// Mtype is 'gauge' or 'counter'.
// Delta is not nil for 'counter',
// otherwise Value is not nil for 'gauge'.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (m Metrics) GetValue() string {
	switch m.MType {
	case MTypeGauge:
		return strconvValue(*m.Value)
	case MTypeCounter:
		return strconvDelta(*m.Delta)
	default:
		return "value conversion error"
	}
}

func strconvValue(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func strconvDelta(d int64) string {
	return strconv.FormatInt(d, 10)
}
