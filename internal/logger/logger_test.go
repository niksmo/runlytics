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
		assert.Error(t, Init("Jinfo"))
		assert.Equal(t, zap.NewNop(), Log)
		t.Cleanup(clear)
	})

	t.Run("One initialization", func(t *testing.T) {
		expectedLvl := "debug"
		err := Init(expectedLvl)
		require.Nil(t, err)
		assert.NotEqual(t, zap.NewNop(), Log)
		assert.Equal(t, expectedLvl, Log.Level().String())
		t.Cleanup(clear)
	})

	t.Run("Many initializations", func(t *testing.T) {
		err := Init("debug")
		require.Nil(t, err)
		assert.NotEqual(t, zap.NewNop(), Log)
		assert.Equal(t, "debug", Log.Level().String())

		err = Init("info")
		require.Nil(t, err)
		assert.NotEqual(t, zap.NewNop(), Log)
		assert.Equal(t, "info", Log.Level().String())

		t.Cleanup(clear)
	})

}
