package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestLogger(t *testing.T) {
	clear := func() {
		Log = zap.NewNop()
	}

	t.Run("Wrong level", func(t *testing.T) {
		assert.Error(t, Initialize("Jinfo"))
		assert.Equal(t, zap.NewNop(), Log)
		t.Cleanup(clear)
	})

	t.Run("First initialization", func(t *testing.T) {
		expectedLvl := "debug"
		err := Initialize(expectedLvl)
		require.Nil(t, err)
		assert.NotEqual(t, zap.NewNop(), Log)
		assert.Equal(t, expectedLvl, Log.Level().String())
		t.Cleanup(clear)
	})

	t.Run("Many initializations", func(t *testing.T) {
		err := Initialize("debug")
		require.Nil(t, err)
		assert.NotEqual(t, zap.NewNop(), Log)
		assert.Equal(t, "debug", Log.Level().String())

		err = Initialize("info")
		require.Nil(t, err)
		assert.NotEqual(t, zap.NewNop(), Log)
		assert.Equal(t, "info", Log.Level().String())

		t.Cleanup(clear)
	})

}
