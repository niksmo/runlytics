package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
	storage := NewMemStorage()
	assert.Empty(t, storage.count)
	assert.Empty(t, storage.gauge)
}

func TestMemStorageAddCount(t *testing.T) {
	storage := NewMemStorage()
	expectedName := "testMetricName1"
	var expectedValue int64 = 30234
	storage.AddCounter(expectedName, expectedValue)
	value, ok := storage.count[expectedName]
	assert.True(t, ok)
	assert.Equal(t, expectedValue, value)
}

func TestMemStorageAddGauge(t *testing.T) {
	storage := NewMemStorage()
	expectedName := "testMetricName2"
	var expectedValue float64 = 0.23984723491234
	storage.AddGauge(expectedName, expectedValue)
	value, ok := storage.gauge[expectedName]
	assert.True(t, ok)
	assert.Equal(t, expectedValue, value)
}
