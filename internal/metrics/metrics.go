package metrics

import (
	"strconv"
	"strings"
)

const (
	MTypeGauge   = "gauge"
	MTypeCounter = "counter"
)

// Mtype is 'gauge' or 'counter'.
// Delta is not nil for 'counter',
// otherwise Value is not nil for 'gauge'
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (m *Metrics) String() string {
	var slice []string
	slice = append(slice, "ID:"+m.ID)
	slice = append(slice, "MType:"+m.MType)

	if m.Delta == nil {
		slice = append(slice, "Delta:nil")
	} else {
		slice = append(slice, "Delta:"+m.StrconvValue())
	}

	if m.Value == nil {
		slice = append(slice, "Value:nil")
	} else {
		slice = append(slice, "Value:"+m.StrconvValue())
	}

	return "Metrics{" + strings.Join(slice, ", ") + "}"
}

// Convert metrics payload value to string presentation.
func (m *Metrics) StrconvValue() string {

	if m.Value != nil && m.Delta != nil {
		return ""
	}

	if m.Value != nil {
		return strconv.FormatFloat(*m.Value, 'f', -1, 64)
	}

	if m.Delta != nil {
		return strconv.FormatInt(*m.Delta, 10)
	}

	return ""
}
