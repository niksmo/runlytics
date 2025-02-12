package metrics

import (
	"errors"
	"fmt"
	"strings"
)

// Mtype is 'gauge' or 'counter'.
// Delta is not nil for 'counter',
// otherwise Value is not nil for 'gauge'.
type MetricsUpdate struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (mu *MetricsUpdate) Verify() error {
	var errSlice VerifyErrors
	var err error

	if err = verifyFieldID(mu.ID); err != nil {
		errSlice = append(errSlice, err)
	}

	if err = verifyFiledMType(mu.MType); err != nil {
		errSlice = append(errSlice, err)
	}

	if mu.MType == MTypeCounter {
		if err = verifyFieldDelta(mu.Delta); err != nil {
			errSlice = append(errSlice, err)
		}
	}

	if mu.MType == MTypeGauge {
		if err = verifyFieldValue(mu.Value); err != nil {
			errSlice = append(errSlice, err)
		}
	}

	if len(errSlice) != 0 {
		return errSlice
	}

	return nil
}

type MetricsBatchUpdate []MetricsUpdate

func (mbu *MetricsBatchUpdate) Verify() error {
	var stringSlice []string
	for i, item := range *mbu {
		if err := item.Verify(); err != nil {
			stringSlice = append(
				stringSlice, fmt.Sprintf("%d: %s", i, err.Error()),
			)
		}
	}
	if len(stringSlice) != 0 {
		return errors.New("[" + strings.Join(stringSlice, ", ") + "]")
	}
	return nil
}

type MetricsRead struct {
	ID    string `json:"id"`
	MType string `json:"type"`
}

func (mr *MetricsRead) Verify() error {
	var errSlice VerifyErrors
	var err error

	if err = verifyFieldID(mr.ID); err != nil {
		errSlice = append(errSlice, err)
	}

	if err = verifyFiledMType(mr.MType); err != nil {
		errSlice = append(errSlice, err)
	}

	if len(errSlice) != 0 {
		return errSlice
	}

	return nil
}
