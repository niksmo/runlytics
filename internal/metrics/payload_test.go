package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsUpdate(t *testing.T) {
	t.Run("Verify method", func(t *testing.T) {
		var delta int64 = 54321
		var value = 123.45000000
		t.Run("No errors", func(t *testing.T) {
			metricsList := []MetricsUpdate{
				{ID: "0", MType: MTypeCounter, Delta: &delta},
				{ID: "0", MType: MTypeGauge, Value: &value},
			}
			for _, m := range metricsList {
				assert.NoError(t, m.Verify())
			}
		})

		t.Run("Empty ID field", func(t *testing.T) {
			metricsList := []MetricsUpdate{
				{ID: " ", MType: MTypeCounter, Delta: &delta},
				{ID: " ", MType: MTypeGauge, Value: &value},
			}
			for _, m := range metricsList {
				err := m.Verify()
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrRequired)
				assert.Equal(t, "'ID': required", err.Error())
			}
		})

		t.Run("Wrong MType field", func(t *testing.T) {
			metricsList := []MetricsUpdate{
				{ID: "0", MType: " ", Delta: &delta},
				{ID: "0", MType: " ", Value: &value},
				{ID: "0", MType: "wcounter", Delta: &delta},
				{ID: "0", MType: "wgauge", Value: &value},
			}
			for _, m := range metricsList {
				err := m.Verify()
				require.Error(t, err)
				assert.Equal(t, "'MType': allowed 'gauge', 'counter'", err.Error())
			}
		})

		t.Run("Empty Value field in Gauge", func(t *testing.T) {
			m := MetricsUpdate{ID: "0", MType: MTypeGauge, Value: nil}
			err := m.Verify()
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrRequired)
			assert.Equal(t, "'Value': required", err.Error())
		})

		t.Run("Empty Delta field in Counter", func(t *testing.T) {
			m := MetricsUpdate{ID: "0", MType: MTypeCounter, Delta: nil}
			err := m.Verify()
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrRequired)
			assert.Equal(t, "'Delta': required", err.Error())
		})

		t.Run("Empty ID, MType, Value", func(t *testing.T) {
			metricsList := []MetricsUpdate{
				{ID: "", MType: "", Delta: nil},
				{ID: " ", MType: " ", Value: nil},
			}
			for _, m := range metricsList {
				err := m.Verify()
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrRequired)
				assert.Equal(t, "'ID': required; 'MType': allowed 'gauge', 'counter'", err.Error())
			}
		})
	})
}

func TestMetricsRead(t *testing.T) {
	t.Run("Verify method", func(t *testing.T) {
		t.Run("No errors", func(t *testing.T) {
			metricsList := []MetricsRead{
				{ID: "0", MType: MTypeCounter},
				{ID: "0", MType: MTypeGauge},
			}
			for _, m := range metricsList {
				assert.NoError(t, m.Verify())
			}
		})

		t.Run("Empty ID field", func(t *testing.T) {
			metricsList := []MetricsRead{
				{ID: "", MType: MTypeCounter},
				{ID: " ", MType: MTypeCounter},
				{ID: "", MType: MTypeGauge},
				{ID: " ", MType: MTypeGauge},
			}
			for _, m := range metricsList {
				err := m.Verify()
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrRequired)
				assert.Equal(t, "'ID': required", err.Error())
			}
		})

		t.Run("Wrong MType field", func(t *testing.T) {
			metricsList := []MetricsRead{
				{ID: "0", MType: ""},
				{ID: "0", MType: " "},
				{ID: "0", MType: "wgauge"},
			}
			for _, m := range metricsList {
				err := m.Verify()
				require.Error(t, err)
				assert.Equal(t, "'MType': allowed 'gauge', 'counter'", err.Error())
			}
		})

		t.Run("Empty ID and MType", func(t *testing.T) {
			metricsList := []MetricsRead{
				{ID: "", MType: ""},
				{ID: " ", MType: " "},
			}
			for _, m := range metricsList {
				err := m.Verify()
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrRequired)
				assert.Equal(t, "'ID': required; 'MType': allowed 'gauge', 'counter'", err.Error())
			}
		})
	})
}
