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
			INT_ENV    = "INT"
			STRING_ENV = "STRING"
			BOOL_ENV   = "BOOL"
		)

		t.Setenv(INT_ENV, "12345")
		t.Setenv(STRING_ENV, "Hello, world!")
		t.Setenv(BOOL_ENV, "true")

		var envSet env.EnvSet
		intPtr := envSet.Int(INT_ENV)
		strPtr := envSet.String(STRING_ENV)
		boolPtr := envSet.Bool(BOOL_ENV)
		err := envSet.Parse()
		require.NoError(t, err)

		assert.Equal(t, 12345, *intPtr)
		assert.True(t, envSet.IsSet(INT_ENV))

		assert.Equal(t, "Hello, world!", *strPtr)
		assert.True(t, envSet.IsSet(STRING_ENV))

		assert.Equal(t, true, *boolPtr)
		assert.True(t, envSet.IsSet(BOOL_ENV))
	})

	t.Run("Env not provided", func(t *testing.T) {
		const (
			INT_ENV    = "INT"
			STRING_ENV = "STRING"
			BOOL_ENV   = "BOOL"
		)

		t.Setenv(INT_ENV, "12345")
		t.Setenv(BOOL_ENV, "true")

		var envSet env.EnvSet
		intPtr := envSet.Int(INT_ENV)
		strPtr := envSet.String(STRING_ENV) // not provided
		boolPtr := envSet.Bool(BOOL_ENV)
		err := envSet.Parse()
		require.NoError(t, err)

		assert.Zero(t, *strPtr)
		assert.False(t, envSet.IsSet(STRING_ENV))

		assert.Equal(t, 12345, *intPtr)
		assert.True(t, envSet.IsSet(INT_ENV))

		assert.Equal(t, true, *boolPtr)
		assert.True(t, envSet.IsSet(BOOL_ENV))
	})

	t.Run("Parse int error", func(t *testing.T) {
		const INT_ENV = "INT"

		t.Setenv(INT_ENV, "123.45")

		var envSet env.EnvSet
		intPtr := envSet.Int(INT_ENV)
		err := envSet.Parse()
		require.Error(t, err)

		assert.Zero(t, *intPtr)
		assert.False(t, envSet.IsSet(INT_ENV))
	})

	t.Run("Parse bool error", func(t *testing.T) {
		const BOOL_ENV = "BOOL"

		t.Setenv(BOOL_ENV, "TrUe")

		var envSet env.EnvSet
		boolPtr := envSet.Bool(BOOL_ENV)
		err := envSet.Parse()
		require.Error(t, err)
		assert.Zero(t, *boolPtr)
		assert.False(t, envSet.IsSet(BOOL_ENV))
	})

}
