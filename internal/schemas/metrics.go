package schemas

import (
	"fmt"
	"strings"
)

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
		slice = append(slice, fmt.Sprintf("Delta:%v", *m.Delta))
	}

	if m.Value == nil {
		slice = append(slice, "Value:nil")
	} else {
		slice = append(slice, fmt.Sprintf("Value:%v", *m.Value))
	}

	return "Metrics{" + strings.Join(slice, ", ") + "}"
}
