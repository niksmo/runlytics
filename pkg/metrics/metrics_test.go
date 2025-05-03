package metrics_test

import (
	"testing"

	"github.com/niksmo/runlytics/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsConstructor(t *testing.T) {
	t.Run("invalid type without value", func(t *testing.T) {
		id := "0"
		mType := "invalidType"
		value := ""
		m := metrics.NewFromStrArgs(id, mType, value)
		assert.Equal(t, id, m.ID)
		assert.Equal(t, mType, m.MType)
		assert.Nil(t, m.Value)
		assert.Nil(t, m.Delta)
	})
	t.Run("invalid type with int value", func(t *testing.T) {
		id := "1"
		mType := "invalidType"
		value := "12345"
		m := metrics.NewFromStrArgs(id, mType, value)
		assert.Equal(t, id, m.ID)
		assert.Equal(t, mType, m.MType)
		assert.Nil(t, m.Value)
		assert.Nil(t, m.Delta)
	})
	t.Run("invalid type with float value", func(t *testing.T) {
		id := "2"
		mType := "invalidType"
		value := "123.45"
		m := metrics.NewFromStrArgs(id, mType, value)
		assert.Equal(t, id, m.ID)
		assert.Equal(t, mType, m.MType)
		assert.Nil(t, m.Value)
		assert.Nil(t, m.Delta)
	})

	t.Run("counter type without value", func(t *testing.T) {
		id := "3"
		mType := metrics.MTypeCounter
		value := ""
		m := metrics.NewFromStrArgs(id, mType, value)
		assert.Equal(t, id, m.ID)
		assert.Equal(t, mType, m.MType)
		assert.Nil(t, m.Value)
		assert.Nil(t, m.Delta)
	})
	t.Run("counter type with invalid value", func(t *testing.T) {
		id := "4"
		mType := metrics.MTypeCounter
		value := "invalidValue"
		m := metrics.NewFromStrArgs(id, mType, value)
		assert.Equal(t, id, m.ID)
		assert.Equal(t, mType, m.MType)
		assert.Nil(t, m.Value)
		assert.Nil(t, m.Delta)
	})
	t.Run("counter type with regular value", func(t *testing.T) {
		id := "5"
		mType := metrics.MTypeCounter
		value := "12345"
		expectedValue := int64(12345)
		m := metrics.NewFromStrArgs(id, mType, value)
		assert.Equal(t, id, m.ID)
		assert.Equal(t, mType, m.MType)
		assert.Nil(t, m.Value)
		assert.Equal(t, expectedValue, *m.Delta)
	})

	t.Run("gauge type without value", func(t *testing.T) {
		id := "6"
		mType := metrics.MTypeGauge
		value := ""
		m := metrics.NewFromStrArgs(id, mType, value)
		assert.Equal(t, id, m.ID)
		assert.Equal(t, mType, m.MType)
		assert.Nil(t, m.Value)
		assert.Nil(t, m.Delta)
	})
	t.Run("gauge type with invalid value", func(t *testing.T) {
		id := "7"
		mType := metrics.MTypeGauge
		value := "invalidValue"
		m := metrics.NewFromStrArgs(id, mType, value)
		assert.Equal(t, id, m.ID)
		assert.Equal(t, mType, m.MType)
		assert.Nil(t, m.Value)
		assert.Nil(t, m.Delta)
	})
	t.Run("gauge type with regular value", func(t *testing.T) {
		id := "8"
		mType := metrics.MTypeGauge
		value := "123.45"
		expectedValue := 123.45
		m := metrics.NewFromStrArgs(id, mType, value)
		assert.Equal(t, id, m.ID)
		assert.Equal(t, mType, m.MType)
		assert.Equal(t, expectedValue, *m.Value)
		assert.Nil(t, m.Delta)
	})
}

func TestMetricsGetValue(t *testing.T) {

	t.Run("Invalid type without values", func(t *testing.T) {
		id := "0"
		mType := "invalid"
		m := metrics.Metrics{ID: id, MType: mType}
		assert.Empty(t, m.GetValue())
	})

	t.Run("Invalid type with Delta", func(t *testing.T) {
		id := "1"
		mType := "invalid"
		delta := int64(12345)
		m := metrics.Metrics{ID: id, MType: mType, Delta: &delta}
		assert.Empty(t, m.GetValue())
	})

	t.Run("Invalid type with Value", func(t *testing.T) {
		id := "2"
		mType := "invalid"
		value := 123.45
		m := metrics.Metrics{ID: id, MType: mType, Value: &value}
		assert.Empty(t, m.GetValue())
	})

	t.Run("Counter without values", func(t *testing.T) {
		id := "0"
		mType := metrics.MTypeCounter
		m := metrics.Metrics{ID: id, MType: mType}
		assert.Empty(t, m.GetValue())
	})

	t.Run("Counter with Value", func(t *testing.T) {
		id := "1"
		mType := metrics.MTypeCounter
		value := 123.45
		m := metrics.Metrics{ID: id, MType: mType, Value: &value}
		assert.Empty(t, m.GetValue())
	})

	t.Run("Counter with Delta", func(t *testing.T) {
		id := "2"
		mType := metrics.MTypeCounter
		delta := int64(12345)
		expectedValueStr := "12345"
		m := metrics.Metrics{ID: id, MType: mType, Delta: &delta}
		assert.Equal(t, expectedValueStr, m.GetValue())
	})

	t.Run("Counter with Value and Delta", func(t *testing.T) {
		id := "3"
		mType := metrics.MTypeCounter
		delta := int64(12345)
		value := 543.21
		expectedValueStr := "12345"
		m := metrics.Metrics{ID: id, MType: mType, Delta: &delta, Value: &value}
		assert.Equal(t, expectedValueStr, m.GetValue())
	})

	t.Run("Gauge without values", func(t *testing.T) {
		id := "0"
		mType := metrics.MTypeGauge
		m := metrics.Metrics{ID: id, MType: mType}
		assert.Empty(t, m.GetValue())
	})

	t.Run("Gauge with Value", func(t *testing.T) {
		id := "1"
		mType := metrics.MTypeGauge
		value := 123.45
		expectedValue := "123.45"
		m := metrics.Metrics{ID: id, MType: mType, Value: &value}
		assert.Equal(t, expectedValue, m.GetValue())
	})

	t.Run("Gauge with Delta", func(t *testing.T) {
		id := "2"
		mType := metrics.MTypeGauge
		delta := int64(12345)
		m := metrics.Metrics{ID: id, MType: mType, Delta: &delta}
		assert.Empty(t, m.GetValue())
	})

	t.Run("Gauge with Value and Delta", func(t *testing.T) {
		id := "3"
		mType := metrics.MTypeGauge
		delta := int64(12345)
		value := 543.21
		expectedValueStr := "543.21"
		m := metrics.Metrics{ID: id, MType: mType, Delta: &delta, Value: &value}
		assert.Equal(t, expectedValueStr, m.GetValue())
	})
}

func TestMetricsVerify(t *testing.T) {
	t.Run("verify ID", func(t *testing.T) {
		t.Run("empty ID", func(t *testing.T) {
			m := metrics.Metrics{}
			err := m.Verify(metrics.VerifyID)
			require.Error(t, err)
			assert.ErrorIs(t, err, metrics.ErrIDRequired)
		})

		t.Run("regular ID", func(t *testing.T) {
			m := metrics.Metrics{ID: "1"}
			err := m.Verify(metrics.VerifyID)
			assert.NoError(t, err)
		})
	})

	t.Run("verify Type", func(t *testing.T) {
		t.Run("Invalid", func(t *testing.T) {
			m := metrics.Metrics{MType: "invalid"}
			err := m.Verify(metrics.VerifyType)
			require.Error(t, err)
			assert.ErrorIs(t, err, metrics.ErrInvalidType)
		})

		t.Run("Counter", func(t *testing.T) {
			m := metrics.Metrics{MType: metrics.MTypeCounter}
			err := m.Verify(metrics.VerifyType)
			assert.NoError(t, err)
		})

		t.Run("Gauge", func(t *testing.T) {
			m := metrics.Metrics{MType: metrics.MTypeGauge}
			err := m.Verify(metrics.VerifyType)
			assert.NoError(t, err)
		})
	})

	t.Run("verify Delta", func(t *testing.T) {
		t.Run("nil Delta", func(t *testing.T) {
			m := metrics.Metrics{MType: metrics.MTypeCounter}
			err := m.Verify(metrics.VerifyDelta)
			require.Error(t, err)
			assert.ErrorIs(t, err, metrics.ErrDeltaRequired)
		})
		t.Run("negative Delta", func(t *testing.T) {
			delta := int64(-1)
			m := metrics.Metrics{MType: metrics.MTypeCounter, Delta: &delta}
			err := m.Verify(metrics.VerifyDelta)
			require.Error(t, err)
			assert.ErrorIs(t, err, metrics.ErrDeltaLessZero)
		})
		t.Run("regular Delta", func(t *testing.T) {
			delta := int64(0)
			m := metrics.Metrics{MType: metrics.MTypeCounter, Delta: &delta}
			err := m.Verify(metrics.VerifyDelta)
			require.NoError(t, err)
		})
	})

	t.Run("verify Value", func(t *testing.T) {
		t.Run("nil Value", func(t *testing.T) {
			m := metrics.Metrics{MType: metrics.MTypeGauge}
			err := m.Verify(metrics.VerifyValue)
			require.Error(t, err)
			assert.ErrorIs(t, err, metrics.ErrValueRequired)
		})
		t.Run("negative Value", func(t *testing.T) {
			value := -123.45
			m := metrics.Metrics{MType: metrics.MTypeGauge, Value: &value}
			err := m.Verify(metrics.VerifyValue)
			assert.NoError(t, err)
		})

		t.Run("zero Value", func(t *testing.T) {
			value := 0.00
			m := metrics.Metrics{MType: metrics.MTypeGauge, Value: &value}
			err := m.Verify(metrics.VerifyValue)
			assert.NoError(t, err)
		})

		t.Run("positive Value", func(t *testing.T) {
			value := 123.45
			m := metrics.Metrics{MType: metrics.MTypeGauge, Value: &value}
			err := m.Verify(metrics.VerifyValue)
			assert.NoError(t, err)
		})
	})

	t.Run("composite verify", func(t *testing.T) {
		t.Run("Invalid type", func(t *testing.T) {
			m := metrics.Metrics{MType: "Invalid"}
			err := m.Verify(
				metrics.VerifyID,
				metrics.VerifyType,
				metrics.VerifyDelta,
				metrics.VerifyValue,
			)
			require.Error(t, err)
			assert.ErrorIs(t, err, metrics.ErrIDRequired)
			assert.ErrorIs(t, err, metrics.ErrInvalidType)
			assert.NotErrorIs(
				t, err, metrics.ErrDeltaRequired, "Delta should not check",
			)
			assert.NotErrorIs(
				t, err, metrics.ErrValueRequired, "Value should not check",
			)
		})

		t.Run("Counter all possible errors", func(t *testing.T) {
			m := metrics.Metrics{MType: metrics.MTypeCounter}
			err := m.Verify(
				metrics.VerifyID,
				metrics.VerifyType,
				metrics.VerifyDelta,
				metrics.VerifyValue,
			)
			require.Error(t, err)
			assert.ErrorIs(t, err, metrics.ErrIDRequired)
			assert.ErrorIs(t, err, metrics.ErrDeltaRequired)
			assert.NotErrorIs(t, err, metrics.ErrInvalidType)
		})

		t.Run("Counter regular, no errors", func(t *testing.T) {
			id := "1"
			mType := metrics.MTypeCounter
			delta := int64(12345)
			m := metrics.Metrics{ID: id, MType: mType, Delta: &delta}
			err := m.Verify(
				metrics.VerifyID,
				metrics.VerifyType,
				metrics.VerifyDelta,
				metrics.VerifyValue,
			)
			assert.NoError(t, err)
		})

		t.Run("Gauge all possible errors", func(t *testing.T) {
			m := metrics.Metrics{MType: metrics.MTypeGauge}
			err := m.Verify(
				metrics.VerifyID,
				metrics.VerifyType,
				metrics.VerifyDelta,
				metrics.VerifyValue,
			)
			require.Error(t, err)
			assert.ErrorIs(t, err, metrics.ErrIDRequired)
			assert.ErrorIs(t, err, metrics.ErrValueRequired)
			assert.NotErrorIs(t, err, metrics.ErrInvalidType)
		})

		t.Run("Gauge regular, no errors", func(t *testing.T) {
			id := "1"
			mType := metrics.MTypeGauge
			value := 123.45
			m := metrics.Metrics{ID: id, MType: mType, Value: &value}
			err := m.Verify(
				metrics.VerifyID,
				metrics.VerifyType,
				metrics.VerifyDelta,
				metrics.VerifyValue,
			)
			assert.NoError(t, err)
		})
	})
}

func TestMetricsListVerify(t *testing.T) {
	t.Run("Has errors", func(t *testing.T) {
		m0Delta := int64(12345)
		m1Value := 123.45
		var m2Value *float64
		m0 := metrics.Metrics{
			ID: "0", MType: metrics.MTypeCounter, Delta: &m0Delta,
		}
		m1 := metrics.Metrics{
			ID: "", MType: metrics.MTypeGauge, Value: &m1Value,
		}
		m2 := metrics.Metrics{
			ID: "2", MType: metrics.MTypeGauge, Value: m2Value,
		}
		ml := metrics.MetricsList{m0, m1, m2}
		err := ml.Verify(
			metrics.VerifyID,
			metrics.VerifyType,
			metrics.VerifyDelta,
			metrics.VerifyValue,
		)
		assert.Error(t, err)
	})

	t.Run("Regular", func(t *testing.T) {
		m0Delta := int64(12345)
		m1Value := 123.45
		m2Value := 567.89
		m0 := metrics.Metrics{
			ID: "0", MType: metrics.MTypeCounter, Delta: &m0Delta,
		}
		m1 := metrics.Metrics{
			ID: "1", MType: metrics.MTypeGauge, Value: &m1Value,
		}
		m2 := metrics.Metrics{
			ID: "2", MType: metrics.MTypeGauge, Value: &m2Value,
		}
		ml := metrics.MetricsList{m0, m1, m2}
		err := ml.Verify(
			metrics.VerifyID,
			metrics.VerifyType,
			metrics.VerifyDelta,
			metrics.VerifyValue,
		)
		assert.NoError(t, err)
	})
}
