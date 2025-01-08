package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
	storage := NewMemStorage()
	assert.Empty(t, storage.counter)
	assert.Empty(t, storage.gauge)
}

func TestMemStorageAddCount(t *testing.T) {
	storage := NewMemStorage()
	expectedName := "testMetricName1"
	var expectedValue int64 = 30234
	storage.AddCounter(expectedName, expectedValue)
	value, ok := storage.counter[expectedName]
	assert.True(t, ok)
	assert.Equal(t, expectedValue, value)
}

func TestMemStorageSetGauge(t *testing.T) {
	storage := NewMemStorage()
	expectedName := "testMetricName2"
	var expectedValue = 0.23984723491234
	storage.SetGauge(expectedName, expectedValue)
	value, ok := storage.gauge[expectedName]
	assert.True(t, ok)
	assert.Equal(t, expectedValue, value)
}

func TestMemStorageGetCounter(t *testing.T) {
	type test struct {
		name    string
		storage func() *memStorage
		arg     string
		want    int64
		wantErr error
	}

	tests := []test{
		{
			name: "Should return value",
			storage: func() *memStorage {
				s := NewMemStorage()
				s.counter["testMetric"] = 192347
				return s
			},
			arg:     "testMetric",
			want:    int64(192347),
			wantErr: nil,
		},
		{
			name: "Should return not exists error",
			storage: func() *memStorage {
				return NewMemStorage()
			},
			arg:     "testMetric",
			want:    0,
			wantErr: ErrNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.storage()
			value, err := s.GetCounter(test.arg)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, ErrNotExists)
				return
			}
			assert.Equal(t, test.want, value)
		})

	}
}

func TestMemStorageGetGauge(t *testing.T) {
	type test struct {
		name    string
		storage func() *memStorage
		arg     string
		want    float64
		wantErr error
	}

	tests := []test{
		{
			name: "Should return value",
			storage: func() *memStorage {
				s := NewMemStorage()
				s.gauge["testMetric"] = 192347
				return s
			},
			arg:     "testMetric",
			want:    float64(192347),
			wantErr: nil,
		},
		{
			name: "Should return not exists error",
			storage: func() *memStorage {
				return NewMemStorage()
			},
			arg:     "testMetric",
			want:    0,
			wantErr: ErrNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.storage()
			value, err := s.GetGauge(test.arg)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, ErrNotExists)
				return
			}
			assert.Equal(t, test.want, value)
		})

	}
}

func TestMemStorageGetList(t *testing.T) {
	const (
		gmN = "gauge1"
		cmN = "counter1"
		gmV = 0.1234
		cmV = 5678
	)
	ms := NewMemStorage()
	ms.gauge[gmN] = gmV
	ms.counter[cmN] = cmV

	gauge, counter := ms.GetData()

	assert.Equal(t, ms.gauge, gauge)
	assert.Equal(t, ms.counter, counter)

	gauge[gmN] = 5.6321
	counter[cmN] = 97634

	assert.NotEqual(t, ms.gauge, gauge)
	assert.NotEqual(t, ms.counter, counter)

}
