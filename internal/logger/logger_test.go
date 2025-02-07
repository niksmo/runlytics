package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestLogger(t *testing.T) {
	clear := func() {
		Log = zap.NewNop()
	}

	t.Run("Wrong level", func(t *testing.T) {
		Init("Jinfo")
		assert.Equal(t, zap.NewNop(), Log)
		t.Cleanup(clear)
	})

	t.Run("One initialization", func(t *testing.T) {
		expectedLvl := "debug"
		Init(expectedLvl)
		assert.NotEqual(t, zap.NewNop(), Log)
		assert.Equal(t, expectedLvl, Log.Level().String())
		t.Cleanup(clear)
	})

	t.Run("Many initializations", func(t *testing.T) {
		Init("debug")
		assert.NotEqual(t, zap.NewNop(), Log)
		assert.Equal(t, "debug", Log.Level().String())

		Init("info")
		assert.NotEqual(t, zap.NewNop(), Log)
		assert.Equal(t, "info", Log.Level().String())
		t.Cleanup(clear)
	})

}
