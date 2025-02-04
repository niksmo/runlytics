package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsMethods(t *testing.T) {
	t.Run("String method", func(t *testing.T) {
		t.Run("Correct metrics", func(t *testing.T) {
			var delta int64 = 12345
			var value = 123.45000000
			tests := []struct {
				metrics Metrics
				want    string
			}{
				{
					metrics: Metrics{ID: "0", MType: MTypeCounter, Delta: nil, Value: nil},
					want:    "Metrics{ID:0, MType:counter, Delta:nil, Value:nil}",
				},
				{
					metrics: Metrics{ID: "1", MType: MTypeGauge, Delta: nil, Value: nil},
					want:    "Metrics{ID:1, MType:gauge, Delta:nil, Value:nil}",
				},
				{
					metrics: Metrics{ID: "2", MType: MTypeCounter, Delta: &delta, Value: nil},
					want:    "Metrics{ID:2, MType:counter, Delta:12345, Value:nil}",
				},
				{
					metrics: Metrics{ID: "3", MType: MTypeGauge, Delta: nil, Value: &value},
					want:    "Metrics{ID:3, MType:gauge, Delta:nil, Value:123.45}",
				},
			}

			for _, test := range tests {
				assert.Equal(t, test.want, test.metrics.String())
			}
		})

		t.Run("Wrong metrics", func(t *testing.T) {
			var delta int64 = 12345
			var value = 123.45000000
			tests := []struct {
				metrics Metrics
				want    string
			}{
				{
					metrics: Metrics{ID: "w0", MType: MTypeCounter, Delta: &delta, Value: &value},
					want:    "Metrics{ID:w0, MType:counter, Delta:, Value:}",
				},
				{
					metrics: Metrics{ID: "w1", MType: MTypeGauge, Delta: &delta, Value: &value},
					want:    "Metrics{ID:w1, MType:gauge, Delta:, Value:}",
				},
			}

			for _, test := range tests {
				assert.Equal(t, test.want, test.metrics.String())
			}
		})
	})

	t.Run("StrconvValue method", func(t *testing.T) {
		t.Run("Value and Delta is nil", func(t *testing.T) {
			mCounter := Metrics{ID: "c0", MType: MTypeCounter}
			mGauge := Metrics{ID: "g1", MType: MTypeGauge}

			for _, m := range []Metrics{mCounter, mGauge} {
				assert.Zero(t, m.StrconvValue())
			}
		})
		t.Run("Gauge and Value", func(t *testing.T) {
			value := 123.4500000000
			mGauge := Metrics{ID: "gauge", MType: MTypeGauge, Value: &value}
			assert.Equal(t, "123.45", mGauge.StrconvValue())
		})

		t.Run("Counter and Delta", func(t *testing.T) {
			var delta int64 = 12345
			mCounter := Metrics{ID: "counter", MType: MTypeCounter, Delta: &delta}
			assert.Equal(t, "12345", mCounter.StrconvValue())
		})

		t.Run("Wrong type with Value or Delta", func(t *testing.T) {
			var delta int64 = 12345
			var value = 123.4500000000
			wrongCounter := Metrics{ID: "wc0", MType: "BCounter", Delta: &delta}
			wrongGauge := Metrics{ID: "wg1", MType: "BGauge", Value: &value}

			assert.Equal(t, "12345", wrongCounter.StrconvValue())
			assert.Equal(t, "123.45", wrongGauge.StrconvValue())
		})
	})
}
