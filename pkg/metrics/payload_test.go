package metrics

import (
	"encoding/json"
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
				assert.Equal(t, "'ID':required", err.Error())
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
				assert.Equal(t, "'MType':allowed 'gauge'|'counter'", err.Error())
			}
		})

		t.Run("Empty Value field in Gauge", func(t *testing.T) {
			m := MetricsUpdate{ID: "0", MType: MTypeGauge, Value: nil}
			err := m.Verify()
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrRequired)
			assert.Equal(t, "'Value':required expect float64", err.Error())
		})

		t.Run("Empty Delta field in Counter", func(t *testing.T) {
			m := MetricsUpdate{ID: "0", MType: MTypeCounter, Delta: nil}
			err := m.Verify()
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrRequired)
			assert.Equal(t, "'Delta':required expect int64", err.Error())
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
				assert.Equal(t, "'ID':required; 'MType':allowed 'gauge'|'counter'", err.Error())
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
				assert.Equal(t, "'ID':required", err.Error())
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
				assert.Equal(t, "'MType':allowed 'gauge'|'counter'", err.Error())
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
				assert.Equal(t, "'ID':required; 'MType':allowed 'gauge'|'counter'", err.Error())
			}
		})
	})
}

func TestMetricsBatchUpdate(t *testing.T) {
	t.Run("JSON encode", func(t *testing.T) {
		expected := `[
		{"id":"0","type":"gauge","value":123.45},
		{"id":"1","type":"counter","delta":54321}
		]`
		gaugeValue := 123.45
		gaugeStruct := MetricsUpdate{ID: "0", MType: MTypeGauge, Value: &gaugeValue}

		counterDelta := int64(54321)
		counterStruct := MetricsUpdate{ID: "1", MType: MTypeCounter, Delta: &counterDelta}

		batchUpdate := MetricsBatchUpdate{gaugeStruct, counterStruct}

		jsonData, err := json.Marshal(batchUpdate)
		require.NoError(t, err)
		assert.JSONEq(t, expected, string(jsonData))
	})

	t.Run("JSON decode", func(t *testing.T) {
		jsonData := []byte(`[
		{"id":"0","type":"gauge","value":123.45},
		{"id":"1","type":"counter","delta":54321}
		]`)

		gaugeValue := 123.45
		gaugeStruct := MetricsUpdate{ID: "0", MType: MTypeGauge, Value: &gaugeValue}

		counterDelta := int64(54321)
		counterStruct := MetricsUpdate{ID: "1", MType: MTypeCounter, Delta: &counterDelta}

		expected := MetricsBatchUpdate{gaugeStruct, counterStruct}

		var actual MetricsBatchUpdate
		err := json.Unmarshal(jsonData, &actual)
		require.NoError(t, err)
		assert.ObjectsAreEqualValues(expected, actual)
	})

	t.Run("Verify method", func(t *testing.T) {
		t.Run("No error", func(t *testing.T) {
			gaugeValue := 123.45
			gaugeStruct := MetricsUpdate{ID: "0", MType: MTypeGauge, Value: &gaugeValue}

			counterDelta := int64(54321)
			counterStruct := MetricsUpdate{ID: "1", MType: MTypeCounter, Delta: &counterDelta}

			batchUpdate := MetricsBatchUpdate{gaugeStruct, counterStruct}

			err := batchUpdate.Verify()
			assert.NoError(t, err)
		})

		t.Run("Should return error", func(t *testing.T) {
			expected := "[0: 'ID':required, 1: 'MType':allowed 'gauge'|'counter'," +
				" 2: 'ID':required; 'MType':allowed 'gauge'|'counter']"
			gaugeValue := 123.45
			gaugeStruct := MetricsUpdate{ID: "", MType: MTypeGauge, Value: &gaugeValue}

			counterDelta := int64(54321)
			counterStruct := MetricsUpdate{ID: "1", MType: "bcounter", Delta: &counterDelta}

			batchUpdate := MetricsBatchUpdate{gaugeStruct, counterStruct, {}}

			err := batchUpdate.Verify()
			require.Error(t, err)
			assert.Equal(t, expected, err.Error())

		})
	})
}
