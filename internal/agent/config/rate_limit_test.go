package config

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRateLimitFlag(t *testing.T) {

	t.Run("Valid rate limit cmd arg", func(t *testing.T) {
		expected := 4
		actual := getRateLimitFlag(expected)
		assert.Equal(t, expected, actual)
	})

	t.Run("Valid rate limit env arg", func(t *testing.T) {
		cmdValue := 10
		expected := 4
		os.Setenv(rateLimitEnv, strconv.Itoa(expected))
		actual := getRateLimitFlag(cmdValue)
		assert.Equal(t, expected, actual)

	})

	t.Run("Invalid rate limit env arg, cmd is valid", func(t *testing.T) {
		expected := 4
		invalidEnvVelue := "5.5"
		os.Setenv(rateLimitEnv, invalidEnvVelue)
		actual := getRateLimitFlag(expected)
		assert.Equal(t, expected, actual)
	})

	t.Run("Invalid rate limit env and cmd args, should return default", func(t *testing.T) {
		expected := rateLimitDefault
		invalidEnvValue := "5.5"
		invalidCmdValue := 0
		os.Setenv(rateLimitEnv, invalidEnvValue)
		actual := getRateLimitFlag(invalidCmdValue)
		assert.Equal(t, expected, actual)
	})
}
