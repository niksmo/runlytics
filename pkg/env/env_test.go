package env_test

import (
	"testing"

	"github.com/niksmo/runlytics/pkg/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvSet(t *testing.T) {
	t.Run("regular use no errors", func(t *testing.T) {
		const (
			intEnv    = "INT"
			stringEnv = "STRING"
			boolEnv   = "BOOL"
		)

		t.Setenv(intEnv, "12345")
		t.Setenv(stringEnv, "Hello, world!")
		t.Setenv(boolEnv, "true")

		envSet := env.New()
		intPtr := envSet.Int(intEnv)
		strPtr := envSet.String(stringEnv)
		boolPtr := envSet.Bool(boolEnv)
		err := envSet.Parse()
		require.NoError(t, err)

		assert.Equal(t, 12345, *intPtr)
		assert.True(t, envSet.IsSet(intEnv))

		assert.Equal(t, "Hello, world!", *strPtr)
		assert.True(t, envSet.IsSet(stringEnv))

		assert.Equal(t, true, *boolPtr)
		assert.True(t, envSet.IsSet(boolEnv))
	})

	t.Run("Env not provided", func(t *testing.T) {
		const (
			intEnv    = "INT"
			stringEnv = "STRING"
			boolEnv   = "BOOL"
		)

		t.Setenv(intEnv, "12345")
		t.Setenv(boolEnv, "true")

		envSet := env.New()
		intPtr := envSet.Int(intEnv)
		strPtr := envSet.String(stringEnv) // not provided
		boolPtr := envSet.Bool(boolEnv)
		err := envSet.Parse()
		require.NoError(t, err)

		assert.Zero(t, *strPtr)
		assert.False(t, envSet.IsSet(stringEnv))

		assert.Equal(t, 12345, *intPtr)
		assert.True(t, envSet.IsSet(intEnv))

		assert.Equal(t, true, *boolPtr)
		assert.True(t, envSet.IsSet(boolEnv))
	})

	t.Run("Parse int error", func(t *testing.T) {
		const intEnv = "INT"

		t.Setenv(intEnv, "123.45")

		envSet := env.New()
		intPtr := envSet.Int(intEnv)
		err := envSet.Parse()
		require.Error(t, err)

		assert.Zero(t, *intPtr)
		assert.False(t, envSet.IsSet(intEnv))
	})

	t.Run("Parse bool error", func(t *testing.T) {
		const boolEnv = "BOOL"

		t.Setenv(boolEnv, "TrUe")

		envSet := env.New()
		boolPtr := envSet.Bool(boolEnv)
		err := envSet.Parse()
		require.Error(t, err)
		assert.Zero(t, *boolPtr)
		assert.False(t, envSet.IsSet(boolEnv))
	})

}
