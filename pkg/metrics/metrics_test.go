package metrics

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsCounter(t *testing.T) {
	type test struct {
		metrics MetricsCounter
		want    string
	}
	t.Run("StrconvValue", func(t *testing.T) {
		tests := []test{
			{
				metrics: MetricsCounter{Delta: -12345},
				want:    "-12345",
			},
			{
				metrics: MetricsCounter{Delta: 0},
				want:    "0",
			},
			{
				metrics: MetricsCounter{Delta: 12345},
				want:    "12345",
			},
		}
		for _, test := range tests {
			assert.Equal(t, test.want, test.metrics.StrconvValue())
		}
	})
}

func TestMetricsGauge(t *testing.T) {
	type test struct {
		metrics MetricsGauge
		want    string
	}
	t.Run("StrconvValue", func(t *testing.T) {
		tests := []test{
			{
				metrics: MetricsGauge{Value: -123.45000000},
				want:    "-123.45",
			},
			{
				metrics: MetricsGauge{Value: 0},
				want:    "0",
			},
			{
				metrics: MetricsGauge{Value: 123.45000000},
				want:    "123.45",
			},
		}
		for _, test := range tests {
			assert.Equal(t, test.want, test.metrics.StrconvValue())
		}
	})
}

func TestVerifyError(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		var ve VerifyErrors
		assert.Equal(t, "", ve.Error())
	})

	t.Run("One simple", func(t *testing.T) {
		var ve VerifyErrors
		ve = append(ve, errors.New("one"))
		assert.Equal(t, "one", ve.Error())
	})

	t.Run("Two simple", func(t *testing.T) {
		var ve VerifyErrors
		ve = append(ve, errors.New("one"), errors.New("two"))
		assert.Equal(t, "one; two", ve.Error())
	})

	t.Run("Two, first simple and second with wrap", func(t *testing.T) {
		var ve VerifyErrors
		ve = append(
			ve,
			errors.New("one"),
			fmt.Errorf("two: %w", errors.New("three")),
		)
		assert.Equal(t, "one; two: three", ve.Error())
	})
}
